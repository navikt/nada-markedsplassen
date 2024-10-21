import { Alert } from "@navikt/ds-react"
import LoaderSpinner from "../lib/spinner"
import { JoinableViewCard } from "./joinableViewCard"
import { useGetJoinableViewsForUser } from "../../lib/rest/joinableViews"

export const JoinableViewsList = () => {
    const joinableViews = useGetJoinableViewsForUser()
    return <div>
        {joinableViews.isLoading && <LoaderSpinner />}
        {joinableViews.error && <Alert variant="error">Kan ikke Hente sammenføybare viewer.</Alert>}
        {joinableViews.data &&
            <div className="flex-col space-y-2">
                {joinableViews.data.map((it:any) => <JoinableViewCard key={it.id} joinableView={it} />)}
            </div>
        }
    </div>
}