import { Link } from "@navikt/ds-react"
import { AccessRequestModal } from "../dataproducts/access/datasetAccess"
import { ExternalLinkIcon } from "@navikt/aksel-icons"

interface PendingAccessRequestBarProps {
    accessRequest: any
}

export const PendingAccessRequestBar = ({ accessRequest }: PendingAccessRequestBarProps) => {
    return (
        <div key={accessRequest.id} className="w-[60rem] mb-5 mt-5 border pt-2 pb-2 pl-4 pr-4 flex flex-row justify-between rounded border-gray-200">
            <div>
                <h3> <Link rel="norefferer" href={`/dataproduct/${accessRequest.dataproductID}/${accessRequest.dataproductSlug}/${accessRequest.datasetID}`}>
                    {`${accessRequest?.datasetName} - ${accessRequest?.dataproductName}`}
                </Link>
                </h3>
                {accessRequest.owner}
                <br></br>
                <div className="flex flex-row">
                    <div>
                        {!accessRequest.expires ? "Alltid tilgang fra " : "Tilgangsperiode: "}
                        {new Date(accessRequest.created).toLocaleDateString('no-NO')}
                        {accessRequest.expires && ` - ${new Date(accessRequest.expires).toLocaleDateString('no-NO')}`}
                    </div>
                    <div className="ml-[2rem]">
                        {accessRequest.polly?.url ? (
                            <Link target="_blank" rel="norefferer" href={accessRequest.polly.url}>
                                Åpne behandling
                                <ExternalLinkIcon title="a11y-title" fontSize="1.5rem" />
                            </Link>
                        ) : (
                            'Ingen behandling'
                        )}
                    </div>
                </div>
            </div>
            <div>
                <AccessRequestModal requestID={accessRequest.id}></AccessRequestModal>
            </div>
        </div>)
}