import {
    Alert,
    Button,
    Heading,
    Loader,
    Table,
} from "@navikt/ds-react"
import {
    CheckmarkCircleIcon,
    PlayIcon,
    StopIcon,
    XMarkOctagonIcon,
    RocketIcon
} from '@navikt/aksel-icons';
import LoaderSpinner from "../lib/spinner"
import {Fragment, useState} from "react";
import {
    Diff,
    FirewallTag,
    Workstation_STATE_RUNNING,
    Workstation_STATE_STOPPED,
    Workstation_STATE_STOPPING,
    WorkstationContainer as DTOWorkstationContainer,
    WorkstationJob, WorkstationJobs,
    WorkstationJobStateCompleted,
    WorkstationJobStateFailed,
    WorkstationJobStateRunning, WorkstationLogs,
    WorkstationMachineType, WorkstationOptions,
    WorkstationOutput, WorkstationStartJob, WorkstationStartJobs
} from "../../lib/rest/generatedDto";
import {
    createWorkstationJob, startWorkstation, stopWorkstation,
    useConditionalWorkstationLogs,
    useGetWorkstation,
    useGetWorkstationJobs,
    useGetWorkstationOptions, useGetWorkstationStartJobs
} from "../../lib/rest/workstation";
import {formatDistanceToNow} from 'date-fns';
import MachineTypeSelector from "../workstation/machineTypeSelector";
import ContainerImageSelector from "../workstation/containerImageSelector";
import FirewallTagSelector from "../workstation/firewallTagSelector";
import UrlListInput from "../workstation/urlListInput";
import DiffViewerComponent from "../workstation/diffViewer";
import JobViewerComponent from "../workstation/jobViewer";
import WorkstationLogState from "../workstation/logState";

interface WorkstationJobsStateProps {
    workstationJobs?: any
}

const WorkstationJobsState = ({workstationJobs}: WorkstationJobsStateProps) => {
    if (!workstationJobs || !workstationJobs.jobs || workstationJobs.jobs.length === 0) {
        return (
            <div className="flex flex-col gap-4 pt-4">
                <Alert variant={'warning'}>Ingen endringer</Alert>
            </div>
        )
    }

    return (
        <div className="grid gap-4">
            <Table zebraStripes size="medium">
                <Table.Header>
                    <Table.Row>
                        <Table.HeaderCell scope="col">Start tid</Table.HeaderCell>
                        <Table.HeaderCell scope="col">Status</Table.HeaderCell>
                        <Table.HeaderCell scope="col">Endringer</Table.HeaderCell>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    {workstationJobs.jobs.map((job: WorkstationJob, i: number) => (
                        <Table.Row key={i}>
                            <Table.DataCell>{formatDistanceToNow(new Date(job.startTime), {addSuffix: true})}</Table.DataCell>
                            <Table.DataCell>
                                {job.state === WorkstationJobStateRunning ? (
                                    <Fragment>
                                        Pågår <Loader size="xsmall" title="Pågår"/>
                                    </Fragment>
                                ) : job.state === WorkstationJobStateCompleted ? (
                                    <Fragment>
                                        Ferdig <CheckmarkCircleIcon title="Ferdig" fontSize="1.5rem"/>
                                    </Fragment>
                                ) : job.state === WorkstationJobStateFailed ? (
                                    <Fragment>
                                        Feilet <XMarkOctagonIcon title="feilet" fontSize="1.5rem"/>
                                    </Fragment>
                                ) : (
                                    job.state
                                )}
                            </Table.DataCell>
                            <Table.DataCell>
                                {job.diff && <DiffViewerComponent diff={job.diff}/>}
                                {!job.diff && <JobViewerComponent job={job}/>}
                            </Table.DataCell>
                        </Table.Row>
                    ))}
                </Table.Body>
            </Table>
        </div>
    )
}

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
                    <div className="flex"><PlayIcon title="a11y-title" fontSize="1.5rem"/>Start</div>
                </Button>
                <Button disabled={stopButtonDisabled} onClick={handleOnStop}>
                    <div className="flex"><StopIcon title="a11y-title" fontSize="1.5rem"/>Stopp</div>
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
                        <div className="flex"><RocketIcon title="a11y-title" fontSize="1.5rem"/>Åpne din Knast i nytt vindu</div>
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

export const Workstation = () => {
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
