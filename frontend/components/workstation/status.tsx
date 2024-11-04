import {
    Heading,
} from "@navikt/ds-react"
import {
    WorkstationJobs,
    WorkstationLogs,
    WorkstationOptions,
    WorkstationOutput, WorkstationStartJobs
} from "../../lib/rest/generatedDto";
import {
    startWorkstation, stopWorkstation,
} from "../../lib/rest/workstation";
import WorkstationState from "../workstation/state";

interface WorkstationStatusProps {
    workstation?: WorkstationOutput;
    workstationOptions?: WorkstationOptions | null;
    workstationLogs?: WorkstationLogs | null;
    workstationJobs?: WorkstationJobs | null;
    workstationStartJobs?: WorkstationStartJobs | null;
}

const WorkstationStatus = ({workstation, workstationStartJobs }: WorkstationStatusProps) => {
    const handleOnStart = () => {
        startWorkstation().then(() => {
            console.log("ok")
        }).catch((e: any) => {
            console.log(e)
        })
    }

    const handleOnStop = () => {
        stopWorkstation().then(() => {
            console.log("ok")
        }).catch((e: any) => {
            console.log(e)
        })
    }

    const handleOpenWorkstationWindow = () => {
        window.open(`https://${workstation?.host}/`, "_blank")
    }

    return (
                <div>
                    <Heading level="1" size="medium">Status</Heading>
                    <WorkstationState workstationData={workstation}
                                      workstationStartJobs={workstationStartJobs}
                                      handleOnStart={handleOnStart}
                                      handleOnStop={handleOnStop}
                                      handleOpenWorkstationWindow={handleOpenWorkstationWindow}/>
                </div>
    )
}

export default WorkstationStatus;
