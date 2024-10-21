import { Alert, CopyButton, Link } from "@navikt/ds-react"

export interface ErrorStripeProps {
    error?: any
}

const ErrorStripe = ({ error }: ErrorStripeProps) => {
    const message = error?.message ?? "An error occurred"
    const id = error?.id || undefined

    return error && <Alert variant="error" className="w-full">
        <div className="pl-2">{message}</div>
        <p className="items-end">You can try again, or contact
            <Link className="pl-2 pr-2" href="https://navikt.slack.com/archives/CGRMQHT50" target="_blank">#nada on Slack</Link>
            or
            <Link className="pl-2 pr-2" href="mailto:nada@nav.no" target="_blank">nada@nav.no</Link>
            for support if the error persists.
        </p>
        {id && <div className="flex">
            Please kindly attach Error ID <p className="flex pl-2 pr-2 text-nav-red">{id}<CopyButton copyText={id} size="small"></CopyButton></p> in your message when contact support.
        </div>}
    </Alert>
}

export default ErrorStripe