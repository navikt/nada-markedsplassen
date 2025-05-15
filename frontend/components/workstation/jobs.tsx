import {Alert, Loader, Table} from "@navikt/ds-react";
import {
    WorkstationJob,
    JobStateCompleted,
    JobStateFailed,
    JobStateRunning
} from "../../lib/rest/generatedDto";
import {formatDistanceToNow} from "date-fns";
import {Fragment} from "react";
import {CheckmarkCircleIcon, XMarkOctagonIcon} from "@navikt/aksel-icons";
import JobViewerComponent from "./jobViewer";
import DiffViewerComponent from "./DiffViewerComponent";
import {useWorkstationJobs} from "./queries";

const WorkstationJobsState = () => {
    const {data: workstationJobs} = useWorkstationJobs()

    if (!workstationJobs?.jobs || workstationJobs?.jobs.length === 0) {
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
                    {workstationJobs?.jobs?.filter((job): job is WorkstationJob => job !== undefined).map((job: WorkstationJob, i: number) => (
                        <Table.Row key={i}>
                            <Table.DataCell>{formatDistanceToNow(new Date(job.startTime), {addSuffix: true})}</Table.DataCell>
                            <Table.DataCell>
                                {job.state === JobStateRunning ? (
                                    <Fragment>
                                        Pågår <Loader size="xsmall" title="Pågår"/>
                                    </Fragment>
                                ) : job.state === JobStateCompleted ? (
                                    <Fragment>
                                        Ferdig <CheckmarkCircleIcon title="Ferdig" fontSize="1.5rem"/>
                                    </Fragment>
                                ) : job.state === JobStateFailed ? (
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

export default WorkstationJobsState;
