import {
    Button,
    Heading,
    Loader,
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
    WorkstationOutput, WorkstationStartJobs
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

interface WorkstationContainerProps {
    workstation?: WorkstationOutput;
    workstationOptions?: WorkstationOptions | null;
    workstationLogs?: WorkstationLogs | null;
    workstationJobs?: WorkstationJobs | null;
    workstationStartJobs?: WorkstationStartJobs | null;
}

const WorkstationContainer = ({
                                  workstation,
                                  workstationOptions,
                                  workstationLogs,
                                  workstationJobs,
                                  workstationStartJobs,
                              }: WorkstationContainerProps) => {
    const existingFirewallRules = workstation ? workstation.config ? workstation.config.firewallRulesAllowList : [] : []
    const [selectedFirewallHosts, setSelectedFirewallHosts] = useState(new Set(existingFirewallRules))
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

    const handleOnCreateOrUpdate = (event: any) => {
        event.preventDefault()
        // TODO: use state variables instead
        createWorkstationJob(
            {
                "machineType": event.target[0].value,
                "containerImage": event.target[1].value,
                "onPremAllowList": Array.from(selectedFirewallHosts),
                "urlAllowList": urlList
            }
        ).then(() => {
        }).catch((e: any) => {
            console.log(e)
        })
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
                <form className="basis-1/2 border-x p-4"
                      onSubmit={handleOnCreateOrUpdate}>
                    <div className="flex flex-col gap-8">
                        <Heading level="1" size="medium">{workstation ? "Endre Knast" : "Opprett Knast"}</Heading>
                        <MachineTypeSelector
                            machineTypes={(workstationOptions?.machineTypes ?? []).filter((type): type is WorkstationMachineType => type !== undefined)}
                            defaultValue={workstation?.config?.machineType}
                        />
                        <ContainerImageSelector
                            containerImages={(workstationOptions?.containerImages ?? []).filter((image): image is DTOWorkstationContainer => image !== undefined)}
                            defaultValue={workstation?.config?.image}
                        />
                        <FirewallTagSelector
                            firewallTags={(workstationOptions?.firewallTags ?? []).filter((tag): tag is FirewallTag => tag !== undefined)}
                            selectedFirewallHosts={selectedFirewallHosts}
                            onToggleSelected={handleFirewallTagChange}
                        />
                        <UrlListInput urlList={urlList} onUrlListUpdate={handleUrlListUpdate} defaultUrlList={workstationOptions?.defaultURLAllowList || []}/>
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
                    <div className="p-4">
                        <Heading level="1" size="medium">Logger</Heading>
                        <WorkstationLogState workstationLogs={workstationLogs}></WorkstationLogState>
                    </div>
                    <div className="p-4">
                        <Heading level="1" size="medium">Endringer</Heading>
                        <WorkstationJobsState workstationJobs={workstationJobs}></WorkstationJobsState>
                    </div>
                </div>
            </div>
        </div>
    )
}

export const Workstation: React.FC = () => {
    const {data: workstation, isLoading: loading} = useGetWorkstation();
    const {data: workstationOptions, isLoading: loadingOptions} = useGetWorkstationOptions();
    const {data: workstationJobs, isLoading: loadingJobs} = useGetWorkstationJobs();
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
        />
    );
};
