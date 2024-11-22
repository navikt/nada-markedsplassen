import {
    FirewallTag,
    WorkstationContainer as DTOWorkstationContainer,
    WorkstationInput,
    WorkstationJob,
    WorkstationJobs, WorkstationJobStateRunning,
    WorkstationLogs, WorkstationMachineType,
    WorkstationOptions,
    WorkstationOutput,
    WorkstationStartJobs
} from "../../lib/rest/generatedDto";
import {Fragment, useRef, useState} from "react";
import {createWorkstationJob, createWorkstationZonalTagBindingJob} from "../../lib/rest/workstation";
import {Button, Heading, Loader} from "@navikt/ds-react";
import MachineTypeSelector from "./formElements/machineTypeSelector";
import ContainerImageSelector from "./formElements/containerImageSelector";
import FirewallTagSelector from "./formElements/firewallTagSelector";
import UrlListInput from "./formElements/urlListInput";
import GlobalAllowUrlListInput from "./formElements/globalAllowURLListInput";

interface WorkstationInputFormProps {
    workstation?: WorkstationOutput;
    workstationOptions?: WorkstationOptions | null;
    workstationLogs?: WorkstationLogs | null;
    workstationJobs?: WorkstationJobs | null;
    workstationStartJobs?: WorkstationStartJobs | null;
    refetchWorkstationJobs: () => void;
    incrementUnreadJobsCounter: () => void;
}

const WorkstationInputForm = (props: WorkstationInputFormProps) => {
    const {
        workstation,
        workstationOptions,
        workstationJobs,
        refetchWorkstationJobs,
        incrementUnreadJobsCounter,
    } = props;

    const machineTypeRef = useRef<HTMLSelectElement>(null);
    const globalUrlAllowListRef = useRef<HTMLFieldSetElement>(null);
    const containerImageRef = useRef<HTMLSelectElement>(null);

    const [selectedFirewallHosts, setSelectedFirewallHosts] = useState(new Set(workstation?.config?.firewallRulesAllowList))
    const [urlList, setUrlList] = useState(workstation ? workstation.urlAllowList : [])
    const runningJobs = workstationJobs?.jobs?.filter((job): job is WorkstationJob => job !== undefined && job.state === WorkstationJobStateRunning);


    const handleUrlListUpdate = (event: any) => {
        setUrlList(event.target.value.split("\n"))
    }

    const handleFirewallTagChange = (tagValue: string, isSelected: boolean) => {
        if (isSelected) {
            setSelectedFirewallHosts(new Set(selectedFirewallHosts.add(tagValue)))
            return
        }
        selectedFirewallHosts.delete(tagValue)

        setSelectedFirewallHosts(new Set(selectedFirewallHosts))
    }

    const handleSubmit = async (event: any) => {
        event.preventDefault()

        const workstationInput: WorkstationInput = {
            machineType: machineTypeRef.current?.value ?? "",
            containerImage: containerImageRef.current?.value ?? "",
            onPremAllowList: Array.from(selectedFirewallHosts),
            urlAllowList: urlList,
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
                    <FirewallTagSelector
                        firewallTags={(workstationOptions?.firewallTags ?? []).filter((tag): tag is FirewallTag => tag !== undefined)}
                        selectedFirewallHosts={selectedFirewallHosts}
                        onToggleSelected={handleFirewallTagChange}
                    />
                    <UrlListInput urlList={urlList}
                                  onUrlListUpdate={handleUrlListUpdate}
                    />
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
