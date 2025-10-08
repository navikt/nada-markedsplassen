import { useState } from "react"
import { useKnastControlPanel } from "./controlPanel"
import { InfoForm } from "./infoForm"
import { SettingsForm } from "./SettingsForm"
import { set } from "lodash"
import { DatasourcesForm } from "./DatasourcesForm"


const Knast = () => {
    const {operationalStatus, environment, connectivity, activeForm, setActiveForm, logs, component: ControlPanel} = useKnastControlPanel()
    return <div className="flex flex-col gap-4">
        <ControlPanel />
        {activeForm==="status" && <InfoForm operationalStatus={operationalStatus} logs={logs} />}
        {activeForm==="settings" && <SettingsForm onSave={() => setActiveForm("status")} onCancel={() => setActiveForm("status")} />}
        {activeForm==="sources" && <DatasourcesForm onSave={() => setActiveForm("status")} onCancel={() => setActiveForm("status")} />}
    </div>
}

export default Knast
    