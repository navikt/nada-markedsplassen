import {
    Workstation_STATE_RUNNING, Workstation_STATE_STOPPED, Workstation_STATE_STOPPING, WorkstationJobStateCompleted,
    WorkstationJobStateFailed,
    WorkstationJobStateRunning,
    WorkstationStartJob
} from "../../lib/rest/generatedDto";
import {Alert, Button, Loader} from "@navikt/ds-react";
import {PlayIcon, RocketIcon, StopIcon} from "@navikt/aksel-icons";
import {
    useCreateZonalTagBindingJob, usePollWorkstationStartJob,
    useStartWorkstation,
    useStopWorkstation,
    useWorkstationMine,
    useWorkstationStartJobs
} from "./queries";
import {useEffect, useState} from "react";

const WorkstationStatus = () => {
    // FIXME: consider reading out the errors and displaying them in the UI
    const workstation= useWorkstationMine()
    const startWorkstation = useStartWorkstation()
    const stopWorkstation= useStopWorkstation()
    const startWorkstationJob = usePollWorkstationStartJob(startWorkstation.data?.id.toString() || "")
    const createZonalTagBindings = useCreateZonalTagBindingJob()


    const handleOnStart = () => {
        startWorkstation.mutate()
    };

    // Pay attention to changes to the startWorkstationJob
    useEffect(() => {
        // When the workstation is started, we need to create the zonal tag bindings
        if (startWorkstationJob.data?.state === WorkstationJobStateCompleted) {
            createZonalTagBindings.mutate()
            // Reset so we stop polling
            startWorkstation.reset()
        }
    }, [startWorkstationJob]);

    const handleOnStop = () => {
        stopWorkstation.mutate()
        workstation.remove()
    };

    const handleOpenWorkstationWindow = () => {
        if (workstation.data?.host) {
            window.open(`https://${workstation.data?.host}/`, "_blank");
        }
    };

    const startStopButtons = (startButtonDisabled: boolean, stopButtonDisabled: boolean) => {
        return (
            <>
                <Button style={{backgroundColor: 'var(--a-green-500)'}} disabled={startButtonDisabled} onClick={handleOnStart}>
                    <div className="flex"><PlayIcon title="Start din Knast" fontSize="1.5rem"/>Start</div>
                </Button>
                <Button style={{backgroundColor: 'var(--a-red-500)'}} disabled={stopButtonDisabled} onClick={handleOnStop}>
                    <div className="flex"><StopIcon title="Stopp din Knast" fontSize="1.5rem"/>Stopp</div>
                </Button>
            </>
        )
    }

    if (!workstation) {
        return (
            <div className="flex flex-col gap-4 pt-4">
                <Alert variant={'warning'}>Du har ikke opprettet en Knast</Alert>
                <div className="flex gap-2">
                    {startStopButtons(true, true)}
                </div>
            </div>
        )
    }

    switch (startWorkstation.data?.state) {
        case WorkstationJobStateRunning:
            return (
                <div className="flex flex-col gap-4">
                    <p>Starter din Knast <Loader size="small" transparent/></p>
                    <div className="flex gap-2">
                        {startStopButtons(true, true)}
                    </div>
                </div>
            )
        case WorkstationJobStateFailed:
            return (
                <div className="flex flex-col gap-4 pt-4">
                    <Alert variant={'error'}>Klarte ikke å starte din Knast: {startWorkstation.data?.errors}</Alert>
                    <div className="flex gap-2">
                        {startStopButtons(false, true)}
                    </div>
                </div>
            )
    }

    switch (workstation.data?.state) {
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

export default WorkstationStatus;
