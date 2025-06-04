import * as React from 'react'
import { Button, Loader } from '@navikt/ds-react'
import { ExternalLinkIcon } from '@navikt/aksel-icons'
import {
  useCreateMetabaseBigQueryOpenDataset,
  useCreateMetabaseBigQueryRestrictedDataset,
  useDeleteMetabaseBigQueryOpenDataset,
  useDeleteMetabaseBigQueryRestrictedDataset,
  useGetMetabaseBigQueryOpenDatasetPeriodically,
  useGetMetabaseBigQueryRestrictedDatasetPeriodically,
} from './queries'
import { DatasetWithAccess } from '../../lib/rest/generatedDto'
import MetabaseSync from './metabaseSync'

interface MetabaseBigQueryLinkProps {
  dataset: DatasetWithAccess
  isOwner: boolean
  url: string | null | undefined
  metabaseDeletedAt: string | null | undefined
}

const MetabaseBigQueryIntegration: React.FC<MetabaseBigQueryLinkProps> = (
  {
    dataset,
    isOwner,
    url,
    metabaseDeletedAt,
  },
) => {
  const hasAllUsers: boolean = (() => {
      if (dataset.access === undefined) return false

      for (const accessItem of dataset.access) {
        console.log('Access item:', accessItem)
        if (accessItem != undefined && accessItem.subject.toLowerCase() === 'group:all-users@nav.no') {
          return true
        }
      }
      return false
    })()

  const createOpenDataset = useCreateMetabaseBigQueryOpenDataset(dataset.id)
  const createRestrictedDataset = useCreateMetabaseBigQueryRestrictedDataset(dataset.id)
  const deleteOpenDataset = useDeleteMetabaseBigQueryOpenDataset(dataset.id)
  const deleteRestrictedDataset = useDeleteMetabaseBigQueryRestrictedDataset(dataset.id)

  const openDatasetStatus = useGetMetabaseBigQueryOpenDatasetPeriodically(dataset.id)
  const restrictedDatasetStatus = useGetMetabaseBigQueryRestrictedDatasetPeriodically(dataset.id)

  const datasetStatus = hasAllUsers ? openDatasetStatus : restrictedDatasetStatus
  const [isDeleting, setIsDeleting] = React.useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = React.useState(false)

  const handleCreate = () => {
    if (hasAllUsers) {
      createOpenDataset.mutate()
    } else {
      createRestrictedDataset.mutate()
    }
  }

  const handleDelete = () => {
    setIsDeleting(true)
    if (hasAllUsers) {
      deleteOpenDataset.mutate()
    } else {
      deleteRestrictedDataset.mutate()
    }

    setTimeout(() => {
      window.location.reload()
      setShowDeleteConfirm(false)
    }, 5000)
  }

  // If URL exists, show the link
  if (url) {
    return (
      <div className="flex flex-col">
        <a
          className="border-l-8 border-border-on-inverted pl-4 py-1 pr-4 w-fit"
          target="_blank"
          rel="noreferrer"
          href={url}
        >
          Åpne i Metabase <ExternalLinkIcon />
        </a>

        {isOwner && metabaseDeletedAt == null && (
          <>
            {showDeleteConfirm ? (
              <div className="mt-2 border-l-8 border-border-on-inverted pl-4 py-1 pr-4">
                <p>Er du sikker på at du vil fjerne datasettet fra Metabase?</p>
                {isDeleting ? (
                  <p className="flex flex-row gap-2 text-text-subtle">
                    Fjerner datasettet fra Metabase
                    <Loader transparent size="small" />
                  </p>
                ) : (
                  <div className="mt-2">
                    <Button onClick={handleDelete} variant="primary" size="small" className="mr-2">
                      Ja, fjern
                    </Button>
                    <Button onClick={() => setShowDeleteConfirm(false)} variant="secondary" size="small">
                      Avbryt
                    </Button>
                  </div>
                )}
              </div>
            ) : (
              <a
                className="border-l-8 border-border-on-inverted pl-4 py-1 pr-4 w-fit mt-2"
                href="#"
                onClick={(e) => {
                  e.preventDefault()
                  setShowDeleteConfirm(true)
                }}
              >
                Fjern datasettet fra Metabase
              </a>
            )}
          </>
        )}
      </div>
    )
  }

  // If user is owner and no URL exists yet
  if (isOwner) {
    if (datasetStatus.data?.isRunning) {
      return (
        <div>
          <p className="border-l-8 border-border-on-inverted py-1 px-4 flex flex-row gap-2 w-fit text-text-subtle">
            Legger til i Metabase
            <Loader transparent size="small" />
          </p>
          {datasetStatus.data?.isRunning && (
            <div className="mt-2 pl-4">
              <MetabaseSync status={datasetStatus.data} />
            </div>
          )}
        </div>
      )
    }

    return (
      <a
        className="border-l-8 border-border-on-inverted pl-4 py-1 pr-4 w-fit"
        href="#"
        onClick={(e) => {
          e.preventDefault()
          handleCreate()
        }}
      >
        Legg til i Metabase som en {hasAllUsers ? 'åpen' : 'lukket'} kilde {hasAllUsers ? "🔓" : "🔐"}
      </a>
    )
  }

  // If not owner and no URL, show nothing
  return null
}

export default MetabaseBigQueryIntegration
