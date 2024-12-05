import React, {useRef, useState} from "react";
import {Alert, AlertProps, Button, CopyButton, HelpText, Pagination, Table} from "@navikt/ds-react";
import {formatDistanceToNow} from "date-fns";
import {useUpdateUrlAllowList, useWorkstationJobs, useWorkstationLogs} from "./queries";
import UrlListInput from "./formElements/urlListInput";
import {WorkstationJob, WorkstationJobStateRunning, WorkstationURLList} from "../../lib/rest/generatedDto";
import {HttpError} from "../../lib/rest/request";

const WorkstationLogState = () => {
    const logs = useWorkstationLogs()

    const workstationJobs = useWorkstationJobs()
    const updateUrlAllowList = useUpdateUrlAllowList()

    const urlListRef = useRef<HTMLTextAreaElement>(null);

    const runningJobs = workstationJobs.data?.jobs?.filter((job): job is WorkstationJob => job !== undefined && job.state === WorkstationJobStateRunning);
    const [page, setPage] = useState(1);

    const rowsPerPage = 10;

    const handleSubmit = (event: any) => {
        event.preventDefault()

        const urls: WorkstationURLList = {
            urlAllowList: urlListRef.current?.value.split("\n").filter((url) => url.trim() !== "") ?? []
        }

        try {
            updateUrlAllowList.mutate(urls)
        } catch (error) {
            console.error("Failed to update url list:", error)
        }
    }

    if (!logs.data || logs.data.proxyDeniedHostPaths.length === 0) {
        return (
            <div className="flex flex-col gap-4 pt-4 alert-help-text">
                <form className="basis-2/3 p-4" onSubmit={handleSubmit}>
                    <div className="flex flex-col gap-8">
                        <UrlListInput ref={urlListRef}/>
                        <div className="flex flex-row gap-3">
                            <Button type="submit" disabled={(runningJobs?.length ?? 0) > 0}>Endre URL-er</Button>
                            {updateUrlAllowList.isError &&
                                <AlertWithCloseButton size="small" variant="error">
                                    Kunne ikke oppdatere URL-listen
                                    <HelpText title="Feilmelding">
                                        {(updateUrlAllowList.error as HttpError)?.message}
                                    </HelpText>
                                </AlertWithCloseButton>}
                            {updateUrlAllowList.isSuccess &&
                                <AlertWithCloseButton size="small" variant="success">URL-listen ble
                                    oppdatert</AlertWithCloseButton>}
                        </div>
                    </div>
                </form>
                <style>
                    {`
                    .alert-help-text {
                    --ac-help-text-popover-bg: var(--a-red-100);
                    --ac-help-text-button-color: var(--a-red-800);
                    --ac-help-text-button-hover-color: var(--a-surface-danger-hover);
                    --ac-help-text-button-focus-color: var(--a-surface-danger-subtle);
                    }
                    `}
                </style>
                <Alert variant={'warning'}>Ingen loggdata tilgjengelig</Alert>
            </div>
        )

    }

    let pageData = logs.data.proxyDeniedHostPaths.slice((page - 1) * rowsPerPage, page * rowsPerPage);
    return (
        <div className="grid gap-4">
            <form className="basis-2/3 p-4" onSubmit={handleSubmit}>
                <div className="flex flex-col gap-8">
                    <UrlListInput ref={urlListRef}/>
                    <div className="flex flex-row gap-3">
                        <Button type="submit" disabled={(runningJobs?.length ?? 0) > 0}>Endre URL-er</Button>
                        {updateUrlAllowList.isError &&
                            <Alert size="small" variant="error">Kunne ikke oppdatere URL-listen</Alert>}
                        {updateUrlAllowList.isSuccess &&
                            <Alert size="small" variant="success">URL-listen ble oppdatert</Alert>}
                    </div>
                </div>
            </form>
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
                            <Table.DataCell
                                title={`${url.HTTPRequest.URL.Host}${url.HTTPRequest.URL.Path}`}>{`${url.HTTPRequest.URL.Host}${url.HTTPRequest.URL.Path.length > 50 ? url.HTTPRequest.URL.Path.substring(0, 50) + '...' : url.HTTPRequest.URL.Path}`}</Table.DataCell>
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
                count={Math.ceil(logs.data.proxyDeniedHostPaths.length / rowsPerPage)}
                size="small"
            />
        </div>

    )
}

const AlertWithCloseButton = ({
                                  children,
                                  variant,
                                  size,
                              }: {
    children?: React.ReactNode;
    variant: AlertProps["variant"];
    size: AlertProps["size"];
}) => {
    const [show, setShow] = React.useState(true);

    return show ? (
        <Alert size={size} variant={variant} closeButton onClose={() => setShow(false)}>
            {children || "Content"}
        </Alert>
    ) : null;
};

export default WorkstationLogState;
