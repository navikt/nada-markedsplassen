import { useControlPanel } from "./controlPanel";
import { useActivateWorkstationURLListForIdent, useCheckOnpremConnectivity, useCreateWorkstationConnectivityWorkflow, useStartWorkstation, useStopWorkstation, useUpdateWorkstationURLListItemForIdent, useWorkstationConnectivityWorkflow, useWorkstationEffectiveTags, useWorkstationMine, useWorkstationOnpremMapping, useWorkstationOptions, useWorkstationURLListForIdent } from "./queries";
import { Alert, Loader } from "@navikt/ds-react";
import React, { use, useEffect } from "react";
import { InfoForm } from "./infoForm";
import { SettingsForm } from "./SettingsForm";
import { DatasourcesForm } from "./DatasourcesForm";
import { InternetOpeningsForm } from "./internetOpeningsForm";
import { EffectiveTags, JobStateRunning, WorkstationConnectJob, WorkstationDisconnectJob, WorkstationOnpremAllowList, WorkstationOptions, WorkstationURLListForIdent } from "../../lib/rest/generatedDto";
import { UseQueryResult } from "@tanstack/react-query";
import { HttpError } from "../../lib/rest/request";
import { se } from "date-fns/locale";
import { set } from "lodash";
import Quiz, { getQuizReadCookie, setQuizReadCookie } from "./Quiz";

const injectExtraInfoToKnast = (knast: any, knastOptions?: WorkstationOptions, workstationOnpremMapping?: WorkstationOnpremAllowList
    , effectiveTags?: EffectiveTags, urlList?: WorkstationURLListForIdent, onpremState?: string, operationalStatus?: string, internetState?: string) => {
    const image = knastOptions?.containerImages?.find((img) => img?.image === knast.image);
    const machineType = knastOptions?.machineTypes?.find((type) => type?.machineType === knast.config.machineType);
    const onpremConfigured = workstationOnpremMapping && workstationOnpremMapping.hosts && workstationOnpremMapping.hosts.length > 0;

    return {
        ...knast, imageTitle: image?.labels["org.opencontainers.image.title"] || "Ukjent miljø",
        machineTypeInfo: machineType, workstationOnpremMapping, effectiveTags, internetUrls: urlList
        , onpremConfigured, onpremState, operationalStatus, internetState
    }
}

const useFrontendUpdatingOnprem = () => {
    const [updatingOnprem, setUpdatingOnprem] = React.useState<boolean>(false);
    const [revertTimer, setRevertTimer] = React.useState<NodeJS.Timeout | null>(null);
    return {
        frontendUpdatingOnprem: updatingOnprem,
        setFrontendUpdatingOnprem: (value: boolean) => {
            if (revertTimer) {
                clearTimeout(revertTimer);
            }
            setUpdatingOnprem(value);
            setRevertTimer(setTimeout(() => {
                setUpdatingOnprem(updatingOnprem);
            }, 3000));
        }
    }
}

const Knast = () => {
    const [activeForm, setActiveForm] = React.useState<"info" | "settings" | "onprem" | "internet">("info")
    const createConnectivityWorkflow = useCreateWorkstationConnectivityWorkflow();
    const activateUrls = useActivateWorkstationURLListForIdent();
    const updateUrlAllowList = useUpdateWorkstationURLListItemForIdent();
    const connectivityJobs = useWorkstationConnectivityWorkflow()
    const urlList = useWorkstationURLListForIdent()
    const [internetState, setInternetState] = React.useState<"activated" | "deactivated" | "updating" | undefined>(
        urlList.data?.items.some(it => it?.expiresAt && new Date(it.expiresAt) > new Date()) ? "activated" : "deactivated"
    );
    const [showQuiz, setShowQuiz] = React.useState(!getQuizReadCookie());
    const [onpremError, setOnpremError] = React.useState<string | null>(null);
    const { frontendUpdatingOnprem, setFrontendUpdatingOnprem } = useFrontendUpdatingOnprem();
    const startKnast = useStartWorkstation()
    const stopKnast = useStopWorkstation()

    const knast = useWorkstationMine()
    const knastOptions = useWorkstationOptions()
    const { operationalStatus, ControlPanel } = useControlPanel(knast.data);
    const workstationOnpremMapping = useWorkstationOnpremMapping()
    const effectiveTags = useWorkstationEffectiveTags()
    const updatingOnprem = frontendUpdatingOnprem || connectivityJobs.data?.disconnect?.state === JobStateRunning || connectivityJobs.data?.connect?.some((job): job is WorkstationConnectJob =>
        job !== undefined && job.state === JobStateRunning);
    const onpremState = updatingOnprem ? "updating" : !!effectiveTags.data?.tags?.length ? "activated" : "deactivated";

    const onActivateOnprem = async (enable: boolean) => {
        try {
            setFrontendUpdatingOnprem(true);
            await createConnectivityWorkflow.mutateAsync(enable ? workstationOnpremMapping.data!! : { hosts: [] })
        } catch (e) {
            console.error("Error in onActivateOnprem:", e);
            setOnpremError("Ukjent feil");
        }
    }

    const onActivateInternet = async (enable: boolean) => {
        setInternetState("updating");
        if (!enable) {
            for (const urlItem of urlList.data?.items.filter(it => !!it && it.id && it.url && it.createdAt && it.duration) || []) {
                await updateUrlAllowList.mutateAsync({
                    ...urlItem!!,
                    id: urlItem!!.id!,
                    url: urlItem!!.url!,
                    createdAt: urlItem!!.createdAt!,
                    duration: urlItem!!.duration!,
                    expiresAt: new Date().toISOString(),
                    selected: urlItem!!.selected
                });
            }
        } else {
            const urlsToActivate = urlList.data!!.items.filter(it => it?.selected).map(it => it?.id!!);
            await activateUrls.mutateAsync(urlsToActivate);
        }
        setInternetState(enable ? "activated" : "deactivated");
    }

    if (knast.isLoading) {
        return <div>Lasting min knast <Loader /></div>
    }

    if (knast.isError) {
        return <Alert variant="error">Feil ved lasting av knast: {knast.error instanceof Error ? knast.error.message : "ukjent feil"}</Alert>
    }

    if (!knast.data) {
        return <div>Ingen knast funnet for bruker</div>
    }

    const knastData = injectExtraInfoToKnast(knast.data, knastOptions.data, workstationOnpremMapping?.data, effectiveTags?.data, urlList.data, onpremState, operationalStatus, internetState);

    return <div className="flex flex-col gap-4">
        {onpremError && <Alert variant="error" onClose={() => setOnpremError(null)}>{onpremError}</Alert>}
        <Quiz onClose={() => {
            setShowQuiz(false)
            setQuizReadCookie();
        }} show={showQuiz} />
        <ControlPanel
            knastInfo={knastData}
            onStartKnast={() => startKnast.mutate()}
            onStopKnast={() => stopKnast.mutate()}
            onSettings={() => setActiveForm("settings")}
            onActivateOnprem={() => onActivateOnprem(true)
            } onActivateInternet={() => onActivateInternet(true)} onDeactivateOnPrem={() => onActivateOnprem(false)}
            onDeactivateInternet={() => onActivateInternet(false)} onConfigureOnprem={() => setActiveForm("onprem")}
            onConfigureInternet={() => setActiveForm("internet")} />
        {activeForm === "info" && <InfoForm knastInfo={knastData} operationalStatus={operationalStatus}
            onActivateOnprem={() => onActivateOnprem(true)}
            onActivateInternet={() => onActivateInternet(true)}
            onDeactivateOnPrem={() => onActivateOnprem(false)}
            onDeactivateInternet={() => onActivateInternet(false)}
            onConfigureOnprem={() => {
                setActiveForm("onprem");
            }}
            onConfigureInternet={() => {
                setActiveForm("internet");
            }} />}
        {activeForm === "settings" && <SettingsForm onConfigureDatasources={() => setActiveForm("onprem")} onConfigureInternet={() => setActiveForm("internet")} knastInfo={knastData} options={knastOptions.data} onSave={() => { }} onCancel={() => setActiveForm("info")} />}
        {activeForm === "onprem" && <DatasourcesForm knastInfo={knastData} onCancel={() => setActiveForm("info")} />}
        {activeForm === "internet" && <InternetOpeningsForm onSave={() => {
        }} onCancel={() => setActiveForm("info")} />}
    </div>
}

export default Knast
