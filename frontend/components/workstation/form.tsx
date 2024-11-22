import {
    WorkstationInput,
    WorkstationJob,
    WorkstationJobStateRunning,
} from "../../lib/rest/generatedDto";
import {Fragment, useRef} from "react";
import {createWorkstationJob, createWorkstationZonalTagBindingJob} from "../../lib/rest/workstation";
import {Button, Loader} from "@navikt/ds-react";
import MachineTypeSelector from "./formElements/machineTypeSelector";
import ContainerImageSelector from "./formElements/containerImageSelector";
import FirewallTagSelector from "./formElements/firewallTagSelector";
import UrlListInput from "./formElements/urlListInput";
import GlobalAllowUrlListInput from "./formElements/globalAllowURLListInput";
import {useWorkstation} from "./WorkstationStateProvider";

interface WorkstationInputFormProps {
    refetchWorkstationJobs: () => void;
    incrementUnreadJobsCounter: () => void;
}

const WorkstationInputForm = (props: WorkstationInputFormProps) => {
    const {
        refetchWorkstationJobs,
        incrementUnreadJobsCounter,
    } = props;

    const {workstation, workstationJobs} = useWorkstation()

    const machineTypeRef = useRef<HTMLSelectElement>(null);
    const globalUrlAllowListRef = useRef<HTMLFieldSetElement>(null);
    const containerImageRef = useRef<HTMLSelectElement>(null);
    const urlListRef = useRef<HTMLTextAreaElement>(null);
    const firewallHostsRef = useRef<HTMLInputElement>(null);

    const runningJobs = workstationJobs?.filter((job): job is WorkstationJob => job !== undefined && job.state === WorkstationJobStateRunning);

    const handleSubmit = async (event: any) => {
        event.preventDefault()

        const workstationInput: WorkstationInput = {
            machineType: machineTypeRef.current?.value ?? "",
            containerImage: containerImageRef.current?.value ?? "",
            onPremAllowList: firewallHostsRef.current?.value.split("\n").filter((host) => host.trim() !== "") ?? [],
            urlAllowList: urlListRef.current?.value.split("\n").filter((url) => url.trim() !== "") ?? [],
            disableGlobalURLAllowList: globalUrlAllowListRef.current?.querySelector("input:checked")?.value === "true",
        };

        try {
            incrementUnreadJobsCounter();
            await createWorkstationJob(workstationInput)
            refetchWorkstationJobs();
            await createWorkstationZonalTagBindingJob();
        } catch (error) {
            console.error("Failed to create or update workstation job:", error)
        }
    }

    return (
        <div className="flex">
            <form className="basis-2/3 p-4" onSubmit={handleSubmit}>
                <div className="flex flex-col gap-8">
                    <p>Du kan <strong>når som helst gjøre endringer på din Knast</strong>, f.eks, hvis du trenger en
                        større maskintype, ønsker å prøve et annet utviklingsmiljø, eller trenger nye åpninger.</p>
                    <p>All data som er lagret under <strong>/home</strong> vil lagres på tvers av endringer</p>
                    <MachineTypeSelector ref={machineTypeRef}/>
                    <ContainerImageSelector ref={containerImageRef}/>
                    <FirewallTagSelector ref={firewallHostsRef}/>
                    <UrlListInput ref={urlListRef}/>
                    <GlobalAllowUrlListInput ref={globalUrlAllowListRef}/>
                    <div className="flex flex-row gap-3">
                        {(workstation === null || workstation === undefined) ?
                            <Button type="submit" disabled={(runningJobs?.length ?? 0) > 0}>Opprett</Button> :
                            <Fragment>
                                <Button type="submit" disabled={(runningJobs?.length ?? 0) > 0}>Endre</Button>
                                {(runningJobs?.length ?? 0) > 0 && <Loader size="large" title="Venter..."/>}
                            </Fragment>
                        }
                    </div>
                </div>
            </form>
        </div>
    )
}

export default WorkstationInputForm;
