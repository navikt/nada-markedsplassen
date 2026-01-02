import { BodyShort, Heading, HGrid, Tabs, VStack } from "@navikt/ds-react";
import Head from "next/head";
import { JoinableViewsList } from "../../dataProc/joinableViewsList";
import { revokeDatasetAccess, revokeRestrictedMetabaseAccess, useFetchAllUsersAccesses, useFetchUserAccesses } from "../../../lib/rest/access";
import { DataproductWithDataset, UserAccessDataproduct, UserAccessDatasets } from "../../../lib/rest/generatedDto";
import LoaderSpinner from "../../lib/spinner";
import { AccessModal } from "./datasetAccess";
import { useEffect, useState } from "react";
import { DataproductExpansionCard } from "../DataproductExpansionCard";
import { DatasetLinkCard } from "../DatasetLinkCard";
import { HttpError } from "../../../lib/rest/request";
import ErrorStripe from "../../lib/errorStripe";

interface Props {
  ownedDataproducts: DataproductWithDataset[]
  defaultView?: string
}

function UserAccessesPage({ ownedDataproducts, defaultView }: Props) {
  const { data: accesses, error: userAccessesError, isLoading: userAccessesIsLoading } = useFetchUserAccesses()
  const { data: allUsersAccesses, error: allUsersAccessesError, isLoading: allUsersAccessesIsLoading } = useFetchAllUsersAccesses()
  const [personalAccesses, setPersonalAccesses] = useState(accesses?.personal || [])
  const [serviceAccountAccesses, setServiceAccountAccesses] = useState(accesses?.serviceAccountGranted)
  const onPersonalAccessRevoked = (dataset: UserAccessDatasets) => {
    setPersonalAccesses(prev =>
      prev.map(dp => ({
        ...dp,
        datasets: dp.datasets.filter(ds => ds.datasetID !== dataset.datasetID)
      }))
        .filter(dp => dp.datasets.length > 0)
    )
  }
  const onServiceAccountAccessRevoked = (serviceAccount: string, dataset: UserAccessDatasets) => {
    setServiceAccountAccesses(prev => {
      if (!prev) return prev

      const updatedDataproducts = prev[serviceAccount]
        .map(dp => ({
          ...dp,
          datasets: dp.datasets.filter(ds => ds.datasetID !== dataset.datasetID)
        }))
        .filter(dp => dp.datasets.length > 0)

      if (updatedDataproducts.length === 0) {
        const { [serviceAccount]: _, ...rest } = prev
        return rest
      }

      return {
        ...prev,
        [serviceAccount]: updatedDataproducts
      }
    })
  }

  useEffect(() => {
    setPersonalAccesses(accesses?.personal || [])
    setServiceAccountAccesses(accesses?.serviceAccountGranted)
  }, [accesses])

  return (
    <div className="grid gap-4">
      <Head>
        <title>Mine tilganger</title>
      </Head>
      <h2>Mine tilganger</h2>
      <Tabs
        defaultValue={defaultView ?? "owner"}>
        <Tabs.List>
          <Tabs.Tab
            value="owner"
            label="Eier"
          />
          <Tabs.Tab
            value="granted"
            label="Innvilgede tilganger"
          />
          <Tabs.Tab
            value="serviceAccountGranted"
            label="Tilganger servicebrukere"
          />
          <Tabs.Tab
            value="allUsers"
            label="Åpne datasett"
          />
          <Tabs.Tab
            value="joinable"
            label="Views tilrettelagt for kobling"
          />
        </Tabs.List>
        <Tabs.Panel value="owner" className="w-full space-y-2 p-4">
          <VStack padding="space-16" gap="space-16" className="max-w-5xl">
            <OwnedDataproducts dataproducts={ownedDataproducts} />
          </VStack>
        </Tabs.Panel>
        <Tabs.Panel value="granted" className="w-full space-y-2 p-4">
          <VStack padding="space-16" gap="space-16" className="max-w-5xl">
            <Accesses
              accesses={personalAccesses}
              isLoading={userAccessesIsLoading}
              isRevokable
              onRevoke={onPersonalAccessRevoked}
              error={userAccessesError}
              level="3"
            />
          </VStack>
        </Tabs.Panel>
        <Tabs.Panel value="serviceAccountGranted" className="w-full space-y-2 p-4">
          {serviceAccountAccesses && Object.entries(serviceAccountAccesses).map(([sa, access]) => {
            return (
              <VStack padding="space-16" gap="space-16" className="max-w-5xl" key={sa}>
                <Heading level="3" size="medium" className="border-b">{sa.split(":")[1]}:</Heading>
                <Accesses
                  accesses={access}
                  isLoading={userAccessesIsLoading}
                  isRevokable
                  onRevoke={(ds: UserAccessDatasets) => onServiceAccountAccessRevoked(sa, ds)}
                  error={userAccessesError}
                  level="4"
                />
              </VStack>
            )
          })}
        </Tabs.Panel>
        <Tabs.Panel value="allUsers" className="w-full space-y-2 p-4">
          <VStack padding="space-16" gap="space-16" className="max-w-5xl">
            <Accesses
              accesses={allUsersAccesses}
              isLoading={allUsersAccessesIsLoading}
              error={allUsersAccessesError}
              isRevokable
              onRevoke={onPersonalAccessRevoked}
              level="3"
            />
          </VStack>
        </Tabs.Panel>
        <Tabs.Panel value="joinable" className="w-full p-4">
          <JoinableViewsList />
        </Tabs.Panel>
      </Tabs>
    </div>
  )
}


interface OwnedDataproductsProps {
  dataproducts: DataproductWithDataset[]
  isLoading?: boolean
}
function OwnedDataproducts({ isLoading, dataproducts }: OwnedDataproductsProps) {
  if (isLoading) return <LoaderSpinner />

  return (
    <>
      {dataproducts.map((dp) =>
      (
        <DataproductExpansionCard
          key={dp.id}
          name={dp.name}
          description={dp.description || ""}
          level="3"
        >
          <HGrid gap="space-16" columns={{ md: 1, lg: 2 }}>
            {dp.datasets
              .toSorted((a, b) => a.name.localeCompare(b.name))
              .map((ds) => {
                return (
                  <DatasetLinkCard
                    key={ds.id}
                    href={`/dataproduct/${dp.id}/${dp.slug}/${ds.id}`}
                    name={ds.name}
                    description={ds.description || ""}
                  />
                )
              })
            }
          </HGrid>
        </DataproductExpansionCard>
      ))}
    </>
  )
}


interface AccessesProps {
  isRevokable?: boolean
  isLoading?: boolean
  accesses?: UserAccessDataproduct[]
  error: HttpError | null
  level: "1" | "2" | "3" | "4"
  onRevoke?: (dataset: UserAccessDatasets) => void
}
function Accesses({ isRevokable, onRevoke, isLoading, accesses, error, level }: AccessesProps) {
  if (error) {
    return <ErrorStripe error={error} />
  }
  if (isLoading || !accesses) return <LoaderSpinner />

  const removeAccess = async (dataset: UserAccessDatasets) => {
    for (const a of dataset.accesses) {
      if (a.platform === 'bigquery') {
        await revokeDatasetAccess(a.id)
      } else if (a.platform === 'metabase') {
        await revokeRestrictedMetabaseAccess(a.id)
      }
    }
    if (onRevoke) {
      onRevoke(dataset)
    }
  }

  return (
    <>
      {accesses
        .toSorted((a, b) => a.dataproductName.localeCompare(b.dataproductName))
        .map((dp: UserAccessDataproduct) => (
          <DataproductExpansionCard
            key={dp.dataproductID}
            name={dp.dataproductName}
            description={dp.dataproductDescription}
            level={level}
          >
            <HGrid gap="space-16" columns={{ md: 1, lg: 2 }}>
              {dp.datasets
                .toSorted((a: UserAccessDatasets, b: UserAccessDatasets) => a.datasetName.localeCompare(b.datasetName))
                .map((ds: UserAccessDatasets) => {
                  return (
                    <DatasetLinkCard
                      key={ds.datasetID}
                      href={`/dataproduct/${dp.dataproductID}/${dp.dataproductSlug}/${ds.datasetID}`}
                      name={ds.datasetName}
                      description={ds.datasetDescription}
                    >
                      <BodyShort>Plattformer: {ds.accesses.filter(a => !a.revoked).map(a => a.platform).join(", ")}</BodyShort>
                      {isRevokable && (
                        <div className="ml-auto" onClick={(e) => { e.preventDefault() }}>
                          <AccessModal subject={ds.accesses[0].subject} datasetName={ds.datasetName} action={async () => removeAccess(ds)} />
                        </div>
                      )}
                    </DatasetLinkCard>
                  )
                })
              }
            </HGrid>
          </DataproductExpansionCard>
        ))}
    </>
  )
}

export default UserAccessesPage
