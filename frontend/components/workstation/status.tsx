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

const WorkstationStatus = () => {
    const {workstation} = useWorkstation();

    const handleOnStart = async () => {
        try {
            await startWorkstation();
            await createWorkstationZonalTagBindingJob();
        } catch (error) {
            console.error("Failed to start workstation", error);
        }
    };

    const handleOnStop = async () => {
        try {
            await stopWorkstation();
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
            <WorkstationState handleOnStart={handleOnStart} handleOnStop={handleOnStop} handleOpenWorkstationWindow={handleOpenWorkstationWindow} />
            <WorkstationZonalTagBindings />
        </div>
    );
}

export default WorkstationStatus;
