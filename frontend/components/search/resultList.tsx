import LoaderSpinner from '../lib/spinner'
import SearchResultLink from './searchResultLink'
import { Tabs } from '@navikt/ds-react'
import React, { useContext } from 'react'
import { SearchParam } from '../../pages/search'
import { UserState } from '../../lib/context'
import { useSearchTeamKatalogen } from '../../lib/rest/teamkatalogen'
import { useGetProductAreas } from '../../lib/rest/productAreas'
import { SearchResult } from '../../lib/rest/search'
import { deleteStory } from '../../lib/rest/stories'
import { useRouter } from 'next/router'
import { Dataproduct, InsightProduct, Story } from '../../lib/rest/generatedDto'
import ErrorStripe from "../lib/errorStripe";

const Results = ({ children }: { children: React.ReactNode }) => (
  <div className="results">{children}</div>
)

type ResultListInterface = {
  search?: {data: SearchResult|undefined, isLoading: boolean, error: any}
  dataproducts?: Dataproduct[]
  stories?: Story[]
  insightProducts?: InsightProduct[]
  searchParam?: SearchParam
  updateQuery?: (updatedParam: SearchParam) => void
}

const ResultList = ({
  search,
  dataproducts,
  stories,
  insightProducts,
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
    return (
      <Results>
        {dataproducts.map((d, idx) => (
          <SearchResultLink
            key={idx}
            group={d.owner}
            name={d.name}
            keywords={d.keywords}
            link={`/dataproduct/${d.id}/${d.slug}`}
            {...getTeamKatalogenInfo(d.owner?.teamkatalogenURL)}
          />
        ))}
      </Results>
    )
  }

  if (stories) {
    return (
      <div>
        <Results>
          {stories?.map((s, idx) => (
            <SearchResultLink
              key={idx}
              group={{
                group: s.group,
                teamkatalogenURL: s.teamkatalogenURL,
              }}
              id={s.id}
              name={s.name}
              resourceType={"datafortelling"}
              link={`/story/${s.id}`}
              {...getTeamKatalogenInfo(s.teamkatalogenURL)}
              keywords={s.keywords}
              editable={true}
              description={s.description}
              deleteResource={()=> deleteStory(s.id).then(()=>router.reload())}
            />
          ))}
        </Results>
      </div>
    )
  }

  if (insightProducts) {
    return (
      <div>
        <Results>
          {insightProducts?.map((p, idx) => (
            <SearchResultLink
              key={idx}
              group={{
                group: p.group,
                teamkatalogenURL: p.teamkatalogenURL,
              }}
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
        </Results>
      </div>
    )
  }
  return <></>
}
export default ResultList
