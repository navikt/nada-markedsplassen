import {Alert, AlertProps, CopyButton, Heading, Link, ReadMore} from "@navikt/ds-react"
import React from "react";
import {translateCode, translateParam, translateStatusCode} from "./errorTranslate";

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

    return error &&
        <AlertWithCloseButton variant="error">
            <Heading spacing size="small" level="3">
                {translateStatusCode(statusCode)}
            </Heading>
            <p className="items-end">Det skjedde mens vi {translateCode(code)} Det gjelder {translateParam(param)}. Ta kontakt med oss på
                <Link className="pl-2 pr-2" href="https://navikt.slack.com/archives/CGRMQHT50" target="_blank">
                    Slack i kanalen #nada
                </Link>
                hvis det fortsetter å feile. {id && <div className="flex">Legg ved referanse: <span className="text-nav-red">{id}</span> <CopyButton copyText={id} size="small"></CopyButton></div>}
            </p>
            <ReadMore header="Vis den faktiske feilen">
                {message}
            </ReadMore>
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
