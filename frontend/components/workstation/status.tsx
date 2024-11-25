import {
    Heading,
} from "@navikt/ds-react"
import {
    createWorkstationZonalTagBindingJob,
    startWorkstation, stopWorkstation
} from "../../lib/rest/workstation";
import WorkstationState from "../workstation/state";
import WorkstationZonalTagBindings from "./WorkstationZonalBindings";
import {useWorkstation} from "./WorkstationStateProvider";
import usePollingWorkstationStartJobs from "./hooks/usePollingWorkstationStartJobs";
import usePollingWorkstationZonalTagBindingJobs from "./hooks/usePollingWorkstationZonalTagBindingJobs";
import usePollingWorkstationLogs from "./hooks/usePollingWorkstationLogs";

const WorkstationStatus = () => {
    const {workstation} = useWorkstation();

    const {startPolling: startLogsPolling, stopPolling: stopLogsPolling} = usePollingWorkstationLogs()

    const {startPolling: startWorkstationStartJobsPolling} = usePollingWorkstationStartJobs({
        onComplete: async () => {
            await createWorkstationZonalTagBindingJob()
            startZonalTagBindingPolling()
        }
    });

    const {startPolling: startZonalTagBindingPolling} = usePollingWorkstationZonalTagBindingJobs();

    const handleOnStart = async () => {
        try {
            await startWorkstation();
            startWorkstationStartJobsPolling();
            startLogsPolling();
        } catch (error) {
            console.error("Failed to start workstation", error);
        }
    };

    const handleOnStop = async () => {
        try {
            await stopWorkstation();
            stopLogsPolling();
        } catch (error) {
            console.error("Failed to stop workstation", error);
        }
    };

    const handleOpenWorkstationWindow = () => {
        window.open(`https://${workstation?.host}/`, "_blank");
    };

    return (
        <div>
            <Heading level="1" size="medium">Status</Heading>
            <WorkstationState handleOnStart={handleOnStart} handleOnStop={handleOnStop}
                              handleOpenWorkstationWindow={handleOpenWorkstationWindow}/>
            <WorkstationZonalTagBindings/>
        </div>
    );
}

export default WorkstationStatus;
