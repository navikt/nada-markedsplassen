import { useControlPanel } from "./controlPanel";
import { useStartWorkstation, useStopWorkstation, useWorkstationEffectiveTags, useWorkstationMine, useWorkstationOnpremMapping, useWorkstationOptions } from "./queries";
import { Alert, Loader } from "@navikt/ds-react";
import React from "react";
import { InfoForm } from "./infoForm";
import { SettingsForm } from "./SettingsForm";
import { DatasourcesForm } from "./DatasourcesForm";
import { InternetOpeningsForm } from "./internetOpeningsForm";
import { EffectiveTags, WorkstationOnpremAllowList, WorkstationOptions } from "../../lib/rest/generatedDto";
import { UseQueryResult } from "@tanstack/react-query";
import { HttpError } from "../../lib/rest/request";

const injectExtraInfoToKnast = (knast: any, knastOptions?: WorkstationOptions, workstationOnpremMapping?: WorkstationOnpremAllowList, effectiveTags?: EffectiveTags) => {
    const image = knastOptions?.containerImages?.find((img) => img?.image === knast.image);
    const machineType = knastOptions?.machineTypes?.find((type) => type?.machineType === knast.config.machineType);
    return { ...knast, imageTitle: image?.labels["org.opencontainers.image.title"] || "Ukjent miljø", machineTypeInfo: machineType, workstationOnpremMapping};
}

const Knast = () => {
    const [activeForm, setActiveForm] = React.useState<"info" | "settings" | "onprem" | "internet">("info")
    const startKnast = useStartWorkstation()
    const stopKnast = useStopWorkstation()

    const knast = useWorkstationMine()
    const knastOptions = useWorkstationOptions()
    const { operationalStatus, ControlPanel } = useControlPanel(knast.data);
    const workstationOnpremMapping = useWorkstationOnpremMapping()
    const effectiveTags = useWorkstationEffectiveTags()

    if (knast.isLoading) {
        return <div>Lasting min knast <Loader /></div>
    }

    if (knast.isError) {
        return <Alert variant="error">Feil ved lasting av knast: {knast.error instanceof Error ? knast.error.message : "ukjent feil"}</Alert>
    }

    if (!knast.data) {
        return <div>Ingen knast funnet for bruker</div>
    }

    const knastData = injectExtraInfoToKnast(knast.data, knastOptions.data, workstationOnpremMapping?.data, effectiveTags?.data);

    return <div className="flex flex-col gap-4">
        <ControlPanel knastInfo={knastData} onStartKnast={() => startKnast.mutate()}
            onStopKnast={() => stopKnast.mutate()}
            onSettings={() => setActiveForm("settings")}
            onConnectOnprem={function (): void {
                throw new Error("Function not implemented.");
            }} onConnectInternet={function (): void {
                throw new Error("Function not implemented.");
            }} onDisconnectOnprem={function (): void {
                throw new Error("Function not implemented.");
            }} onDisconnectInternet={function (): void {
                throw new Error("Function not implemented.");
            }} onConfigureOnprem={() => setActiveForm("onprem")}
            onConfigureInternet={() => setActiveForm("internet")} />
        {activeForm === "info" && <InfoForm knastInfo={knastData} operationalStatus={operationalStatus} />}
        {activeForm === "settings" && <SettingsForm onSave={function (): void {
            throw new Error("Function not implemented.");
        }} onCancel={() => setActiveForm("info")} />}
        {activeForm === "onprem" && <DatasourcesForm onSave={() => { }} onCancel={() => setActiveForm("info")} />}
        {activeForm === "internet" && <InternetOpeningsForm onSave={() => { }} onCancel={() => setActiveForm("info")} />}
    </div>
}

export default Knast
