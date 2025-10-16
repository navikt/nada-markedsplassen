import { ControlPanel } from "./controlPanel";
import { useStartWorkstation, useStopWorkstation, useWorkstationMine } from "./queries";
import { Alert, Loader } from "@navikt/ds-react";
import React from "react";
import { InfoForm } from "./infoForm";
import { SettingsForm } from "./SettingsForm";
import { DatasourcesForm } from "./DatasourcesForm";
import { InternetOpeningsForm } from "./internetOpeningsForm";

const Knast = () => {
    const [activeForm, setActiveForm] = React.useState<"info" | "settings" | "onprem" | "internet">("info")
    const startKnast = useStartWorkstation()
    const stopKnast = useStopWorkstation()

    const knast = useWorkstationMine()
    if (knast.isLoading) {
        return <div>Lasting min knast <Loader /></div>
    }

    if (knast.isError) {
        return <Alert variant="error">Feil ved lasting av knast: {knast.error instanceof Error ? knast.error.message : "ukjent feil"}</Alert>
    }

    if (!knast.data) {
        return <div>Ingen knast funnet for bruker</div>
    }

    return <div className="flex flex-col gap-4">
        <ControlPanel knastInfo={knast?.data} onStartKnast={() => startKnast.mutate()}
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
        {activeForm === "info" && <InfoForm knastInfo={knast.data} />}
        {activeForm === "settings" && <SettingsForm onSave={function (): void {
            throw new Error("Function not implemented.");
        }} onCancel={() => setActiveForm("info")} />}
        {activeForm === "onprem" && <DatasourcesForm onSave={() => { }} onCancel={() => setActiveForm("info")} />}
        {activeForm === "internet" && <InternetOpeningsForm onSave={() => { }} onCancel={() => setActiveForm("info")} />}
    </div>
}

export default Knast
