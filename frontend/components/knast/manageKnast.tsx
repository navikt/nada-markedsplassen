import { useControlPanel } from "./controlPanel";
import { useActivateWorkstationURLListForIdent, useCreateWorkstationConnectivityWorkflow, useStartWorkstation, useStopWorkstation, useUpdateWorkstationURLListItemForIdent, useWorkstationConnectivityWorkflow, useWorkstationEffectiveTags, useWorkstationExists, useWorkstationJobs, useWorkstationMine, useWorkstationOnpremMapping, useWorkstationOptions, useWorkstationResyncJobs, useWorkstationURLListForIdent } from "./queries";
import { Alert, Loader, Tabs } from "@navikt/ds-react";
import React from "react";
import { InfoForm } from "./infoForm";
import { SettingsForm } from "./SettingsForm";
import { DatasourcesForm } from "./DatasourcesForm";
import { InternetOpeningsForm } from "./internetOpeningsForm";
import { EffectiveTags, JobStateRunning, WorkstationConnectJob, WorkstationOnpremAllowList, WorkstationOptions, WorkstationURLListForIdent } from "../../lib/rest/generatedDto";
import Quiz, { getQuizReadCookie, setQuizReadCookie } from "./Quiz";
import { useOnpremMapping } from "../onpremmapping/queries";
import { IconCircle, IconGear, IconInternetOpening, IconNavData } from "./widgets/knastIcons";
import { LaptopIcon } from "@navikt/aksel-icons";
import { LogViewer } from "./widgets/logViewer";
import { ColorInfo } from "./designTokens";
import { ConfigureKnastForm } from "./configureKnast";


export const ManageKnastPage = () => {
    const [activeForm, setActiveForm] = React.useState<"overview" |"environment" | "onprem" | "internet" | "log">("overview")
    const createConnectivityWorkflow = useCreateWorkstationConnectivityWorkflow();
    const activateUrls = useActivateWorkstationURLListForIdent();
    const updateUrlAllowList = useUpdateWorkstationURLListItemForIdent();
    const connectivityJobs = useWorkstationConnectivityWorkflow()
    const urlList = useWorkstationURLListForIdent()
    const activeItemInUrlList = urlList.data?.items.some(it => it?.expiresAt && new Date(it.expiresAt) > new Date());
    const [lastUpdateInternet, setLastUpdateInternet] = React.useState<Date | undefined>(undefined);
    const [showQuiz, setShowQuiz] = React.useState(!getQuizReadCookie());
    const [onpremError, setOnpremError] = React.useState<string | null>(null);
    const onpremMapping = useOnpremMapping()
    const [frontendUpdatingOnprem, setFrontendUpdatingOnprem] = React.useState<Date | undefined>(undefined)
    const startKnast = useStartWorkstation()
    const stopKnast = useStopWorkstation()
    const knast = useWorkstationMine()
    const knastOptions = useWorkstationOptions()
    const { operationalStatus, ControlPanel } = useControlPanel(knast.data);
    const workstationOnpremMapping = useWorkstationOnpremMapping()
    const effectiveTags = useWorkstationEffectiveTags()

    const recentlyFrontendUpdateOnprem = frontendUpdatingOnprem && Date.now() - (frontendUpdatingOnprem?.getTime() || 0) < 5 * 1000;
    const updatingOnprem = recentlyFrontendUpdateOnprem || connectivityJobs.data?.disconnect?.state === JobStateRunning || connectivityJobs.data?.connect?.some((job): job is WorkstationConnectJob =>
        job !== undefined && job.state === JobStateRunning);

    const onpremState = updatingOnprem ? "updating" : !!effectiveTags.data?.tags?.length ? "activated" : "deactivated";

    const internetState = lastUpdateInternet && Date.now() - lastUpdateInternet.getTime() < 5 * 1000 ? "updating" : activeItemInUrlList ? "activated" : "deactivated";

    const isDVHSource = (host: string) => {
        return Object.entries(onpremMapping.data?.hosts ?? {}).find(([type, _]) => type === "tns")?.[1].some(h => h?.Host === host);
    }

    const injectExtraInfoToKnast = (knast: any, knastOptions?: WorkstationOptions, workstationOnpremMapping?: WorkstationOnpremAllowList
        , effectiveTags?: EffectiveTags, urlList?: WorkstationURLListForIdent, onpremState?: string, operationalStatus?: string, internetState?: string) => {
        const image = knastOptions?.containerImages?.find((img) => img?.image === knast.image);
        const machineType = knastOptions?.machineTypes?.find((type) => type?.machineType === knast?.config?.machineType);
        const onpremConfigured = workstationOnpremMapping && workstationOnpremMapping.hosts && workstationOnpremMapping.hosts.length > 0;

        return {
            ...knast, imageTitle: image?.labels["org.opencontainers.image.title"] || "Ukjent miljø",
            machineTypeInfo: machineType, workstationOnpremMapping: workstationOnpremMapping?.hosts.map(it => ({ host: it, isDVHSource: isDVHSource(it) })), effectiveTags, internetUrls: urlList
            , onpremConfigured, onpremState, operationalStatus, internetState, validOnpremHosts: workstationOnpremMapping?.hosts.filter((it: string) => !knast.allowSSH || !isDVHSource(it)) || []
        }
    }

    const onActivateOnprem = async (enable: boolean) => {
        const validHosts = workstationOnpremMapping.data!!.hosts.filter((it: string) => !knastData?.allowSSH || !isDVHSource(it));
        if (enable && validHosts.length === 0) {
            return;
        }
        try {
            setFrontendUpdatingOnprem(new Date());
            await createConnectivityWorkflow.mutateAsync(enable ? { hosts: validHosts } : { hosts: [] })
        } catch (e) {
            console.error("Error in onActivateOnprem:", e);
            setOnpremError("Ukjent feil");
        }
    }

    const onActivateInternet = async (enable: boolean) => {
        setLastUpdateInternet(new Date());
        try {
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
            setLastUpdateInternet(undefined);
        } catch (e) {
            console.error("Error in onActivateInternet:", e);
            setOnpremError("Ukjent feil");
            setLastUpdateInternet(undefined);
        }
    }

    let knastData = knast.data
    if (knast.data) {
        knastData = injectExtraInfoToKnast(knast.data, knastOptions.data, workstationOnpremMapping?.data, effectiveTags?.data, urlList.data, onpremState, operationalStatus, internetState);
    }

    return !knast.data
        ? <div>Laster knast informasjon... <Loader /></div>
        : < div className="flex flex-col gap-4">
            {onpremError && <Alert variant="error" onClose={() => setOnpremError(null)}>{onpremError}</Alert>}
            <Quiz onClose={() => {
                setShowQuiz(false)
                setQuizReadCookie();
            }} show={showQuiz} />

            <ControlPanel
                knastInfo={knastData}
                onStartKnast={() => startKnast.mutate()}
                onStopKnast={() => stopKnast.mutate()}
                onSettings={() => setActiveForm("environment")}
                onActivateOnprem={() => onActivateOnprem(true)
                } onActivateInternet={() => onActivateInternet(true)} onDeactivateOnPrem={() => onActivateOnprem(false)}
                onDeactivateInternet={() => onActivateInternet(false)} onConfigureOnprem={() => setActiveForm("onprem")}
                onConfigureInternet={() => setActiveForm("internet")} />
            {knast.error && <Alert variant="error">{knast.error.message}</Alert>}
            <Tabs value={activeForm} onChange={tab => setActiveForm(tab as any)} className="w-230">
                <Tabs.List>
                    <Tabs.Tab
                        value="overview"
                        label="Oversikt"
                        icon={<LaptopIcon aria-hidden color={ColorInfo} />}
                    />
                    <Tabs.Tab
                        value="environment"
                        label="Konfigurasjon"
                        icon={<IconGear aria-hidden width={20} height={20}/>}
                    />
                    <Tabs.Tab
                        value="log"
                        label="Logger"
                        icon={<IconCircle aria-hidden width={20} height={20}/>}
                    />
                </Tabs.List>
                <Tabs.Panel value="overview" className="p-4">
                    <InfoForm knastInfo={knastData} operationalStatus={operationalStatus}
                        onActivateOnprem={() => onActivateOnprem(true)}
                        onActivateInternet={() => onActivateInternet(true)}
                        onDeactivateOnPrem={() => onActivateOnprem(false)}
                        onDeactivateInternet={() => onActivateInternet(false)}
                        onConfigureOnprem={() => {
                            setActiveForm("onprem");
                        }}
                        onConfigureInternet={() => {
                            setActiveForm("internet");
                        }} />
                </Tabs.Panel>
                <Tabs.Panel value="environment" className="p-4">
                    <ConfigureKnastForm form="environment" knastData={knastData} knastOptions={knastOptions} />
                </Tabs.Panel>
                <Tabs.Panel value="onprem" className="p-4">
                    <ConfigureKnastForm form="onprem" />
                </Tabs.Panel>
                <Tabs.Panel value="internet" className="p-4">
                    <ConfigureKnastForm form="internet" />
                </Tabs.Panel>
                <Tabs.Panel value="log" className="p-4">
                    <LogViewer />
                </Tabs.Panel>

            </Tabs>
        </div>
}
