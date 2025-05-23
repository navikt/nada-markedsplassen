import {
    JobStateRunning,
    Workstation_STATE_RUNNING,
    Workstation_STATE_STARTING,
    WorkstationInput,
    WorkstationJob,
} from '../../lib/rest/generatedDto'
import {useState} from "react";
import { Alert, Button, Loader } from '@navikt/ds-react'
import ContainerImageSelector from "./formElements/containerImageSelector";
import {
    useCreateWorkstationJob,
    useWorkstationJobs,
    useWorkstationMine,
    useWorkstationOptions
} from "./queries";
import WorkstationChanges from "./WorkstationChanges";
import useMachineTypeSelector from "./formElements/machineTypeSelector";

const WorkstationAdministrate = () => {
    const workstation = useWorkstationMine()
    const options = useWorkstationOptions()
    const workstationJobs = useWorkstationJobs()
    const createWorkstationJob = useCreateWorkstationJob()
    const machineTypeSelector = useMachineTypeSelector(workstation.data?.config?.machineType || options.data?.machineTypes?.find(type => type !== undefined)?.machineType || "")

    const [selectedContainerImage, setSelectedContainerImage] = useState<string>(workstation.data?.config?.image || options.data?.containerImages?.find(image => image !== undefined)?.image || "");
    const [showRestartAlert, setShowRestartAlert] = useState(false);

    const runningJobs = workstationJobs.data?.jobs?.filter((job): job is WorkstationJob => job !== undefined && job.state === JobStateRunning);
    const isRunningOrStarting = workstation.data?.state === Workstation_STATE_RUNNING || workstation.data?.state === Workstation_STATE_STARTING;

    const handleSubmit = (event: any) => {
        event.preventDefault()

        const input: WorkstationInput = {
            machineType: machineTypeSelector.selectedMachineType,
            containerImage: selectedContainerImage ?? "",
        };

        try {
            createWorkstationJob.mutate(input)
            if (isRunningOrStarting) {
                setShowRestartAlert(true);
            }
        } catch (error) {
            console.error("Failed to create or update workstation job:", error)
        }
    }

    if (workstation.isLoading || options.isLoading || workstationJobs.isLoading || machineTypeSelector.isLoading) {
        return <Loader/>
    }

    return (
        <div className="flex gap-4">
            <div className="flex flex-col gap-4">
                {showRestartAlert && (
                  <Alert variant="info" closeButton onClose={() => setShowRestartAlert(false)}>
                      Du må starte Knasten på nytt for at endringene skal tre i kraft.
                  </Alert>
                )}
                <form className="p-4" onSubmit={handleSubmit}>
                    <div className="flex flex-col gap-8">
                        <p>Du kan <strong>når som helst gjøre endringer på din Knast</strong>, f.eks, hvis du trenger en
                            større maskintype, eller ønsker å prøve et annet utviklingsmiljø.</p>
                        <p>All data som er lagret under <strong>/home</strong> vil lagres på tvers av endringer</p>
                        <machineTypeSelector.MachineTypeSelector></machineTypeSelector.MachineTypeSelector>
                        <ContainerImageSelector initialContainerImage={selectedContainerImage}
                                                handleSetContainerImage={setSelectedContainerImage}/>
                        <div className="flex flex-row gap-3">
                            <Button type="submit" disabled={(runningJobs?.length ?? 0) > 0}>
                                Lagre endringer til Knasten
                            </Button>
                        </div>
                    </div>
                </form>
                <WorkstationChanges jobs={workstationJobs.data?.jobs?.filter((job): job is WorkstationJob => job !== undefined) || []}/>
            </div>
        </div>
    )
}

export default WorkstationAdministrate;
