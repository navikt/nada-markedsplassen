import {
    Button,
    Heading,
    Loader,
    Tabs,
} from "@navikt/ds-react"
import LoaderSpinner from "../lib/spinner"
import {Fragment, useState} from "react";
import {
    FirewallTag,
    Workstation_STATE_RUNNING,
    WorkstationContainer as DTOWorkstationContainer,
    WorkstationJob, WorkstationJobs,
    WorkstationJobStateRunning, WorkstationLogs,
    WorkstationMachineType, WorkstationOptions,
    WorkstationOutput, WorkstationStartJobs, WorkstationInput
} from "../../lib/rest/generatedDto";
import {
    createWorkstationJob, startWorkstation, stopWorkstation,
    useConditionalWorkstationLogs,
    useGetWorkstation,
    useGetWorkstationJobs,
    useGetWorkstationOptions, useGetWorkstationStartJobs
} from "../../lib/rest/workstation";
import MachineTypeSelector from "../workstation/machineTypeSelector";
import ContainerImageSelector from "../workstation/containerImageSelector";
import FirewallTagSelector from "../workstation/firewallTagSelector";
import UrlListInput from "../workstation/urlListInput";
import WorkstationLogState from "../workstation/logState";
import WorkstationJobsState from "../workstation/jobs";
import WorkstationState from "../workstation/state";
import {CaptionsIcon, CogRotationIcon, GlobeIcon} from "@navikt/aksel-icons";
import PythonSetup from "../workstation/pythonSetup";

interface WorkstationContainerProps {
    workstation?: WorkstationOutput;
    workstationOptions?: WorkstationOptions | null;
    workstationLogs?: WorkstationLogs | null;
    workstationJobs?: WorkstationJobs | null;
    workstationStartJobs?: WorkstationStartJobs | null;
    refetchWorkstationJobs: () => void;
}

const WorkstationContainer = ({
                                  workstation,
                                  workstationOptions,
                                  workstationLogs,
                                  workstationJobs,
                                  workstationStartJobs,
    refetchWorkstationJobs,
                              }: WorkstationContainerProps) => {
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
            await createWorkstationJob(workstationInput)
            refetchWorkstationJobs();
        } catch (error) {
            console.error("Failed to create or update workstation job:", error)
        }
    }

    const handleOnStart = () => {
        startWorkstation().then(() => {
            console.log("ok")
        }).catch((e: any) => {
            console.log(e)
        })
    }

    const handleOnStop = () => {
        stopWorkstation().then(() => {
            console.log("ok")
        }).catch((e: any) => {
            console.log(e)
        })
    }

    const handleOpenWorkstationWindow = () => {
        window.open(`https://${workstation?.host}/`, "_blank")
    }

    return (
        <div className="flex flex-col gap-8">
            <p>Her kan du opprette og gjøre endringer på din personlige Knast</p>
            <div className="flex">
                <form className="basis-1/2 border-x p-4" onSubmit={handleOnCreateOrUpdate}>
                    <div className="flex flex-col gap-8">
                        <Heading level="1" size="medium">{workstation ? "Endre Knast" : "Opprett Knast"}</Heading>
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
                <div className="flex flex-col gap-4 basis-1/2">
                    <div className="flex flex-col border-1 p-4 gap-2">
                        <Heading level="1" size="medium">Status</Heading>
                        <WorkstationState workstationData={workstation}
                                          workstationStartJobs={workstationStartJobs}
                                          handleOnStart={handleOnStart}
                                          handleOnStop={handleOnStop}
                                          handleOpenWorkstationWindow={handleOpenWorkstationWindow}/>
                    </div>
                </div>
            </div>
            <div className="flex flex-col">
                <Tabs defaultValue="endringer">
                    <Tabs.List>
                        <Tabs.Tab
                            value="endringer"
                            label="Endringer"
                            icon={<CogRotationIcon aria-hidden />}
                        />
                        <Tabs.Tab
                            value="logger"
                            label="Blokkerte URLer"
                            icon={<CaptionsIcon aria-hidden />}
                        />
                        <Tabs.Tab
                            value="python"
                            label="Oppsett av Python"
                            icon={<GlobeIcon aria-hidden />}
                        />
                    </Tabs.List>
                    <Tabs.Panel value="endringer" className="p-4">
                        <WorkstationJobsState workstationJobs={workstationJobs}></WorkstationJobsState>
                    </Tabs.Panel>
                    <Tabs.Panel value="logger" className="p-4">
                        <WorkstationLogState workstationLogs={workstationLogs}></WorkstationLogState>
                    </Tabs.Panel>
                    <Tabs.Panel value="python" className="p-4">
                        <PythonSetup/>
                    </Tabs.Panel>
                </Tabs>
            </div>
        </div>
    )
}

export const Workstation: React.FC = () => {
    const {data: workstation, isLoading: loading} = useGetWorkstation();
    const {data: workstationOptions, isLoading: loadingOptions} = useGetWorkstationOptions();
    const {data: workstationJobs, isLoading: loadingJobs, refetch: refetchWorkstationJobs} = useGetWorkstationJobs();
    const isRunning = workstation?.state === Workstation_STATE_RUNNING;
    const {data: workstationLogs} = useConditionalWorkstationLogs(isRunning);
    const {data: workstationStartJobs} = useGetWorkstationStartJobs()

    if (loading || loadingOptions || loadingJobs) return <LoaderSpinner/>;

    return (
        <WorkstationContainer
            workstation={workstation}
            workstationOptions={workstationOptions}
            workstationLogs={workstationLogs}
            workstationJobs={workstationJobs}
            workstationStartJobs={workstationStartJobs}
            refetchWorkstationJobs={refetchWorkstationJobs}
        />
    );
};
