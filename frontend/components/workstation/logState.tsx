import {useState} from "react";
import {Alert, CopyButton, Pagination, Table} from "@navikt/ds-react";
import {formatDistanceToNow} from "date-fns";
import {useWorkstationLogs} from "../knast/queries";

const WorkstationLogState = () => {
    const {data: workstationLogs} = useWorkstationLogs()

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
                        <Table.HeaderCell scope="col">Blokkert URL</Table.HeaderCell>
                        <Table.HeaderCell scope="col">Tid</Table.HeaderCell>
                        <Table.HeaderCell scope="col">Kopier</Table.HeaderCell>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    {pageData.map((url: any, i: number) => (
                        <Table.Row key={i + url.HTTPRequest.URL.Host}>
                            <Table.DataCell title={`${url.HTTPRequest.URL.Host}${url.HTTPRequest.URL.Path}`}>{`${url.HTTPRequest.URL.Host}${url.HTTPRequest.URL.Path.length > 50 ? url.HTTPRequest.URL.Path.substring(0, 50) + '...' : url.HTTPRequest.URL.Path}`}</Table.DataCell>
                            <Table.DataCell>{isNaN(new Date(url.Timestamp).getTime()) ? 'Invalid date' : formatDistanceToNow(new Date(url.Timestamp), {addSuffix: true})}</Table.DataCell>
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

export default WorkstationLogState;
