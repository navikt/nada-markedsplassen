import { Alert, Box, CopyButton, ExpansionCard, Link, Loader, Tooltip } from "@navikt/ds-react"
import { ExternalLink } from "@navikt/ds-icons"
import LoaderSpinner from "../lib/spinner"
import { useState } from "react"
import { useGetJoinableView } from "../../lib/rest/joinableViews"

interface JoinableViewCardProps {
    joinableView: any
}

export const JoinableViewCardContent = ({ joinableViewId }: { joinableViewId: string }) => {
    const { data, isLoading: loading, error } = useGetJoinableView(joinableViewId)
    const urlComps = data?.pseudoDatasources && data?.pseudoDatasources.length
        ? data?.pseudoDatasources[0].bigqueryUrl.split('.') : ["", "", ""]
    const projectID = urlComps[0]
    const datasetID = urlComps[1]
    const bigQueryUrl = `https://console.cloud.google.com/bigquery?d=${datasetID}&p=${projectID}&page=dataset`

    return <div>
        {loading && <LoaderSpinner />}
        {error && <Alert variant="error">Klarte ikke hente data om views tilrettelagt for kobling</Alert>}
        {data && <>
            <Link href={bigQueryUrl}>{"Åpne BigQuery dataset i Google Cloud Console"}<ExternalLink /></Link>
            {data?.pseudoDatasources.map((bqv:any, index: number) => <Box key={index} padding="1" className="w-[55rem]">
                {bqv.deleted?<Tooltip content="Datasettet er slettet fra markedsplassen"><div className="flex flex-row items-center line-through">{bqv.bigqueryUrl}</div></Tooltip>:
                bqv.accessible? <div className="flex flex-row items-center bg-gray-200">{bqv.bigqueryUrl}<CopyButton copyText={bqv.bigqueryUrl}></CopyButton></div>
                    : <Tooltip content="Har ikke tilgang til datasettet"><div className="flex flex-row items-center text-gray-200" >{bqv.bigqueryUrl}</div></Tooltip>}
            </Box>)}
            {data?.expires &&
                <div className="mt-3 italic">
                    BigQuery datasettet slettes {data?.expires.split("T")[0]}
                </div>
            }
        </>}
    </div>
}

export const JoinableViewCard = ({ joinableView }: JoinableViewCardProps) => {
    const [expanded, setExpanded] = useState(false)

    return <div key={joinableView.id} className="w-[60rem]">
        <ExpansionCard aria-label="default-demo" onToggle={open => setExpanded(open)}>
            <ExpansionCard.Header>
                <ExpansionCard.Title>{`${joinableView?.name} - ${joinableView?.created}`}</ExpansionCard.Title>
                <ExpansionCard.Description>
                    <p>Klikk for å vise BigQuery dataset</p></ExpansionCard.Description>
            </ExpansionCard.Header>
            <ExpansionCard.Content>
                {expanded && <JoinableViewCardContent joinableViewId={joinableView.id} />}
            </ExpansionCard.Content>
        </ExpansionCard>
    </div>
}