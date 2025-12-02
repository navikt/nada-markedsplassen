import * as React from 'react'
import { BodyShort, Button, List, Loader } from '@navikt/ds-react'
import { ExternalLinkIcon } from '@navikt/aksel-icons'
import {
  useClearMetabaseBigqueryJobs,
  useCreateMetabaseBigQueryOpenDataset,
  useCreateMetabaseBigQueryRestrictedDataset,
  useOpenRestrictedMetabaseBigQueryDataset,
  useDeleteMetabaseBigQueryOpenDataset,
  useDeleteMetabaseBigQueryRestrictedDataset,
  useGetMetabaseBigQueryOpenDatasetPeriodically,
  useGetMetabaseBigQueryRestrictedDatasetPeriodically,
} from './queries'
import { DatasetWithAccess } from '../../lib/rest/generatedDto'
import MetabaseSync from './metabaseSync'
import { Modal } from '@navikt/ds-react'

interface MetabaseBigQueryLinkProps {
  dataset: DatasetWithAccess
  isOwner: boolean
}

const MetabaseBigQueryIntegration: React.FC<MetabaseBigQueryLinkProps> = (
  {
    dataset,
    isOwner,
  },
) => {

  const hasAllUsers: boolean = (() => {
    if (dataset.access === undefined) return false

    let hasAllUsers = false
    dataset.access.flatMap(a => a?.active).forEach(access => {
      if (access?.subject === 'group:all-users@nav.no') hasAllUsers = true
    })

    return hasAllUsers
  })()

  const createOpenDataset = useCreateMetabaseBigQueryOpenDataset(dataset.id)
  const createRestrictedDataset = useCreateMetabaseBigQueryRestrictedDataset(dataset.id)
  const deleteOpenDataset = useDeleteMetabaseBigQueryOpenDataset(dataset.id)
  const openRestrictedDataset = useOpenRestrictedMetabaseBigQueryDataset(dataset.id)
  const deleteRestrictedDataset = useDeleteMetabaseBigQueryRestrictedDataset(dataset.id)
  const clearMetabaseJobs = useClearMetabaseBigqueryJobs(dataset.id)

  const openDatasetStatus = useGetMetabaseBigQueryOpenDatasetPeriodically(dataset.id)
  const restrictedDatasetStatus = useGetMetabaseBigQueryRestrictedDatasetPeriodically(dataset.id)

  const datasetStatus = hasAllUsers ? openDatasetStatus : restrictedDatasetStatus
  const [isDeleting, setIsDeleting] = React.useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = React.useState(false)
  const [showOpenRestrictedMetabaseConfirm, setShowOpenRestrictedMetabaseConfirm] = React.useState(false)
  const [openingMetabaseDatabase, setOpeningMetabaseDatabase] = React.useState(false)

  const handleReset = () => {
    clearMetabaseJobs.mutate()
    handleCreate()
  }

  const handleCreate = () => {
    if (hasAllUsers) {
      createOpenDataset.mutate()
    } else {
      createRestrictedDataset.mutate()
    }
  }

  const handleDelete = () => {
    setIsDeleting(true)
    if (dataset.metabaseDataset?.Type === 'open') {
      deleteOpenDataset.mutate()
    } else {
      deleteRestrictedDataset.mutate()
    }

    setTimeout(() => {
      window.location.reload()
      setShowDeleteConfirm(false)
    }, 5000)
  }

  const handleOpenRestrictedMetabaseDatabase = () => {
    setOpeningMetabaseDatabase(true)
    if (hasAllUsers) openRestrictedDataset.mutate()
    setTimeout(() => {
      window.location.reload()
      setShowOpenRestrictedMetabaseConfirm(false)
    }, 5000)
  }

  // If URL exists, show the link
  if (dataset.metabaseDataset?.URL) {
    return (
      <div className="flex flex-col">
        <Modal
          open={showOpenRestrictedMetabaseConfirm}
          aria-label="Åpne metabase datasett for alle i Nav"
          onClose={() => setShowOpenRestrictedMetabaseConfirm(false)}
          className='w-full md:w-[60rem] px-8'
          header={{ heading: "Er du sikker på at du vil åpne opp datasettet i Metabase for alle i Nav?" }}
        >
          <Modal.Body className='h-full'>
            <BodyShort className='mt-4'><strong>Dette betyr at alle ansatte i Nav vil få tilgang til:</strong></BodyShort>
            <List as="ul">
              <List.Item>
                Databasen i Metabase
              </List.Item>
              <List.Item>
                Collectionen knyttet til databasen (inkludert dashboards og spørsmål)
              </List.Item>
            </List>
          </Modal.Body>
          <Modal.Footer>
            <Button
              onClick={handleOpenRestrictedMetabaseDatabase}
              variant="primary"
              size="small"
              disabled={openingMetabaseDatabase}
            >
              Bekreft
            </Button>
            <Button
              onClick={() => setShowOpenRestrictedMetabaseConfirm(false)}
              variant="secondary"
              size="small"
            >
              Avbryt
            </Button>
            {openingMetabaseDatabase && <Loader />}
          </Modal.Footer>
        </Modal>
        <a
          className="border-l-8 border-border-on-inverted pl-4 py-1 pr-4 w-fit"
          target="_blank"
          rel="noreferrer"
          href={dataset.metabaseDataset.URL}
        >
          Åpne i Metabase <ExternalLinkIcon />
        </a>

        {isOwner && dataset.metabaseDataset && (
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
              <div className='flex flex-col gap-1'>
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
                <a>
                  {hasAllUsers && dataset.metabaseDataset?.Type === "restricted" && <a
                    className="border-l-8 border-border-on-inverted pl-4 py-1 pr-4 w-fit"
                    href="#"
                    onClick={(e) => {
                      e.preventDefault()
                      setShowOpenRestrictedMetabaseConfirm(true)
                    }}
                  >
                    Åpne opp datasett i Metabase for alle i Nav
                  </a>}
                </a>
              </div>
            )}
          </>
        )}
      </div>
    )
  }

  if (isOwner) {
    if (datasetStatus.data?.isRunning) {
      return (
        <div>
          <p className="border-l-8 border-border-on-inverted py-1 px-4 flex flex-row gap-2 w-fit text-text-subtle">
            Legger til i Metabase
            <Loader transparent size="small" />
          </p>
          <div className="mt-2 pl-4">
            <MetabaseSync handleReset={handleReset} status={datasetStatus.data} />
          </div>
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
        Legg til i Metabase som en {hasAllUsers ? 'åpen' : 'lukket'} kilde {hasAllUsers ? '🔓' : '🔐'}
      </a>
    )
  }

  // If not owner and no URL, show nothing
  return null
}

export default MetabaseBigQueryIntegration
