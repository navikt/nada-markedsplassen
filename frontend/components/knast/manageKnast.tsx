import { useControlPanel } from "./controlPanel";
import { useActivateWorkstationURLListForIdent, useCreateWorkstationConnectivityWorkflow, useDeactivateWorkstationURLListForIdent, useStartWorkstation, useStopWorkstation, useUpdateWorkstationURLListItemForIdent, useWorkstationConnectivityWorkflow, useWorkstationEffectiveTags, useWorkstationExists, useWorkstationJobs, useWorkstationLogs, useWorkstationMine, useWorkstationOnpremMapping, useWorkstationOptions, useWorkstationResyncJobs, useWorkstationURLListForIdent } from "./queries";
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
import { FileSearchIcon, LaptopIcon } from "@navikt/aksel-icons";
import { LogViewer } from "./widgets/logViewer";
import { ColorInfo } from "./designTokens";
import { ConfigureKnastForm } from "./configureKnast";


export const ManageKnastPage = () => {
    const [activeForm, setActiveForm] = React.useState<"overview" | "environment" | "onprem" | "internet" | "log">("overview")
    const createConnectivityWorkflow = useCreateWorkstationConnectivityWorkflow();
    const activateUrls = useActivateWorkstationURLListForIdent();
    const deactivateUrls = useDeactivateWorkstationURLListForIdent();
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
    const logs = useWorkstationLogs()
    const blockedUrls = logs.data?.proxyDeniedHostPaths ?? [];
    const oneHourAgo = new Date(Date.now() - 60 * 60 * 1000);
    
    const aggregatedLogs = (logs?.data?.proxyDeniedHostPaths as any[] ?? []).map(log => ({
        timestamp: log.Timestamp,
        type: "URL blokkert",
        message: `${log.HTTPRequest?.URL.Host}${log.HTTPRequest?.URL.Path}`
    })).concat((connectivityJobs?.data?.connect ?? []).map((it: any) => it.errors.map((error: string) => ({
            timestamp: it.startTime,
            type: "Tilkoblingsfeil",
            message: `${it.host}: ${error}`
        })))).flat()
        .filter((it: any) => new Date(it.timestamp).getTime() > oneHourAgo.getTime())
        .sort((a: any, b: any) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
    
    const logsNumber = aggregatedLogs.length >99? "99+": aggregatedLogs.length;


    const recentlyFrontendUpdateOnprem = frontendUpdatingOnprem && Date.now() - (frontendUpdatingOnprem?.getTime() || 0) < 5 * 1000;
    const updatingOnprem = recentlyFrontendUpdateOnprem || connectivityJobs.data?.disconnect?.state === JobStateRunning || connectivityJobs.data?.connect?.some((job): job is WorkstationConnectJob =>
        job !== undefined && job.state === JobStateRunning);

    const onpremState = updatingOnprem ? "updating" : !!effectiveTags.data?.tags?.length ? "activated" : "deactivated";

    const internetState = lastUpdateInternet && Date.now() - lastUpdateInternet.getTime() < 5 * 1000 ? "updating" : activeItemInUrlList ? "activated" : "deactivated";

    const isDVHSource = (host: string) => {
        return Object.entries(onpremMapping.data?.hosts ?? {}).find(([type, _]) => type === "tns")?.[1].some(h => h?.Host === host);
    }

    const injectExtraInfoToKnast = (knast: any, knastOptions?: WorkstationOptions, workstationOnpremMapping?: WorkstationOnpremAllowList
        , effectiveTags?: EffectiveTags, urlList?: WorkstationURLListForIdent, onpremState?: string, operationalStatus?: string, internetState?: string, blockedUrls?: any[]) => {
        const image = knastOptions?.containerImages?.find((img) => img?.image === knast.image);
        const machineType = knastOptions?.machineTypes?.find((type) => type?.machineType === knast?.config?.machineType);
        const onpremConfigured = workstationOnpremMapping && workstationOnpremMapping.hosts && workstationOnpremMapping.hosts.length > 0;

        return {
            ...knast, imageTitle: image?.labels["org.opencontainers.image.title"] || "Ukjent miljø",
            machineTypeInfo: machineType, workstationOnpremMapping: workstationOnpremMapping?.hosts.map(it => ({ host: it, isDVHSource: isDVHSource(it) })), effectiveTags, internetUrls: urlList
            , onpremConfigured, onpremState, operationalStatus, internetState, validOnpremHosts: workstationOnpremMapping?.hosts.filter((it: string) => !knast.allowSSH || !isDVHSource(it)) || []
            , blockedUrls
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
                const urlsToDeactivate = urlList.data!!.items.filter(it => it?.selected).map(it => it?.id!!);
                await deactivateUrls.mutateAsync(urlsToDeactivate);
            } else {
                const urlsToActivate = urlList.data!!.items.filter(it => it?.selected && new Date(it.expiresAt) < new Date()).map(it => it?.id!!);
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
        knastData = injectExtraInfoToKnast(knast.data, knastOptions.data, workstationOnpremMapping?.data, effectiveTags?.data, urlList.data, onpremState, operationalStatus, internetState
            , blockedUrls
        );
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
            {knast.error && <Alert variant="error">{knast?.error?.message}</Alert>}
            <Tabs value={activeForm} onChange={tab => setActiveForm(tab as any)} className="w-230">
                <Tabs.List>
                    <Tabs.Tab
                        value="overview"
                        label="Oversikt"
                        icon={<LaptopIcon aria-hidden color={ColorInfo} width={22} height={22} />}
                    />
                    <Tabs.Tab
                        value="environment"
                        label="Innstillinger"
                        icon={<IconGear aria-hidden width={22} height={22} />}
                    />
                    <Tabs.Tab
                        value="log"
                        label={<div className="flex flex-row items-center">Logger{!!logsNumber && <div className="rounded-2xl bg-red-600 ml-1 text-white w-6 h-4 text-sm items-center flex flex-col justify-center">{logsNumber}</div>}</div>}
                        icon={<FileSearchIcon aria-hidden width={22} height={22} color={ColorInfo} />}
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
                        }} onShowLogs={() => setActiveForm("log")}
                        onConfigureSSH={() => setActiveForm("environment")}
                    />

                </Tabs.Panel>
                <Tabs.Panel value="environment" className="p-4">
                    <ConfigureKnastForm form="environment" knastData={knastData} knastOptions={knastOptions} />
                </Tabs.Panel>
                <Tabs.Panel value="onprem" className="p-4">
                    <ConfigureKnastForm form="onprem" knastData={knastData}/>
                </Tabs.Panel>
                <Tabs.Panel value="internet" className="p-4">
                    <ConfigureKnastForm form="internet" knastData={knastData}/>
                </Tabs.Panel>
                <Tabs.Panel value="log" className="p-4">
                    <LogViewer logs={aggregatedLogs} isLoading={connectivityJobs.isLoading || logs.isLoading} />
                </Tabs.Panel>

            </Tabs>
        </div>
}
