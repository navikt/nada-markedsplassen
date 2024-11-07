import { Alert, CopyButton, Link } from "@navikt/ds-react"

export interface ErrorStripeProps {
    error?: any
}

const ErrorStripe = ({ error }: ErrorStripeProps) => {
    const message = error?.message ?? "An error occurred"
    const id = error?.requestID || undefined
    const code = error?.code || undefined
    const param = error?.param || undefined

    return error &&
    <Alert variant="error" className="w-full">
        <div className="pl-2">{message}</div>
        <p className="items-end">Du kan prøve igjen, eller ta kontakt på
            <Link className="pl-2 pr-2" href="https://navikt.slack.com/archives/CGRMQHT50" target="_blank">Slack i kanalen #nada</Link>
            eller send en e-post til
            <Link className="pl-2 pr-2" href="mailto:nada@nav.no" target="_blank">nada@nav.no</Link>
            hvis det fortsetter å feil.
        </p>
        {id && <div className="flex">
            Vær vennlig å legg ved <strong>request_id</strong>: <p className="flex pl-2 pr-2 text-nav-red">{id}<CopyButton copyText={id} size="small"></CopyButton></p> i beskjeden din når du tar kontakt.
        </div>}
    </Alert>
}

export default ErrorStripe
