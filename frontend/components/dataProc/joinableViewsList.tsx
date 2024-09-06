import { Alert, Heading, Link, ExpansionCard, CopyButton, Box } from "@navikt/ds-react"
import { ExternalLink } from "@navikt/ds-icons"
import LoaderSpinner from "../lib/spinner"
import { JoinableViewCard } from "./joinableViewCard"
import { useGetJoinableView, useGetJoinableViewsForUser } from "../../lib/rest/joinableViews"

export const JoinableViewsList = () => {
    const joinableViews = useGetJoinableViewsForUser()
    return <div>
        {joinableViews.loading && <LoaderSpinner />}
        {joinableViews.error && <Alert variant="error">Kan ikke Hente sammenføybare viewer.</Alert>}
        {joinableViews.data &&
            <div className="flex-col space-y-2">
                {joinableViews.data.map((it:any) => <JoinableViewCard key={it.id} joinableView={it} />)}
            </div>
        }
    </div>
}