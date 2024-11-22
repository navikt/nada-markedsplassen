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
import MachineTypeSelector from "./machineTypeSelector";
import ContainerImageSelector from "./containerImageSelector";
import FirewallTagSelector from "./firewallTagSelector";
import UrlListInput from "./urlListInput";
import GlobalAllowUrlListInput from "./globalAllowURLListInput";

interface WorkstationInputFormProps {
    workstation?: WorkstationOutput;
    workstationOptions?: WorkstationOptions | null;
    workstationLogs?: WorkstationLogs | null;
    workstationJobs?: WorkstationJobs | null;
    workstationStartJobs?: WorkstationStartJobs | null;
    refetchWorkstationJobs: () => void;
    incrementUnreadJobsCounter: () => void;
    setActiveTab: (tab: string) => void;
}

const WorkstationInputForm = (props: WorkstationInputFormProps) => {
    const {
        workstation,
        workstationOptions,
        workstationJobs,
        refetchWorkstationJobs,
        incrementUnreadJobsCounter,
        setActiveTab,
    } = props;

    const machineTypeRef = useRef<HTMLSelectElement>(null);

    const [selectedFirewallHosts, setSelectedFirewallHosts] = useState(new Set(workstation?.config?.firewallRulesAllowList))
    const [urlList, setUrlList] = useState(workstation ? workstation.urlAllowList : [])
    const [disableGlobalURLAllowList, setDisableGlobalURLAllowList] = useState(workstation?.config?.disableGlobalURLAllowList ?? false)
    const [containerImage, setContainerImage] = useState(workstationOptions?.containerImages?.[0]?.image ?? "");
    const runningJobs = workstationJobs?.jobs?.filter((job): job is WorkstationJob => job !== undefined && job.state === WorkstationJobStateRunning);

    const handleDisableGlobalURLAllowList = (value: any) => {
        setDisableGlobalURLAllowList(value === "true")
    }

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
            containerImage: containerImage,
            onPremAllowList: Array.from(selectedFirewallHosts),
            urlAllowList: urlList,
            disableGlobalURLAllowList: disableGlobalURLAllowList
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
                    <p>Du kan <strong>når som helst gjøre endringer på din Knast</strong>, f.eks, hvis du trenger en større maskintype, ønsker å prøve et annet utviklingsmiljø, eller trenger nye åpninger.</p>
                    <p>All data som er lagret under <strong>/home</strong> vil lagres på tvers av endringer</p>
                    <MachineTypeSelector ref={machineTypeRef}/>
                    <ContainerImageSelector
                        containerImages={(workstationOptions?.containerImages ?? []).filter((image): image is DTOWorkstationContainer => image !== undefined)}
                        defaultValue={workstation?.config?.image}
                        onChange={(event) => setContainerImage(event.target.value)}
                        onDocumentationLinkClick={() => setActiveTab("dokumentasjon")}
                    />
                    <FirewallTagSelector
                        firewallTags={(workstationOptions?.firewallTags ?? []).filter((tag): tag is FirewallTag => tag !== undefined)}
                        selectedFirewallHosts={selectedFirewallHosts}
                        onToggleSelected={handleFirewallTagChange}
                    />
                    <UrlListInput urlList={urlList}
                                  onUrlListUpdate={handleUrlListUpdate}
                    />
                    <GlobalAllowUrlListInput urlList={workstationOptions?.globalURLAllowList ?? ["Klarte ikke hente listen."]}
                                                defaultValue={disableGlobalURLAllowList ? "true" : "false"}
                                             onDisableGlobalURLAllowList={handleDisableGlobalURLAllowList}
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
