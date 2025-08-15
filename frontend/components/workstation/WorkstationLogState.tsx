import { Alert, AlertProps, Button, CopyButton, HelpText, Loader, Pagination, Table, Heading, Tabs } from '@navikt/ds-react';
import { formatDistanceToNow } from "date-fns";
import React, { useState } from 'react';
import { WorkstationURLList } from "../../lib/rest/generatedDto";
import { HttpError } from "../../lib/rest/request";
import { useUpdateUrlAllowList, useWorkstationLogs } from './queries';
import useWorkstationUrlEditor from './urlEditor/useUrlEditor';
import { ClockIcon, CogIcon, FileTextIcon } from '@navikt/aksel-icons';
import TimeRestrictedUrlEditor from './urlEditor/TimeRestrictedUrlEditor';

const WorkstationLogState = () => {
    const logs = useWorkstationLogs()
    const urlEditor = useWorkstationUrlEditor()
    const updateUrlAllowList = useUpdateUrlAllowList()

    const [page, setPage] = useState(1);
    const [activeTab, setActiveTab] = useState("timerestricted");

    const rowsPerPage = 10;

    const handleSubmit = (event: any) => {
        event.preventDefault()

        const urls: WorkstationURLList = {
            urlAllowList: urlEditor.urlList,
            disableGlobalAllowList: !urlEditor.keepGlobalAllowList,
            globalDenyList: []
        }

        try {
            updateUrlAllowList.mutate(urls)
        } catch (error) {
            console.error("Failed to update url list:", error)
        }
    }

    if (urlEditor.isLoading) {
        return <Loader size="large" title="Laster.." />
    }

    const pageData = logs.data?.proxyDeniedHostPaths?.slice((page - 1) * rowsPerPage, page * rowsPerPage) || [];

    return (
        <div className="flex flex-col gap-6">
            <div>
                <Heading level="2" size="medium" className="mb-4">Internettåpninger</Heading>

                <Tabs value={activeTab} onChange={setActiveTab}>
                    <Tabs.List>
                        <Tabs.Tab
                            value="timerestricted"
                            label="Tidsbegrensede åpninger"
                            icon={<ClockIcon aria-hidden />}
                        />
                        <Tabs.Tab
                            value="administrative"
                            label="Administrative åpninger"
                            icon={<CogIcon aria-hidden />}
                        />
                        <Tabs.Tab
                            value="blocked"
                            label="Blokkerte forespørsler"
                            icon={<FileTextIcon aria-hidden />}
                        />
                    </Tabs.List>

                    <Tabs.Panel value="timerestricted" className="p-6 bg-gray-50 rounded-lg">
                        <div className="mb-4">
                            <p className="text-gray-600">
                                Administrer URL-er som får midlertidig tilgang. Disse åpningene utløper automatisk etter den valgte tidsperioden.
                            </p>
                        </div>
                        <TimeRestrictedUrlEditor />
                    </Tabs.Panel>

                    <Tabs.Panel value="administrative" className="p-6 bg-gray-50 rounded-lg">
                        <form onSubmit={handleSubmit}>
                            <div className="flex flex-col gap-6">
                                <urlEditor.urlEditor />
                            </div>
                        </form>
                    </Tabs.Panel>

                    <Tabs.Panel value="blocked" className="p-6 bg-gray-50 rounded-lg">
                        <div className="mb-4">
                            <p className="text-gray-600">
                                Oversikt over URL-er som har blitt blokkert. Du kan kopiere URL-en for å legge den til i åpningslisten.
                            </p>
                        </div>

                        {!logs.data || logs.data.proxyDeniedHostPaths.length === 0 ? (
                            <Alert variant="info">
                                Ingen blokkerte forespørsler funnet. Dette kan betyr at alle dine forespørsler har blitt tillatt.
                            </Alert>
                        ) : (
                            <div className="flex flex-col gap-4">
                                <Table zebraStripes size="medium" className="bg-white rounded-lg shadow-sm">
                                    <Table.Header>
                                        <Table.Row>
                                            <Table.HeaderCell scope="col">Blokkert URL</Table.HeaderCell>
                                            <Table.HeaderCell scope="col">Tidspunkt</Table.HeaderCell>
                                            <Table.HeaderCell scope="col">Handlinger</Table.HeaderCell>
                                        </Table.Row>
                                    </Table.Header>
                                    <Table.Body>
                                        {pageData.map((url: any, i: number) => (
                                            <Table.Row key={i + url.HTTPRequest.URL.Host}>
                                                <Table.DataCell className="font-mono">
                                                    <div title={`${url.HTTPRequest.URL.Host}${url.HTTPRequest.URL.Path}`}>
                                                        {`${url.HTTPRequest.URL.Host}${url.HTTPRequest.URL.Path.length > 50 ? url.HTTPRequest.URL.Path.substring(0, 50) + '...' : url.HTTPRequest.URL.Path}`}
                                                    </div>
                                                </Table.DataCell>
                                                <Table.DataCell>
                                                    {isNaN(new Date(url.Timestamp).getTime())
                                                        ? 'Ugyldig dato'
                                                        : formatDistanceToNow(new Date(url.Timestamp), { addSuffix: true })
                                                    }
                                                </Table.DataCell>
                                                <Table.DataCell>
                                                    <CopyButton
                                                        copyText={`${url.HTTPRequest.URL.Host}${url.HTTPRequest.URL.Path}`}
                                                        text="Kopier URL"
                                                        size="small"
                                                    />
                                                </Table.DataCell>
                                            </Table.Row>
                                        ))}
                                    </Table.Body>
                                </Table>

                                {logs.data.proxyDeniedHostPaths.length > rowsPerPage && (
                                    <div className="flex justify-center">
                                        <Pagination
                                            page={page}
                                            onPageChange={setPage}
                                            count={Math.ceil(logs.data.proxyDeniedHostPaths.length / rowsPerPage)}
                                            size="small"
                                        />
                                    </div>
                                )}
                            </div>
                        )}
                    </Tabs.Panel>
                </Tabs>
            </div>
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
