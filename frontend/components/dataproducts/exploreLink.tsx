import * as React from 'react'
import { Button, Heading, Link, Loader, Modal } from '@navikt/ds-react'
import { ExternalLink } from '@navikt/ds-icons'
import { useState } from 'react'
import { MappingService } from './explore'

export enum ItemType {
  metabase = 1,
  bigQuery,
}

export interface ExploreLinkProps {
  datasetID: string
  url?: string | null
  type: ItemType
  add?: () => void
  remove?: (datasetID: string) => void
  isOwner?: boolean
  mappings?: MappingService[]
  metabaseDeletedAt?: string | null
}

export const ExploreLink = ({
  datasetID,
  url,
  type,
  add,
  remove,
  isOwner,
  mappings,
  metabaseDeletedAt,
}: ExploreLinkProps) => {
  const [showRemoveMapping, setShowRemoveMapping] = useState(false)
  const addToMetabase = !mappings?.includes(MappingService.Metabase)
  const [loading, setLoading] = useState(mappings?.includes(MappingService.Metabase) && !url)
  const [deleteInMetabaseInProgress, setDeleteInMetabaseInProgress] = useState(false);
  const handleDelete = (e: any) => {
    e.preventDefault()
    setDeleteInMetabaseInProgress(true);
    if (remove) {
      remove(datasetID)
      setTimeout(() => {
        window.location.reload();
        setShowRemoveMapping(false)
      }, 5000);
    }
  }

  const handleAdd = () => {
    if (add) {
      setLoading(true)
      add()
    }
  }

  if (url) {
    if (type === ItemType.bigQuery) {
      return (
        <Link
          className="border-l-8 border-border-on-inverted pl-4 py-1 pr-4 w-fit"
          target="_blank"
          rel="norefferer"
          href={url}
        >
          Åpne i Google Cloud Console <ExternalLink />
        </Link>
      )
    }
    if (type === ItemType.metabase) {
      return (
        <div className="flex flex-col">
          <Link
            className="border-l-8 border-border-on-inverted pl-4 py-1 pr-4 w-fit"
            target="_blank"
            rel="norefferer"
            href={url}
          >
            Åpne i Metabase <ExternalLink />
          </Link>
          {isOwner && metabaseDeletedAt == null && (
            <>
              <Modal
                open={showRemoveMapping}
                aria-label="Fjern metabase database"
                onClose={() => setShowRemoveMapping(false)}
                className="max-w-full md:max-w-3xl px-8 h-[20rem]"
              >
                <Modal.Body className="h-full">
                  <div className="flex flex-col gap-8">
                    <Heading level="1" size="medium">
                    Er du sikker på at du vil fjerne datasettet fra metabase?
                    </Heading>
                    <div>
                        Dette vil medføre at du sletter databasen, samlingen, tilgangsgruppene og alle tilhørende spørsmål i metabase, samt service account i GCP.
                    </div>
                    <div className="flex flex-row gap-4">
                      {deleteInMetabaseInProgress ? (
                        <p
                          className="border-l-8 border-border-on-inverted py-1 px-4 flex flex-row gap-2 w-fit text-text-subtle"
                         >
                            Fjerner datasettet fra Metabase
                            <Loader transparent size="small" />
                        </p> ) : (

                        <div>
                      <Button
                        onClick={handleDelete}
                        variant="primary"
                        size="small"
                      >
                        Fjern
                      </Button>
                      <Button
                        onClick={() => setShowRemoveMapping(false)}
                        variant="secondary"
                        size="small"
                      >
                        Avbryt
                      </Button>
                        </div>
                        )
                    }
                    </div>
                  </div>
                </Modal.Body>
              </Modal>
                <Link
                  className="border-l-8 border-border-on-inverted pl-4 py-1 pr-4 w-fit"
                  href="#"
                  onClick={() => setShowRemoveMapping(true)}
                >
                  Fjern datasettet fra Metabase
                </Link>
            </>
          )}
        </div>
      )
    }
  }

  if (isOwner) {
    if (loading) {
      return (
        <p
          className="border-l-8 border-border-on-inverted py-1 px-4 flex flex-row gap-2 w-fit text-text-subtle"
        >
          Legger til i Metabase
          <Loader transparent size="small" />
        </p>
      )
    }
    if (addToMetabase) {
      return (
        <Link
          className="border-l-8 border-border-on-inverted pl-4 py-1 pr-4 w-fit"
          href="#"
          onClick={handleAdd}
        >
          Legg til i Metabase
        </Link>
      )
    }
  }
  return <></>
}

export default ExploreLink
