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
import {Fragment, useState} from "react";
import {createWorkstationJob} from "../../lib/rest/workstation";
import {Button, Heading, Loader} from "@navikt/ds-react";
import MachineTypeSelector from "./machineTypeSelector";
import ContainerImageSelector from "./containerImageSelector";
import FirewallTagSelector from "./firewallTagSelector";
import UrlListInput from "./urlListInput";

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

    const existingFirewallRules = workstation ? workstation.config ? workstation.config.firewallRulesAllowList : [] : []
    const [selectedFirewallHosts, setSelectedFirewallHosts] = useState(new Set(existingFirewallRules))
    const [urlList, setUrlList] = useState(workstation ? workstation.urlAllowList : [])
    const [machineType, setMachineType] = useState(workstationOptions?.machineTypes?.[0]?.machineType ?? "");
    const [containerImage, setContainerImage] = useState(workstationOptions?.containerImages?.[0]?.image ?? "");
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

    const handleOnCreateOrUpdate = async (event: any) => {
        event.preventDefault()

        const workstationInput: WorkstationInput = {
            machineType: machineType,
            containerImage: containerImage,
            onPremAllowList: Array.from(selectedFirewallHosts),
            urlAllowList: urlList
        };

        try {
            incrementUnreadJobsCounter();
            await createWorkstationJob(workstationInput)
            refetchWorkstationJobs();
        } catch (error) {
            console.error("Failed to create or update workstation job:", error)
        }
    }

    return (
        <div className="flex">
            <form className="basis-2/3 p-4" onSubmit={handleOnCreateOrUpdate}>
                <div className="flex flex-col gap-8">
                    <p>Du kan <strong>når som helst gjøre endringer på din Knast</strong>, f.eks, hvis du trenger en større maskintype, ønsker å prøve et annet utviklingsmiljø, eller trenger ny åpninger.</p>
                    <MachineTypeSelector
                        machineTypes={(workstationOptions?.machineTypes ?? []).filter((type): type is WorkstationMachineType => type !== undefined)}
                        defaultValue={workstation?.config?.machineType}
                        onChange={(event) => setMachineType(event.target.value)}
                    />
                    <ContainerImageSelector
                        containerImages={(workstationOptions?.containerImages ?? []).filter((image): image is DTOWorkstationContainer => image !== undefined)}
                        defaultValue={workstation?.config?.image}
                        onChange={(event) => setContainerImage(event.target.value)}
                    />
                    <FirewallTagSelector
                        firewallTags={(workstationOptions?.firewallTags ?? []).filter((tag): tag is FirewallTag => tag !== undefined)}
                        selectedFirewallHosts={selectedFirewallHosts}
                        onToggleSelected={handleFirewallTagChange}
                    />
                    <UrlListInput urlList={urlList} onUrlListUpdate={handleUrlListUpdate}
                                  defaultUrlList={workstationOptions?.defaultURLAllowList || []}
                    />
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
