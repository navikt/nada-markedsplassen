import { BodyShort, ExpansionCard, Heading, HGrid, LinkCard, Tabs, VStack } from "@navikt/ds-react";
import Head from "next/head";
import { JoinableViewsList } from "../../dataProc/joinableViewsList";
import Link from "next/link";
import { revokeDatasetAccess, revokeRestrictedMetabaseAccess, useFetchUserAccesses } from "../../../lib/rest/access";
import { UserAccessDataproduct, UserAccessDatasets } from "../../../lib/rest/generatedDto";
import LoaderSpinner from "../../lib/spinner";
import { AccessModal } from "./datasetAccess";
import { useEffect, useState } from "react";

interface Props {
  defaultView?: string
}

function UserAccessesPage({ defaultView }: Props) {
  const { data: accesses, error, isLoading } = useFetchUserAccesses()
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
        defaultValue={defaultView ?? "granted"}>
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
            value="joinable"
            label="Views tilrettelagt for kobling"
          />
        </Tabs.List>
        {/*
        <Tabs.Panel value="owner" className="w-full space-y-2 p-4">
          <AccessesList datasetAccesses={data.accessable.owned} />
        </Tabs.Panel>
          <AccessesList datasetAccesses={data.accessable.granted}
            showAllUsersAccesses={showAllUsersAccesses}
            isRevokable={true}
          />
          <AccessesList datasetAccesses={data.accessable.serviceAccountGranted}
            isServiceAccounts={true}
            isRevokable={true}
          />
        */}
        <Tabs.Panel value="granted" className="w-full space-y-2 p-4">
          <VStack padding="space-16" gap="space-16">
            <Accesses
              accesses={personalAccesses}
              isLoading={isLoading}
              isRevokable
              onRevoke={onPersonalAccessRevoked}
            />
          </VStack>
        </Tabs.Panel>
        <Tabs.Panel value="serviceAccountGranted" className="w-full space-y-2 p-4">
          {serviceAccountAccesses && Object.entries(serviceAccountAccesses).map(([sa, accs]) => {
            return (
              <VStack padding="space-16" gap="space-16">
                <Heading level="3" size="medium" className="border-b">{sa.split(":")[1]}:</Heading>
                <Accesses
                  accesses={accs}
                  isLoading={isLoading}
                  isRevokable
                  onRevoke={(ds: UserAccessDatasets) => onServiceAccountAccessRevoked(sa, ds)}
                />
              </VStack>
            )
          })}
        </Tabs.Panel>
        <Tabs.Panel value="joinable" className="w-full p-4">
          <JoinableViewsList />
        </Tabs.Panel>
      </Tabs>
    </div>
  )
}


interface AccessesProps {
  isRevokable?: boolean
  isLoading?: boolean
  accesses?: UserAccessDataproduct[]
  onRevoke?: (dataset: UserAccessDatasets) => void
}
function Accesses({ isRevokable, onRevoke, isLoading, accesses }: AccessesProps) {
  if (isLoading) return <LoaderSpinner />
  if (!accesses) return <LoaderSpinner />

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
  console.log("accesses: ", accesses)

  return (
    <>
      {accesses.map((dp: UserAccessDataproduct) => {
        return (<>
          <ExpansionCard aria-label={`Dataprodukt: ${dp.dataproductName}`} className="max-w-5xl">
            <ExpansionCard.Header>
              <ExpansionCard.Title>
                {dp.dataproductName}
              </ExpansionCard.Title>
              <ExpansionCard.Description>
                {dp.dataproductDescription}
              </ExpansionCard.Description>
            </ExpansionCard.Header>
            <ExpansionCard.Content>
              <HGrid gap="space-16" columns={{ md: 1, lg: 2 }}>
                {dp.datasets
                  .toSorted((a: UserAccessDatasets, b: UserAccessDatasets) => a.datasetName.localeCompare(b.datasetName))
                  .map((ds: UserAccessDatasets) => {
                    return (
                      <LinkCard className="mb-4">
                        <LinkCard.Title>
                          <LinkCard.Anchor asChild>
                            <Link href={`/dataproduct/${dp.dataproductID}/${dp.dataproductSlug}/${ds.datasetID}`}>
                              {ds.datasetName}
                            </Link>
                          </LinkCard.Anchor>
                        </LinkCard.Title>
                        <LinkCard.Description className="line-clamp-3">
                          {ds.datasetDescription}
                        </LinkCard.Description>
                        <LinkCard.Footer>
                          <BodyShort>Plattformer: {ds.accesses.filter(a => !a.revoked).map(a => a.platform).join(", ")}</BodyShort>
                          {isRevokable && (
                            <div className="ml-auto" onClick={(e) => { e.preventDefault() }}>
                              <AccessModal subject={ds.accesses[0].subject} datasetName={ds.datasetName} action={async () => removeAccess(ds)} />
                            </div>
                          )}
                        </LinkCard.Footer>
                      </LinkCard>
                    )
                  })
                }
              </HGrid>
            </ExpansionCard.Content>
          </ExpansionCard>
        </>
        )
      })}
    </>
  )
}

export default UserAccessesPage
