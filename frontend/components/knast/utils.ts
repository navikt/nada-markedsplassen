import { Workstation_STATE_RUNNING, Workstation_STATE_STARTING, Workstation_STATE_STOPPED, Workstation_STATE_STOPPING, WorkstationOutput } from "../../lib/rest/generatedDto"

export const GetOperationalStatus = (ws: WorkstationOutput | undefined): string =>{
    switch(ws?.state) {
        case Workstation_STATE_RUNNING:
            return "Running"
        case Workstation_STATE_STARTING:
            return "Starting"
        case Workstation_STATE_STOPPING:
            return "Stopping"
        case Workstation_STATE_STOPPED:
            return "Stopped"
        default:
            return "Unknown"
    }
}

export const GetKnastDailyCost = (knastInfo: any) => knastInfo?.machineTypeInfo?.hourlyCost ? `Kr ${(knastInfo.machineTypeInfo.hourlyCost * 24).toFixed(0)},-/døgn` : "Ukjent"