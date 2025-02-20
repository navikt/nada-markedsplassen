import {Alert, AlertProps, CopyButton, Detail, Heading, Label, Link, Table} from "@navikt/ds-react"
import React, { use } from "react";
import {translateCode, translateKind, translateParam, translateStatusCode} from "./errorTranslate";
import { backendHost } from "../header/user";
import { useRouter } from "next/router";

export interface ErrorStripeProps {
    error?: any
}

const ErrorStripe = ({ error }: ErrorStripeProps) => {
    const message = error?.message ?? "An error occurred"
    const id = error?.requestId || undefined
    const code = error?.code || undefined
    const param = error?.param || undefined
    const kind = error?.kind || undefined
    const statusCode = error?.statusCode || undefined
    const router = useRouter()

    const copyText = `Referanse: ${id}\nKode: ${code}\nParameter: ${param}\nBeskjed: ${message}`

    if(statusCode === 401) {
        return <Alert variant="info">Autentiseringsfeil. <Link className="text-blue-500" href={`${backendHost()}/api/login?redirect_uri=${encodeURIComponent(
                      router.asPath
                    )}`}>Logg inn</Link> for å fortsette</Alert>
    }

    return error &&
        <AlertWithCloseButton variant="error">
            <Heading spacing size="small" level="3">
                {translateKind(kind)}
            </Heading>
            <div className="grid gap-4">
                <div>
                    <p>Ta kontakt med oss på
                        <Link className="pl-2 pr-2" href="https://navikt.slack.com/archives/CGRMQHT50" target="_blank">
                            Slack i kanalen #nada
                        </Link>
                        hvis det fortsetter å feile, og legg ved en kopi av detaljene under.
                    </p>
                </div>
                <div>
                    <Table size="small">
                        <Table.Header>
                            <Table.Row>
                                <Table.HeaderCell scope="col">Type</Table.HeaderCell>
                                <Table.HeaderCell scope="col">Kode</Table.HeaderCell>
                                <Table.HeaderCell scope="col"></Table.HeaderCell>
                            </Table.Row>
                        </Table.Header>
                        <Table.Body>
                            <Table.Row key="referanse">
                                <Table.DataCell>Referanse</Table.DataCell>
                                <Table.DataCell>{id}</Table.DataCell>
                                <Table.DataCell>Referanse vi kan slå opp på i våre logger</Table.DataCell>
                            </Table.Row>
                            <Table.Row key="code">
                                <Table.DataCell>Kode</Table.DataCell>
                                <Table.DataCell>{code}</Table.DataCell>
                                <Table.DataCell>{translateCode(code)}</Table.DataCell>
                            </Table.Row>
                            <Table.Row key="param">
                                <Table.DataCell>Ressurs</Table.DataCell>
                                <Table.DataCell>{param}</Table.DataCell>
                                <Table.DataCell>{translateParam(param)}</Table.DataCell>
                            </Table.Row>
                            <Table.Row key="msg">
                                <Table.DataCell>Feilmelding</Table.DataCell>
                                <Table.DataCell></Table.DataCell>
                                <Table.DataCell>{message}</Table.DataCell>
                            </Table.Row>
                        </Table.Body>
                    </Table>
                </div>
                <div>
                    <CopyButton copyText={copyText} text="Kopier" />
                </div>
            </div>
        </AlertWithCloseButton>
}

export default ErrorStripe

const AlertWithCloseButton = ({children, variant}: {
    children?: React.ReactNode;
    variant: AlertProps["variant"];
}) => {
    const [show, setShow] = React.useState(true);

    return show ? (
        <Alert variant={variant} fullWidth closeButton onClose={() => setShow(false)}>
            {children || "Content"}
        </Alert>
    ) : null;
};
