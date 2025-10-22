import { useControlPanel } from "./controlPanel";
import { useCheckOnpremConnectivity, useCreateWorkstationConnectivityWorkflow, useStartWorkstation, useStopWorkstation, useWorkstationConnectivityWorkflow, useWorkstationEffectiveTags, useWorkstationMine, useWorkstationOnpremMapping, useWorkstationOptions } from "./queries";
import { Alert, Loader } from "@navikt/ds-react";
import React, { use, useEffect } from "react";
import { InfoForm } from "./infoForm";
import { SettingsForm } from "./SettingsForm";
import { DatasourcesForm } from "./DatasourcesForm";
import { InternetOpeningsForm } from "./internetOpeningsForm";
import { EffectiveTags, JobStateRunning, WorkstationConnectJob, WorkstationOnpremAllowList, WorkstationOptions } from "../../lib/rest/generatedDto";
import { UseQueryResult } from "@tanstack/react-query";
import { HttpError } from "../../lib/rest/request";
import { se } from "date-fns/locale";

const useOnpremState = ()=>{
    const [onpremState, setOnpremState] = React.useState<"activated" | "deactivated" | "activating" | "deactivating" | undefined>(undefined);
    const connectivity = useCheckOnpremConnectivity();

    useEffect(()=>{
        if(connectivity.isSuccess && connectivity.data !== undefined && (connectivity.data ==="activated" || connectivity.data === "deactivated")){
            setOnpremState(connectivity.data);
        }
    }, [connectivity.data, connectivity.isSuccess]);

    return {onpremState, setOnpremState};

}

const injectExtraInfoToKnast = (knast: any, knastOptions?: WorkstationOptions, workstationOnpremMapping?: WorkstationOnpremAllowList
    , effectiveTags?: EffectiveTags, onpremState?: string, operationalStatus?: string) => {
    const image = knastOptions?.containerImages?.find((img) => img?.image === knast.image);
    const machineType = knastOptions?.machineTypes?.find((type) => type?.machineType === knast.config.machineType);
    const onpremConfigured = workstationOnpremMapping && workstationOnpremMapping.hosts && workstationOnpremMapping.hosts.length > 0;

    return {
        ...knast, imageTitle: image?.labels["org.opencontainers.image.title"] || "Ukjent miljø",
        machineTypeInfo: machineType, workstationOnpremMapping, effectiveTags, onpremConfigured
        , onpremState, operationalStatus
    }
}

const Knast = () => {
    const [activeForm, setActiveForm] = React.useState<"info" | "settings" | "onprem" | "internet">("info")
    const createConnectivityWorkflow = useCreateWorkstationConnectivityWorkflow();
    const { onpremState, setOnpremState } = useOnpremState();

    const startKnast = useStartWorkstation()
    const stopKnast = useStopWorkstation()

    const knast = useWorkstationMine()
    const knastOptions = useWorkstationOptions()
    const { operationalStatus, ControlPanel } = useControlPanel(knast.data);
    const workstationOnpremMapping = useWorkstationOnpremMapping()
    const effectiveTags = useWorkstationEffectiveTags()

    console.log("onpremState", onpremState);

    const onActivateOnprem = (enable: boolean) => {
        createConnectivityWorkflow.mutate(enable ? workstationOnpremMapping.data!! : { hosts: [] }, {
            onSuccess: () => 
                setOnpremState(enable ? "activating" : "deactivating"),
            
        });
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

    const knastData = injectExtraInfoToKnast(knast.data, knastOptions.data, workstationOnpremMapping?.data, effectiveTags?.data, onpremState, operationalStatus);

    return <div className="flex flex-col gap-4">
        <ControlPanel
            knastInfo={knastData}
            onStartKnast={() => startKnast.mutate()}
            onStopKnast={() => stopKnast.mutate()}
            onSettings={() => setActiveForm("settings")}
            onActivateOnprem={() => onActivateOnprem(true)
            } onActivateInternet={function (): void {
                throw new Error("Function not implemented.");
            }} onDeactivateOnPrem={() => onActivateOnprem(false)}
            onDeactivateInternet={function (): void {
                throw new Error("Function not implemented.");
            }} onConfigureOnprem={() => setActiveForm("onprem")}
            onConfigureInternet={() => setActiveForm("internet")} />
        {activeForm === "info" && <InfoForm knastInfo={knastData} operationalStatus={operationalStatus}
            onActivateOnprem={() => onActivateOnprem(true)}
            onActivateInternet={function (): void {
                throw new Error("Function not implemented.");
            }}
            onDeactivateOnPrem={() => onActivateOnprem(false)}
            onDeactivateInternet={function (): void {
                throw new Error("Function not implemented.");
            }} onConfigureOnprem={function (): void {
                throw new Error("Function not implemented.");
            }} onConfigureInternet={function (): void {
                throw new Error("Function not implemented.");
            }} />}
        {activeForm === "settings" && <SettingsForm onSave={function (): void {
            throw new Error("Function not implemented.");
        }} onCancel={() => setActiveForm("info")} />}
        {activeForm === "onprem" && <DatasourcesForm onSave={() => { }} onCancel={() => setActiveForm("info")} />}
        {activeForm === "internet" && <InternetOpeningsForm onSave={() => { }} onCancel={() => setActiveForm("info")} />}
    </div>
}

export default Knast
