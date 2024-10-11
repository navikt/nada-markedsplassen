import * as React from 'react'
import { Detail, Heading, Link } from '@navikt/ds-react'
import ReactMarkdown from 'react-markdown'
import { PluggableList } from 'react-markdown/lib'
import remarkGfm from 'remark-gfm'
import { CoApplicant, Table } from '@navikt/ds-icons'
import humanizeDate from '../../lib/humanizeDate'
import DeleteModal from '../lib/deleteModal'
import { useState } from 'react'
import { useRouter } from 'next/router'
import TagPill from '../lib/tagPill'
import { deleteInsightProduct } from '../../lib/rest/insightProducts'

export interface SearchResultProps {
  resourceType?: string
  link: string
  id?: string
  name: string
  innsiktsproduktType?: string
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
}

export const SearchResultLink = ({
  resourceType,
  link,
  id,
  type,
  keywords,
  name,
  innsiktsproduktType,
  group,
  description,
  datasets,
  teamkatalogenTeam,
  productArea,
  editable,
  deleteResource
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
    }
  }
  const openDeleteModal = () => setModal(true)

  const confirmDelete = () => {
    const deletePromise = resourceType == "innsiktsprodukt"?
      deleteInsightProduct(id || ''):
    deleteResource?.(id || '');
    deletePromise?.then(() => {
      setModal(false)
      router.reload()
    }).catch((reason) => {
      setError(reason.toString())
    })
  }

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
      <Link href={link} className="nada-search-result w-[47rem]">
        <div className="flex flex-col w-full px-4 py-2 gap-2">
          <div className="flex flex-col">
            <div className='flex flex-row justify-between'>
              <div>
                {
                  // have to ignore in order to use dangerouslySetInnerHTML :(
                  //@ts-ignore
                  <Heading className="text-text-action" level="2" size="small"
                    dangerouslySetInnerHTML={{ __html: `${name.replaceAll("_", "_<wbr>")} ${innsiktsproduktType ? '(' + innsiktsproduktType + ')' : ''}` }} />
                }
              </div>
              {editable && <div>
                <Link className="m-2" href="#" onClick={editResource}
                >Endre metadata</Link>
                <Link className='m-2' href="#" onClick={openDeleteModal}>Slett</Link>
              </div>}
            </div>
            <Detail className="flex gap-2 items-center text-text-subtle"><CoApplicant /> {owner + `${productArea ? " - " + productArea : ""}`}</Detail>
          </div>
          <div className="flex flex-col gap-4">
            {description && (
              <ReactMarkdown
                remarkPlugins={[remarkGfm] as PluggableList}
                disallowedElements={['h1', 'h2', 'h3', 'h4', 'h5', 'code', 'pre']}
                unwrapDisallowed={true}
              >
                {description.split(/\r?\n/).slice(0, 4).join('\n').replaceAll("((START))", "_").replaceAll("((STOP))", "_")}
              </ReactMarkdown>
            )}
            {datasets && !!datasets.length && (
              <div>
                <Heading size="xsmall" level="3" className="flex items-center gap-2"><Table /> Datasett</Heading>
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
