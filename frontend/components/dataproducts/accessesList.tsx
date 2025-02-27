import { Alert, Box, Button, ExpansionCard, Label } from "@navikt/ds-react"
import SearchResultLink from "../search/searchResultLink"
import { SubjectType } from "../../lib/rest/access"
import { AccessModal } from "./access/datasetAccess"
import { revokeDatasetAccess } from '../../lib/rest/access'
import { useState } from "react"
import { Access } from '../../lib/rest/generatedDto'

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
    isRevokable?: boolean
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

export const AccessesList = ({ datasetAccesses, isServiceAccounts, showAllUsersAccesses, isRevokable }: DataproductsListProps) => {
    const [formError, setFormError] = useState('')
    const groupedDatasetAccesses = groupDatasetAccessesByDataproduct(datasetAccesses, showAllUsersAccesses)

    const removeAccess = async (accessID: string, setOpen: Function, setRemovingAccess: Function) => {
        setRemovingAccess(true)
        try {
           await revokeDatasetAccess(accessID)
           window.location.reload()
        } catch (e: any) {
           setFormError(e.message)
        } finally {
           setOpen(false)
        }
    }

    return (
        <>
            {formError && <Alert variant={'error'}>{formError}</Alert>}
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
                                    {datasetAccesses?.map((datasetAccess, i) => {
                                        return <Box key={i} className="text-left gap-y-2 gap-x-5 w-[55rem]">
                                                {datasetAccess.subject !== null && datasetAccess.subject.split(":")[0] === SubjectType.ServiceAccount ?
                                                    <div className="flex gap-x-2 items-center">
                                                        <div>Servicebruker: </div>
                                                        <Label>{datasetAccess.subject.split(":")[1]}</Label>
                                                    </div> : <></>
                                                }
                                                <div className="flex gap-x-2 items-center justify-center">
                                                    <SearchResultLink
                                                        group={{
                                                            group: datasetAccess.group
                                                        }}
                                                        name={datasetAccess.name}
                                                        type={'Dataset'}
                                                        link={`/dataproduct/${datasetAccess.dataproductID}/${datasetAccess.dpSlug}/${datasetAccess.id}`}
                                                    />
                                                    {isRevokable && (
                                                        <div className="h-[2rem]">
                                                            <AccessModal accessID={datasetAccess.accessID} subject={datasetAccess.subject} datasetName={datasetAccess.name} action={removeAccess} />
                                                        </div>
                                                    )}
                                                </div>
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
