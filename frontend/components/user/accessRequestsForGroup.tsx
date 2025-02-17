import { Heading, Link, Table, Tabs } from "@navikt/ds-react"
import { PendingAccessRequestBar } from "./pendingAccessRequestBar"

interface pendingAccessRequestsForGroupProps {
    accessRequests: any[]
}

export const AccessRequestsForGroup = ({ accessRequests }: pendingAccessRequestsForGroupProps) => {
    //TODO: support approved and denied requests
    const pendingRequest = accessRequests?.filter((r) => r.status === 'pending')
    return <><div>
        {!!pendingRequest?.length ? (<div>
            {pendingRequest.map((r: any, i: number) => (
                <PendingAccessRequestBar accessRequest={r} key={i}></PendingAccessRequestBar>
            ))
            }</div>
        ) : (
            <div>Ingen tilgangssøknader</div>
        )}
    </div>
    </>
}