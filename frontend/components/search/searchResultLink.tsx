import { ExternalLinkIcon, PersonGroupIcon, TableIcon } from '@navikt/aksel-icons'
import { Detail, Heading, Link } from '@navikt/ds-react'
import { useRouter } from 'next/router'
import React, { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import humanizeDate from '../../lib/humanizeDate'
import { deleteInsightProduct } from '../../lib/rest/insightProducts'
import DeleteModal from '../lib/deleteModal'
import TagPill from '../lib/tagPill'

export interface SearchResultProps {
  resourceType?: string
  link: string
  id?: string
  name: string
  innsiktsproduktType?: string
  lastModified?: string
  group?: {
    group?: string | null,
    teamkatalogenURL?: string | null
  } | null
  keywords?: string[]
  type?: string
  description?: string
  datasets?: {
    name: string
    dataSourceLastModified: string
  }[]
  teamkatalogenTeam?: string,
  productArea?: string,
  editable?: boolean,
  deleteResource?: (id: string) => Promise<any>,
	externalLink?: boolean
}

export const SearchResultLink = ({
  resourceType,
  link,
  id,
  type,
  keywords,
  name,
  innsiktsproduktType,
  lastModified,
  group,
  description,
  datasets,
  teamkatalogenTeam,
  productArea,
  editable,
  deleteResource,
	externalLink
}: SearchResultProps) => {
  const [modal, setModal] = useState(false)

  const owner = teamkatalogenTeam || group?.group
  const router = useRouter();
  const [error, setError] = useState<string | undefined>(undefined)

  const editResource = () => {
    if (resourceType == 'datafortelling') {
      router.push(`/stories/${id}/edit`)
    } else if (resourceType == 'innsiktsprodukt') {
      router.push(`/insightProduct/edit?id=${id}`)
    } else if (isMetabaseDashboard) {
      router.push(`/metabaseDashboard/edit?id=${id}`)
    }
  }
  const openDeleteModal = () => setModal(true)

  const confirmDelete = () => {
    const deletePromise = resourceType == "innsiktsprodukt" ?
      deleteInsightProduct(id || '') :
      deleteResource?.(id || '');
    deletePromise?.then(() => {
      setModal(false)
      router.reload()
    }).catch((reason) => {
      setError(reason.toString())
    })
  }

	const isMetabaseDashboard = resourceType === "metabase-dashboard"

  return (
    <div>
      <DeleteModal
        open={modal}
        onCancel={() => setModal(false)}
        onConfirm={confirmDelete}
        name={name}
        error={error || ''}
        resource={resourceType || ''}
      />
      <Link href={link} className="nada-search-result w-[47rem]" {...(externalLink ? {target: '_blank', rel: 'noopener noreferrer'} : {})}>
        <div className="flex flex-col w-full px-4 py-2 gap-2">
          <div className="flex flex-col">
            <div className='flex flex-row justify-between'>
              <div>
								<Heading className="text-text-action flex flex-row items-center break-all" level="2" size="small">
									{name}
									{innsiktsproduktType && ` (${innsiktsproduktType})`}
									{externalLink && <ExternalLinkIcon className='ml-0.5 mb-0.5' />}
								</Heading>
              </div>
              {editable && <div>
								<Link className="m-2" href="#" onClick={editResource}>Endre metadata</Link> :
                <Link className='m-2' href="#" onClick={openDeleteModal}>{isMetabaseDashboard ? "Fjern public lenke" : "Slett"}</Link>
              </div>}
            </div>
            <Detail className="flex gap-2 items-center text-text-subtle"><PersonGroupIcon title="a11y-title" />
              {owner + `${productArea ? " - " + productArea : ""}`}</Detail>
          </div>
          <div className="flex flex-col gap-4">
            {description && (
              <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                disallowedElements={['h1', 'h2', 'h3', 'h4', 'h5', 'code', 'pre']}
                unwrapDisallowed={true}
              >
                {description.split(/\r?\n/).slice(0, 4).join('\n').replaceAll("((START))", "_").replaceAll("((STOP))", "_")}
              </ReactMarkdown>
            )}
            {lastModified && (
                <Detail className="text-text-subtle">Sist oppdatert: {humanizeDate(lastModified)}</Detail>
            )}
            {datasets && !!datasets.length && (
              <div>
                <Heading size="xsmall" level="3" className="flex items-center gap-2"><TableIcon title="a11y-title" />
                  Datasett</Heading>
                <div className='ml-[1.6rem] flex flex-col gap-2'>
                  {datasets.map((ds, index) => (
                    <div key={index}>
                      <p dangerouslySetInnerHTML={{ __html: ds.name.replaceAll("_", "_<wbr>") }} />
                      <Detail className="text-text-subtle">Sist oppdatert: {humanizeDate(ds.dataSourceLastModified)}</Detail>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
          {keywords && keywords?.length > 0 &&
            <div className='flex gap-x-1'>
              {keywords.map((k, i) => {
                return (
                  <TagPill
                    key={i}
                    keyword={k}
                    horizontal={true}
                  >
                    {k}
                  </TagPill>
                )
              })
              }
            </div>
          }
        </div>
      </Link>
    </div>
  );
}

export default SearchResultLink
