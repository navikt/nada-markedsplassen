import {
    Heading,
} from "@navikt/ds-react"
import {
    EffectiveTag,
    EffectiveTags,
    FirewallTag,
    Workstation_STATE_RUNNING,
    WorkstationJobs,
    WorkstationLogs,
    WorkstationOptions,
    WorkstationOutput, WorkstationStartJobs, WorkstationZonalTagBindingJob, WorkstationZonalTagBindingJobs
} from "../../lib/rest/generatedDto";
import {
    createWorkstationZonalTagBindingJob,
    getWorkstation,
    startWorkstation, stopWorkstation
} from "../../lib/rest/workstation";
import WorkstationState from "../workstation/state";
import {useEffect, useState} from "react";
import WorkstationZonalTagBindings from "./WorkstationZonalBindings";

interface WorkstationStatusProps {
    workstation?: WorkstationOutput;
    workstationOptions?: WorkstationOptions | null;
    workstationLogs?: WorkstationLogs | null;
    workstationJobs?: WorkstationJobs | null;
    workstationStartJobs?: WorkstationStartJobs | null;
    workstationZonalTagBindingJobs?: WorkstationZonalTagBindingJobs | null;
    effectiveTags?: EffectiveTags | null;
}

const WorkstationStatus = ({
                               workstationStartJobs,
                               workstationZonalTagBindingJobs,
                               effectiveTags
                           }: WorkstationStatusProps) => {
    const [workstation, setWorkstation] = useState<WorkstationOutput | undefined>(undefined);
    const [isRunning, setIsRunning] = useState(false);

    useEffect(() => {
        const fetchWorkstation = async () => {
            try {
                const data = await getWorkstation();
                setWorkstation(data);
                setIsRunning(data.state === Workstation_STATE_RUNNING);
            } catch (error) {
                console.error("Failed to fetch workstation state", error);
            }
        };

        fetchWorkstation();
        const interval = setInterval(fetchWorkstation, 5000);

        return () => clearInterval(interval);
    }, []);

    const handleOnStart = async () => {
        try {
            await startWorkstation();
            await createWorkstationZonalTagBindingJob();
            setIsRunning(true);
        } catch (error) {
            console.error("Failed to start workstation", error);
        }
    };

    const handleOnStop = async () => {
        try {
            await stopWorkstation();
            setIsRunning(false);
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
            <WorkstationState
                workstationData={workstation}
                workstationStartJobs={workstationStartJobs}
                handleOnStart={handleOnStart}
                handleOnStop={handleOnStop}
                handleOpenWorkstationWindow={handleOpenWorkstationWindow}
                isRunning={isRunning}
            />
            <WorkstationZonalTagBindings
                jobs={workstationZonalTagBindingJobs?.jobs?.filter((job): job is WorkstationZonalTagBindingJob => job !== undefined)}
                effectiveTags={effectiveTags?.tags?.filter((tag): tag is EffectiveTag => tag !== undefined)}
                expectedTags={workstation?.config?.firewallRulesAllowList}
                workstationIsRunning={isRunning}
            />
        </div>
    );
}

export default WorkstationStatus;
