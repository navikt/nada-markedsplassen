import { Accordion, BodyShort, Box, ExpansionCard, Heading, LinkCard, List, Tabs, Tag } from "@navikt/ds-react";
import Head from "next/head";
import { JoinableViewsList } from "../../dataProc/joinableViewsList";
import Link from "next/link";
import { useFetchUserAccesses } from "../../../lib/rest/access";
import { UserAccessDataproduct, UserAccessDatasets } from "../../../lib/rest/generatedDto";
import LoaderSpinner from "../../lib/spinner";
import { BandageIcon } from "@navikt/aksel-icons";

interface Props {
  defaultView?: string
}

function UserAccessesPage({ defaultView }: Props) {
  const { data: grantedAccesses, error: grantedAccessesError, isLoading } = useFetchUserAccesses()

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
          <Accesses
            accesses={grantedAccesses?.granted}
            isLoading={isLoading}
            isRevokable
          />
        </Tabs.Panel>
        <Tabs.Panel value="serviceAccountGranted" className="w-full space-y-2 p-4">
          <Accesses />
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
}
function Accesses({ isRevokable, isLoading, accesses }: AccessesProps) {
  if (isLoading) return <LoaderSpinner />
  if (!accesses) return <LoaderSpinner />
  return (
    <>
      {accesses.map((dp: UserAccessDataproduct) => {
        return (<>
          <ExpansionCard aria-label="Demo med description">
            <ExpansionCard.Header>
              <ExpansionCard.Title>
                {dp.dataproductName}
              </ExpansionCard.Title>
              <ExpansionCard.Description>
                {dp.dataproductDescription}
              </ExpansionCard.Description>
            </ExpansionCard.Header>
            <ExpansionCard.Content>
              {
                dp.datasets.map((ds: UserAccessDatasets) => {
                  return (
                    <LinkCard className="mb-4">
                      <LinkCard.Title>
                        <LinkCard.Anchor asChild>
                          <Link
                            href={`/dataproduct/${dp.dataproductID}/${dp.dataproductSlug}/${ds.datasetID}`}
                          >
                            {ds.datasetName}
                          </Link>
                        </LinkCard.Anchor>
                      </LinkCard.Title>
                      <LinkCard.Description>
                        {ds.datasetDescription}
                      </LinkCard.Description>
                      <LinkCard.Footer>
                        <Tag size="small" variant="neutral-filled">
                          Tag 1
                        </Tag>
                        <Tag size="small" variant="neutral">
                          Tag 2
                        </Tag>
                      </LinkCard.Footer>
                    </LinkCard>
                  )
                })
              }
            </ExpansionCard.Content>
          </ExpansionCard>
        </>
        )
      })}
    </>
  )
}

export default UserAccessesPage
