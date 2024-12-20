import {
    WorkstationContainer,
    WorkstationInput,
    WorkstationJob,
    WorkstationJobStateRunning,
} from "../../lib/rest/generatedDto";
import {Fragment, useEffect, useRef, useState} from "react";
import {Button, Loader} from "@navikt/ds-react";
import MachineTypeSelector from "./formElements/machineTypeSelector";
import ContainerImageSelector from "./formElements/containerImageSelector";
import FirewallTagSelector from "./formElements/firewallTagSelector";
import GlobalAllowUrlListInput from "./formElements/globalAllowURLListInput";
import {
    useCreateWorkstationJob,
    useWorkstationExists,
    useWorkstationJobs,
    useWorkstationMine,
    useWorkstationOptions
} from "./queries";

interface WorkstationInputFormProps {
    incrementUnreadJobsCounter: () => void;
}

const WorkstationAdministrate = (props: WorkstationInputFormProps) => {
    const {
        incrementUnreadJobsCounter,
    } = props;

    const workstation = useWorkstationMine()
    const options = useWorkstationOptions()
    const workstationJobs = useWorkstationJobs()
    const createWorkstationJob = useCreateWorkstationJob()

    const [disabledGlobalURLAllowList, setDisabledGlobalURLAllowList] = useState<boolean>(workstation.data?.config?.disableGlobalURLAllowList || false);
    const [selectedFirewallTags, setSelectedFirewallTags] = useState<string[]>(workstation.data?.config?.firewallRulesAllowList || []);
    const [selectedContainerImage, setSelectedContainerImage] = useState<string>(workstation.data?.config?.image || options.data?.containerImages?.find(image => image !== undefined)?.image || "");
    const [selectedMachineType, setSelectedMachineType] = useState<string>(workstation.data?.config?.machineType || options.data?.machineTypes?.find(type => type !== undefined)?.machineType || "");

    const runningJobs = workstationJobs.data?.jobs?.filter((job): job is WorkstationJob => job !== undefined && job.state === WorkstationJobStateRunning);

    const handleSubmit = (event: any) => {
        event.preventDefault()

        const input: WorkstationInput = {
            machineType: selectedMachineType ?? "",
            containerImage: selectedContainerImage ?? "",
            onPremAllowList: selectedFirewallTags,
            urlAllowList: workstation.data?.urlAllowList ?? [],
            disableGlobalURLAllowList: disabledGlobalURLAllowList,
        };

        try {
            incrementUnreadJobsCounter();
            createWorkstationJob.mutate(input)
        } catch (error) {
            console.error("Failed to create or update workstation job:", error)
        }
    }

    if (workstation.isLoading || options.isLoading || workstationJobs.isLoading) {
        return <Loader/>
    }

    return (
        <div className="flex">
            <form className="basis-2/3 p-4" onSubmit={handleSubmit}>
                <div className="flex flex-col gap-8">
                    <p>Du kan <strong>når som helst gjøre endringer på din Knast</strong>, f.eks, hvis du trenger en
                        større maskintype, ønsker å prøve et annet utviklingsmiljø, eller trenger nye åpninger.</p>
                    <p>All data som er lagret under <strong>/home</strong> vil lagres på tvers av endringer</p>
                    <MachineTypeSelector initialMachineType={selectedMachineType}
                                         handleSetMachineType={setSelectedMachineType}/>
                    <ContainerImageSelector initialContainerImage={selectedContainerImage}
                                            handleSetContainerImage={setSelectedContainerImage}/>
                    <FirewallTagSelector onFirewallChange={setSelectedFirewallTags}/>
                    <GlobalAllowUrlListInput disabled={disabledGlobalURLAllowList}
                                             setDisabled={setDisabledGlobalURLAllowList}/>
                    <div className="flex flex-row gap-3">
                        <Button type="submit" disabled={(runningJobs?.length ?? 0) > 0}>Lagre endringer til
                            Knasten</Button>
                    </div>
                </div>
            </form>
        </div>
    )
}

export default WorkstationAdministrate;
