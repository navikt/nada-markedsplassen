import { Box, ExpansionCard, Label } from "@navikt/ds-react"
import SearchResultLink from "../search/searchResultLink"
import { SubjectType } from "./access/newDatasetAccess"

interface Dataset {
    __typename?: 'Dataset' | undefined
    id: string
    dataproductID: string
    dataproduct: {
        name: string
        slug: string
    }
    name: string
    keywords: string[]
    slug: string
    owner: { __typename?: 'Owner' | undefined, group: string }
}

//TODO: there are dataproducts without dataset!!!!
interface DataproductsListProps {
    datasetAccesses: any[]
    isServiceAccounts?: boolean
    showAllUsersAccesses?: boolean
}

const groupDatasetAccessesByDataproduct = (datasets: any[], showAllUsersAccesses?: boolean) => {
    let dataproducts = new Map<string, Dataset[]>()

    datasets?.filter((ds) => showAllUsersAccesses === undefined || showAllUsersAccesses || ds.subject !== "group:all-users@nav.no").forEach((dataset) => {
        dataproducts.set(dataset.dataproductID, dataproducts.get(dataset.dataproductID) || [])
        dataproducts.get(dataset.dataproductID)?.push(dataset)
    })

    var datasetsGroupedByDataproduct: Array<any[]> = [];
    dataproducts.forEach((datasets) => {
        datasetsGroupedByDataproduct.push(datasets)
    })

    return datasetsGroupedByDataproduct
}

export const AccessesList = ({ datasetAccesses, isServiceAccounts, showAllUsersAccesses }: DataproductsListProps) => {
    const groupedDatasetAccesses = groupDatasetAccessesByDataproduct(datasetAccesses, showAllUsersAccesses)

    return (
        <>
            {groupedDatasetAccesses.map((datasetAccesses, i) => {
                return (
                    <div key={i}>
                    <div className="w-[60rem]">
                        <ExpansionCard key={i} aria-label="dataproduct-access">
                            <ExpansionCard.Header>
                                <ExpansionCard.Title>{`Dataprodukt - ${datasetAccesses[0].dataproductName}`}</ExpansionCard.Title>
                                { isServiceAccounts ?
                                    <ExpansionCard.Description>Klikk for å se datasettene servicebrukerne dine har tilgang til</ExpansionCard.Description> :
                                    <ExpansionCard.Description>Klikk for å se datasettene du har tilgang til</ExpansionCard.Description>
                                }
                            </ExpansionCard.Header>
                            <ExpansionCard.Content className="">
                                <>
                                    {datasetAccesses?.map((dataset, i) => {
                                        return <Box key={i} className="text-left gap-y-2 w-[47rem]">
                                            {dataset.subject !== null && dataset.subject.split(":")[0] === SubjectType.ServiceAccount ?
                                                <div className="flex gap-x-2 items-center">
                                                    <div>Servicebruker: </div>
                                                    <Label>{dataset.subject.split(":")[1]}</Label>
                                                </div> : <></>
                                            }
                                            <SearchResultLink
                                                group={{
                                                    group: dataset.group
                                                }}
                                                name={dataset.name}
                                                type={'Dataset'}
                                                link={`/dataproduct/${dataset.dataproductID}/${dataset.dpSlug}/${dataset.id}`}
                                            />
                                        </Box>
                                        
                                    })
                                    }
                                </>
                            </ExpansionCard.Content>
                        </ExpansionCard>
                    </div>
                    </div>
                )
            })}
        </>
    )
}
