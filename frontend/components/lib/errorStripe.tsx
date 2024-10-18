import { Alert, CopyButton, Link } from "@navikt/ds-react"
import { HttpError } from "../../lib/rest/request"
import { SlackLogo, SlackLogoMini } from "./icons/slackLogo"
import { EmailIcon, EmailIconMini } from "./icons/emailIcon"

export interface ErrorStripeProps {
    error?: any
}

const ErrorStripe = ({ error }: ErrorStripeProps) => {
    const defaultMessage = new Map([
        [404, "Not found"],
        [400, "Application error"],
        [403, "Permission denied"],
        [401, "Unauthorized"],
        [409, "Applicaiton error"],
        [500, "Server error"],
    ])

    const message = error?.message || error && error.status && defaultMessage.get(error.status) || "An error occurred"
    const status = error?.status || 0
    const id = error?.id || undefined

    return error && <Alert variant="error" className="w-full">
        <div className="flex-row">
            <div className="flex">
                {status && <div>{status}: </div>}
                <div className="pl-2">{message}</div>
            </div>
            <div className="flex w-full items-end">
                <p className="items-end">You can try again, or contact
                <Link className="pl-2 pr-2" href="https://slack.com/app_redirect?channel=nada" target="_blank"><SlackLogoMini/>#nada</Link>
                or
                <Link className ="pl-2 pr-2" href ="mailto:nada@nav.no" target="_blank"><EmailIconMini />nada@nav.no</Link>
                for support if the error persists.
                </p>
            </div>
            {id && <div className="flex">
                Please kindly attach Error ID <p className="flex pl-2 pr-2 text-nav-red">{id}<CopyButton copyText={id} size="small"></CopyButton></p> in your message when contact support.
            </div>}
        </div>
    </Alert>
}

export default ErrorStripe