import { Heading, Tabs } from '@navikt/ds-react'
import { useRouter } from 'next/router'
import React, { useContext } from 'react'
import { UserState } from '../../lib/context'
import { Dataproduct, InsightProduct, PublicMetabaseDashboardOutput, Story } from '../../lib/rest/generatedDto'
import { deleteMetabaseDashboard } from '../../lib/rest/metabaseDashboards'
import { useGetProductAreas } from '../../lib/rest/productAreas'
import { SearchResult } from '../../lib/rest/search'
import { deleteStory } from '../../lib/rest/stories'
import { useSearchTeamKatalogen } from '../../lib/rest/teamkatalogen'
import { SearchParam } from '../../pages/search'
import ErrorStripe from "../lib/errorStripe"
import LoaderSpinner from '../lib/spinner'
import SearchResultLink from './searchResultLink'

const Results = ({ children }: { children: React.ReactNode }) => (
  <div className="results">{children}</div>
)

type ResultListInterface = {
  search?: {data: SearchResult|undefined, isLoading: boolean, error: any}
  dataproducts?: Dataproduct[]
  stories?: Story[]
  insightProducts?: InsightProduct[]
	publicMetabaseDashboards?: PublicMetabaseDashboardOutput[]
  searchParam?: SearchParam
  updateQuery?: (updatedParam: SearchParam) => void
}

const ResultList = ({
  search,
  dataproducts,
  stories,
  insightProducts,
	publicMetabaseDashboards,
  searchParam,
  updateQuery,
}: ResultListInterface) => {
  const userInfo = useContext(UserState)
  const { data: teamkatalogen } = useSearchTeamKatalogen()
  const { data: productAreas } = useGetProductAreas()
  const router = useRouter()

  const isDataProduct = (item: any) => !!item.datasets

  const getTeamKatalogenInfo = (teamkatalogenURL: any) => {
    const getTeamID = (url: string)  => {
      var urlComponents = url?.split("/")
      return urlComponents?.[urlComponents.length - 1]
    }
    const tk = teamkatalogen?.find((it) => getTeamID(it.url) == getTeamID(teamkatalogenURL))
    const po = productAreas?.find((it) => it.id == tk?.productAreaID)

    return {
      productArea: po?.name,
      teamkatalogenTeam: tk?.name
    }
  }

  const groupByTeam = <T,>(
    items: T[],
    getGroup: (item: T) => string | undefined | null,
    getTeamkatalogenURL: (item: T) => string | undefined | null
  ) => {
    const grouped = items.reduce((acc, item) => {
      const key = getGroup(item) ?? 'Ukjent'
      if (!acc[key]) acc[key] = []
      acc[key].push(item)
      return acc
    }, {} as Record<string, T[]>)

    const sortedGroups = Object.keys(grouped).sort((a, b) => {
      const nameA = getTeamKatalogenInfo(getTeamkatalogenURL(grouped[a][0])).teamkatalogenTeam ?? a
      const nameB = getTeamKatalogenInfo(getTeamkatalogenURL(grouped[b][0])).teamkatalogenTeam ?? b
      return nameA.localeCompare(nameB)
    })

    return { grouped, sortedGroups }
  }

  const TeamGroupHeading = ({ teamkatalogenURL, groupKey }: { teamkatalogenURL?: string | null, groupKey: string }) => (
    <Heading size="small" level="3" className="mt-4">
      {getTeamKatalogenInfo(teamkatalogenURL).teamkatalogenTeam ?? groupKey}
    </Heading>
  )

  if (search && !!searchParam) {
    var { data, isLoading: loading, error } = search

    if (error) return <ErrorStripe error={error} />
    if (loading || !data) return <LoaderSpinner />
    const dataproducts = data.results?.filter(
      (d) => isDataProduct(d.result)
    )
    const datastories = data.results?.filter(
      (d) => !isDataProduct(d.result)
    )

    return (
      <Results>
        <Tabs
          defaultValue={searchParam?.preferredType}
          size="medium"
          value={searchParam?.preferredType}
          onChange={(focused) => {
            updateQuery?.({ ...searchParam, preferredType: focused })
          }}
        >
          <Tabs.List>
            <Tabs.Tab
              value="story"
              label={`Fortellinger (${datastories?.length || 0})`}
            />
            <Tabs.Tab
              value="dataproduct"
              label={`Produkter (${dataproducts?.length || 0})`}
            />
          </Tabs.List>
          <Tabs.Panel className="flex flex-col pt-4 gap-4" value="story">
            {datastories?.map(
              (it, idx) =>
              (
                !isDataProduct(it.result) && (
                  <SearchResultLink
                    key={idx}
                    name={it.result.name}
                    type={'story'}
                    keywords={it.result.keywords}
                    description={it.excerpt}
                    lastModified={it.result.lastModified}
                    link={`/story/${it.result.id}`}
                    group={{
                      group: it.result.group,
                      teamkatalogenURL: it.result?.teamkatalogenURL,
                    }}
                    {...getTeamKatalogenInfo(it.result?.teamkatalogenURL)}
                  />
                )
              )

            )}
          </Tabs.Panel>
          <Tabs.Panel className="flex flex-col gap-4" value="dataproduct">
            {dataproducts?.map(
              (d, idx) =>
                isDataProduct(d.result) && (
                  <SearchResultLink
                    key={idx}
                    group={d.result.owner}
                    name={d.result.name}
                    keywords={d.result.keywords}
                    description={d.result.description}
                    link={`/dataproduct/${d.result.id}/${d.result.slug}`}
                    datasets={d.result.datasets}
                    {...getTeamKatalogenInfo(d.result.owner?.teamkatalogenURL)}
                  />
                )
            )}
          </Tabs.Panel>
        </Tabs>
        {data.results?.length == 0 && 'ingen resultater'}
      </Results>
    )
  }
  if (dataproducts) {
    const { grouped, sortedGroups } = groupByTeam(
      dataproducts,
      d => d.owner?.group,
      d => d.owner?.teamkatalogenURL
    )

    return (
      <Results>
        {sortedGroups.map(group => (
          <div key={group} className="flex flex-col gap-2">
            <TeamGroupHeading teamkatalogenURL={grouped[group][0].owner?.teamkatalogenURL} groupKey={group} />
            {grouped[group].map((d, idx) => (
              <SearchResultLink
                key={idx}
                group={d.owner}
                name={d.name}
                keywords={d.keywords}
                link={`/dataproduct/${d.id}/${d.slug}`}
                {...getTeamKatalogenInfo(d.owner?.teamkatalogenURL)}
              />
            ))}
          </div>
        ))}
      </Results>
    )
  }

  if (stories) {
    const { grouped, sortedGroups } = groupByTeam(
      stories,
      s => s.group,
      s => s.teamkatalogenURL
    )

    return (
      <Results>
        {sortedGroups.map(group => (
          <div key={group} className="flex flex-col gap-2">
            <TeamGroupHeading teamkatalogenURL={grouped[group][0].teamkatalogenURL} groupKey={group} />
            {grouped[group].map((s, idx) => (
              <SearchResultLink
                key={idx}
                group={{ group: s.group, teamkatalogenURL: s.teamkatalogenURL }}
                id={s.id}
                name={s.name}
                resourceType={"datafortelling"}
                link={`/story/${s.id}`}
                {...getTeamKatalogenInfo(s.teamkatalogenURL)}
                keywords={s.keywords}
                lastModified={s.lastModified}
                editable={true}
                description={s.description}
                deleteResource={() => deleteStory(s.id).then(() => router.reload())}
              />
            ))}
          </div>
        ))}
      </Results>
)
  }

  if (insightProducts) {
    const { grouped, sortedGroups } = groupByTeam(
      insightProducts,
      p => p.group,
      p => p.teamkatalogenURL
    )

    return (
      <Results>
        {sortedGroups.map(group => (
          <div key={group} className="flex flex-col gap-2">
            <TeamGroupHeading teamkatalogenURL={grouped[group][0].teamkatalogenURL} groupKey={group} />
            {grouped[group].map((p, idx) => (
              <SearchResultLink
                key={idx}
                group={{ group: p.group, teamkatalogenURL: p.teamkatalogenURL }}
                resourceType={"innsiktsprodukt"}
                id={p.id}
                name={p.name}
                link={p.link}
                {...getTeamKatalogenInfo(p.teamkatalogenURL)}
                description={p.description}
                innsiktsproduktType={p.type}
                editable={!!userInfo?.googleGroups?.find((it: any) => it.email == p.group)}
              />
            ))}
          </div>
        ))}
      </Results>
    )
  }

  if (publicMetabaseDashboards) {
    const { grouped, sortedGroups } = groupByTeam(
      publicMetabaseDashboards,
      d => d.group,
      d => d.teamkatalogenURL
    )

    return (
      <Results>
        {sortedGroups.map(group => (
          <div key={group} className="flex flex-col gap-2">
            <TeamGroupHeading teamkatalogenURL={grouped[group][0].teamkatalogenURL} groupKey={group} />
            {grouped[group].map((dashboard, idx) => (
              <SearchResultLink
                key={idx}
                group={{ group: dashboard.group, teamkatalogenURL: dashboard.teamkatalogenURL }}
                resourceType={"metabase-dashboard"}
                id={dashboard.id}
                name={dashboard.name}
                link={dashboard.link}
                externalLink
                {...getTeamKatalogenInfo(dashboard.teamkatalogenURL)}
                description={dashboard.description}
                editable={!!userInfo?.googleGroups?.find((it: any) => it.email == dashboard.group)}
                deleteResource={() => deleteMetabaseDashboard(dashboard.id).then(() => router.reload())}
              />
            ))}
          </div>
        ))}
      </Results>
    )
  }
  return <></>
}
export default ResultList
