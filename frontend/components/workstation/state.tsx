import {
    Workstation_STATE_RUNNING, Workstation_STATE_STOPPED, Workstation_STATE_STOPPING,
    WorkstationJobStateFailed,
    WorkstationJobStateRunning,
    WorkstationStartJob
} from "../../lib/rest/generatedDto";
import {Alert, Button, Loader} from "@navikt/ds-react";
import {PlayIcon, RocketIcon, StopIcon} from "@navikt/aksel-icons";

interface WorkstationStateProps {
    workstationData?: any
    workstationStartJobs?: any
    handleOnStart: () => void
    handleOnStop: () => void
    handleOpenWorkstationWindow?: () => void
}

const WorkstationState = ({
                              workstationData,
                              workstationStartJobs,
                              handleOnStart,
                              handleOnStop,
                              handleOpenWorkstationWindow
                          }: WorkstationStateProps) => {
    const runningWorkstationStartJobs = workstationStartJobs?.jobs?.filter((job: any): job is WorkstationStartJob => job !== undefined && job.state === WorkstationJobStateRunning);

    const startStopButtons = (startButtonDisabled: boolean, stopButtonDisabled: boolean) => {
        return (
            <>
                <Button disabled={startButtonDisabled} onClick={handleOnStart}>
                    <div className="flex"><PlayIcon title="Start din Knast" fontSize="1.5rem"/>Start</div>
                </Button>
                <Button disabled={stopButtonDisabled} onClick={handleOnStop}>
                    <div className="flex"><StopIcon title="Stopp din Knast" fontSize="1.5rem"/>Stopp</div>
                </Button>
            </>
        )
    }

    if (!workstationData) {
        return (
            <div className="flex flex-col gap-4 pt-4">
                <Alert variant={'warning'}>Du har ikke opprettet en Knast</Alert>
                <div className="flex gap-2">
                    {startStopButtons(true, true)}
                </div>
            </div>
        )
    }

    if (runningWorkstationStartJobs?.length ?? 0 > 0) {
        return (
            <div className="flex flex-col gap-4">
                <p>Starter din Knast <Loader size="small" transparent/></p>
                <div className="flex gap-2">
                    {startStopButtons(true, true)}
                </div>
            </div>
        )
    }

    if (workstationStartJobs?.jobs?.[0]?.state === WorkstationJobStateFailed) {
        return (
            <div className="flex flex-col gap-4 pt-4">
                <Alert variant={'error'}>Klarte ikke å starte din Knast: {workstationStartJobs.jobs[0].errors}</Alert>
                <div className="flex gap-2">
                    {startStopButtons(false, true)}
                </div>
            </div>
        );
    }

    switch (workstationData.state) {
        case Workstation_STATE_RUNNING:
            return (
                <div className="flex gap-2">
                    {startStopButtons(true, false)}
                    <Button onClick={handleOpenWorkstationWindow}>
                        <div className="flex"><RocketIcon title="a11y-title" fontSize="1.5rem"/>Åpne din Knast i nytt
                            vindu
                        </div>
                    </Button>
                </div>
            )
        case Workstation_STATE_STOPPING:
            return (
                <div className="flex flex-col gap-4">
                    <p>Stopper din Knast <Loader size="small" transparent/></p>
                    <div className="flex gap-2">
                        {startStopButtons(true, true)}
                    </div>
                </div>
            )
        case Workstation_STATE_STOPPED:
            return (
                <div>
                    <div className="flex gap-2">
                        {startStopButtons(false, true)}
                    </div>
                </div>
            )
    }
}

export default WorkstationState;
