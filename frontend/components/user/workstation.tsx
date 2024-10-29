import {
    Alert,
    Button,
    Heading,
    Select,
    UNSAFE_Combobox,
    Textarea,
    Label,
    Link,
    Table,
    CopyButton,
    Pagination,
    Loader
} from "@navikt/ds-react"
import {PlayIcon, StopIcon, XMarkOctagonIcon, CheckmarkCircleIcon} from '@navikt/aksel-icons';
import LoaderSpinner from "../lib/spinner"
import {Fragment, useState} from "react";
import {
    FirewallTag,
    Workstation_STATE_RUNNING,
    Workstation_STATE_STARTING, Workstation_STATE_STOPPED,
    Workstation_STATE_STOPPING, WorkstationContainer as DTOWorkstationContainer, WorkstationMachineType,
    WorkstationOutput, WorkstationOptions, WorkstationLogs, WorkstationJobs, WorkstationJob, WorkstationJobStateRunning,
    Diff,
    WorkstationDiffContainerImage,
    WorkstationDiffMachineType,
    WorkstationDiffURLAllowList,
    WorkstationDiffOnPremAllowList, WorkstationJobStateCompleted, WorkstationJobStateFailed
} from "../../lib/rest/generatedDto";
import {ExternalLink} from "@navikt/ds-icons";
import {
    createWorkstationJob,
    startWorkstation,
    stopWorkstation,
    useGetWorkstation,
    useGetWorkstationJobs,
    useConditionalWorkstationLogs,
    useGetWorkstationOptions
} from "../../lib/rest/workstation";
import {formatDistanceToNow} from 'date-fns';

interface WorkstationJobsStateProps {
    workstationJobs?: any
}

const WorkstationDiffDescriptions: { [key: string]: string } = {
    [WorkstationDiffContainerImage]: "Kjøremiljø",
    [WorkstationDiffMachineType]: "Maskin type",
    [WorkstationDiffURLAllowList]: "URL Filter",
    [WorkstationDiffOnPremAllowList]: "On-prem kilder",
};

interface DiffViewerProps {
    diff: { [key: string]: Diff | undefined };
}

const DiffViewerComponent: React.FC<DiffViewerProps> = ({diff}) => {
    console.log(diff)
    if (!diff || Object.keys(diff).length === 0) {
        return <div>Ingen endringer å vise.</div>;
    }

    return (
        <div className="diff-viewer">
            {Object.entries(diff).map(([key, value]) => {
                return (
                    <div key={key}>
                        <Heading size="xsmall">{WorkstationDiffDescriptions[key]}</Heading>
                        {value?.value ? (
                            <p>{value.value}</p>
                        ) : (
                            <div>
                                {(value?.added?.length ?? 0) > 0 && (
                                    <p style={{color: 'green'}}>{value?.added.join(', ')}</p>
                                )}
                                {(value?.removed?.length ?? 0) > 0 && (
                                    <p style={{color: 'red'}}>{value?.removed.join(', ')}</p>
                                )}
                            </div>
                        )}
                    </div>
                );
            })}
        </div>
    );
};

interface JobViewerProps {
    job: WorkstationJob | undefined;
}

const JobViewerComponent: React.FC<JobViewerProps> = ({job}) => {
    if (!job) {
        return;
    }

    return (
        <div>
            {job.machineType && (
                <>
                    <Heading size="xsmall">{WorkstationDiffDescriptions[WorkstationDiffMachineType]}</Heading>
                    <p>{job.machineType}</p>
                </>
            )}
            {job.containerImage && (
                <>
                    <Heading size="xsmall">{WorkstationDiffDescriptions[WorkstationDiffContainerImage]}</Heading>
                    <p>{job.containerImage}</p>
                </>
            )}
            {job.urlAllowList && job.urlAllowList.length > 0 && (
                <>
                    <Heading size="xsmall">{WorkstationDiffDescriptions[WorkstationDiffURLAllowList]}</Heading>
                    <p>{job.urlAllowList.join(', ')}</p>
                </>
            )}
            {job.onPremAllowList && job.onPremAllowList.length > 0 && (
                <>
                    <Heading size="xsmall">{WorkstationDiffDescriptions[WorkstationDiffOnPremAllowList]}</Heading>
                    <p>{job.onPremAllowList.join(', ')}</p>
                </>
            )}
        </div>
    );
};

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

interface WorkstationLogStateProps {
    workstationLogs?: any
}

const WorkstationLogState = ({workstationLogs}: WorkstationLogStateProps) => {
    const [page, setPage] = useState(1);
    const rowsPerPage = 10;

    if (!workstationLogs || workstationLogs.proxyDeniedHostPaths.length === 0) {
        return (
            <div className="flex flex-col gap-4 pt-4">
                <Alert variant={'warning'}>Ingen loggdata tilgjengelig</Alert>
            </div>
        )
    }

    let pageData = workstationLogs.proxyDeniedHostPaths.slice((page - 1) * rowsPerPage, page * rowsPerPage);
    return (
        <div className="grid gap-4">
            <Table zebraStripes size="medium">
                <Table.Header>
                    <Table.Row>
                        <Table.HeaderCell scope="col">URL</Table.HeaderCell>
                        <Table.HeaderCell scope="col"></Table.HeaderCell>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    {pageData.map((url: any, i: number) => (
                        <Table.Row key={i + url.HTTPRequest.URL.Host}>
                            <Table.DataCell>{`${url.HTTPRequest.URL.Host}${url.HTTPRequest.URL.Path}`}</Table.DataCell>
                            <Table.DataCell><CopyButton
                                copyText={`${url.HTTPRequest.URL.Host}${url.HTTPRequest.URL.Path}`}/></Table.DataCell>
                        </Table.Row>

                    ))}
                </Table.Body>
            </Table>
            <Pagination
                page={page}
                onPageChange={setPage}
                count={Math.ceil(workstationLogs.proxyDeniedHostPaths.length / rowsPerPage)}
                size="small"
            />
        </div>

    )
}

interface WorkstationStateProps {
    workstationData?: any
    handleOnStart: () => void
    handleOnStop: () => void
}

const WorkstationState = ({workstationData, handleOnStart, handleOnStop}: WorkstationStateProps) => {
    const startStopButtons = (startButtonDisabled: boolean, stopButtonDisabled: boolean) => {
        return (
            <div className="flex gap-2">

                <Button disabled={startButtonDisabled} onClick={handleOnStart}>
                    <div className="flex"><PlayIcon title="a11y-title" fontSize="1.5rem"/>Start</div>
                </Button>
                <Button disabled={stopButtonDisabled} onClick={handleOnStop}>
                    <div className="flex"><StopIcon title="a11y-title" fontSize="1.5rem"/>Stopp</div>
                </Button>
            </div>
        )
    }

    if (!workstationData) {
        return (
            <div className="flex flex-col gap-4 pt-4">
                <Alert variant={'warning'}>Du har ikke opprettet en arbeidsstasjon</Alert>
                {startStopButtons(true, true)}
            </div>
        )
    }

    switch (workstationData.state) {
        case Workstation_STATE_STARTING:
            return (
                <div className="flex flex-col gap-4">
                    <p>Starter arbeidsstasjon <Loader size="small" transparent/></p>
                    {startStopButtons(true, true)}
                </div>
            )
        case Workstation_STATE_RUNNING:
            return (
                <div>
                    {startStopButtons(true, false)}
                </div>
            )
        case Workstation_STATE_STOPPING:
            return (
                <div className="flex flex-col gap-4">
                    <p>Stopper arbeidsstasjon <Loader size="small" transparent/></p>
                    {startStopButtons(true, true)}
                </div>
            )
        case Workstation_STATE_STOPPED:
            return (
                <div>
                    {startStopButtons(false, true)}
                </div>
            )
    }
}

interface WorkstationContainerProps {
    workstation?: WorkstationOutput;
    workstationOptions?: WorkstationOptions | null;
    workstationLogs?: WorkstationLogs | null;
    workstationJobs?: WorkstationJobs | null;
}

const WorkstationContainer = ({
                                  workstation,
                                  workstationOptions,
                                  workstationLogs,
                                  workstationJobs
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

    function toMultilineString(urls: string[] | string | null) {
        if (typeof (urls) == "string") return urls
        else if (urls === null) return ""

        return urls.join("\n")
    }

    return (
        <div className="flex flex-col gap-8">
            <p>Her kan du opprette og gjøre endringer på din personlige arbeidsstasjon</p>
            <div className="flex">
                <form className="basis-1/2 border-x p-4"
                      onSubmit={handleOnCreateOrUpdate}>
                    <div className="flex flex-col gap-8">
                        {workstation === null ?
                            <Heading level="1" size="medium">Opprett arbeidsstasjon</Heading> :
                            <Heading level="1" size="medium">Endre arbeidsstasjon</Heading>
                        }
                        <Select defaultValue={workstation?.config?.machineType} label="Velg maskintype">
                            {workstationOptions?.machineTypes.map((type: WorkstationMachineType | undefined) => (
                                type ? <option key={type.machineType}
                                               value={type.machineType}>{type.machineType} (vCPU: {type.vCPU},
                                        memoryGB: {type.memoryGB})</option> :
                                    "Could not load machine type"
                            ))}
                        </Select>
                        <Select defaultValue={workstation?.config?.image} label="Velg containerImage">
                            {workstationOptions?.containerImages.map((image: DTOWorkstationContainer | undefined) => (
                                image ? <option key={image.image}
                                                value={image.image}>{image.labels?.['org.opencontainers.image.title'] || image.description}</option> :
                                    "Could not load container image"
                            ))}
                        </Select>
                        <UNSAFE_Combobox
                            label="Velg hvilke onprem-kilder du trenger åpninger mot"
                            options={workstationOptions ? workstationOptions.firewallTags?.map((o: FirewallTag | undefined) => (o ? {
                                label: `${o?.name}`,
                                value: o?.name,
                            } : {label: "Could not load firewall tag", value: "Could not load firewall tag"})) : []}
                            selectedOptions={Array.from(selectedFirewallHosts) as string[]}
                            isMultiSelect
                            onToggleSelected={handleFirewallTagChange}
                        />
                        <div className="flex gap-2 flex-col">
                            <Label>Oppgi hvilke internett-URL-er du vil åpne mot</Label>
                            <p className="pt-0">Du kan legge til opptil 2500 oppføringer i en URL-liste. Hver oppføring
                                må stå på en egen linje uten mellomrom eller skilletegn. Oppføringer kan være kun
                                domenenavn (som matcher alle stier) eller inkludere en sti-komponent. <Link
                                    target="_blank"
                                    href="https://cloud.google.com/secure-web-proxy/docs/url-list-syntax-reference">Les
                                    mer om syntax her <ExternalLink/></Link></p>

                            <Textarea onChange={handleUrlListUpdate} defaultValue={toMultilineString(urlList)}
                                      size="medium" maxRows={2500} hideLabel label="Hvilke URL-er vil du åpne mot"
                                      resize/>
                        </div>
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
                        <WorkstationState workstationData={workstation} handleOnStart={handleOnStart}
                                          handleOnStop={handleOnStop}/>
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

    if (loading || loadingOptions || loadingJobs) return <LoaderSpinner/>;

    return (
        <WorkstationContainer
            workstation={workstation}
            workstationOptions={workstationOptions}
            workstationLogs={workstationLogs}
            workstationJobs={workstationJobs}
        />
    );
};
