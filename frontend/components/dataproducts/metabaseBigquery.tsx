import { ExternalLinkIcon } from '@navikt/aksel-icons'
import { BodyShort, Box, Button, List, Loader, Modal } from '@navikt/ds-react'
import * as React from 'react'
import { DatasetWithAccess } from '../../lib/rest/generatedDto'
import DeleteModal from '../lib/deleteModal'
import MetabaseSync from './metabaseSync'
import {
    useClearMetabaseBigqueryJobs,
    useCreateMetabaseBigQueryOpenDataset,
    useCreateMetabaseBigQueryRestrictedDataset,
    useDeleteMetabaseBigQueryOpenDataset,
    useDeleteMetabaseBigQueryRestrictedDataset,
    useGetMetabaseBigQueryOpenDatasetPeriodically,
    useGetMetabaseBigQueryRestrictedDatasetPeriodically,
    useOpenRestrictedMetabaseBigQueryDataset,
} from './queries'

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

  const [showDeleteConfirm, setShowDeleteConfirm] = React.useState(false)
  const [showOpenRestrictedMetabaseConfirm, setShowOpenRestrictedMetabaseConfirm] = React.useState(false)

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
    if (dataset.metabaseDataset?.Type === 'open') {
      deleteOpenDataset.mutate()
    } else {
      deleteRestrictedDataset.mutate()
    }
    setShowDeleteConfirm(false)
  }

  const handleOpenRestrictedMetabaseDatabase = () => {
    if (hasAllUsers) openRestrictedDataset.mutate()
    setShowOpenRestrictedMetabaseConfirm(false)
  }

  const isDeleting = deleteOpenDataset.isPending || deleteRestrictedDataset.isPending
  const isOpeningRestricted = openRestrictedDataset.isPending

  // If URL exists, show the link
  if (dataset.metabaseDataset?.URL) {
    return (
      <div className="flex flex-col">
        <Modal
          open={showOpenRestrictedMetabaseConfirm}
          aria-label="Åpne metabase datasett for alle i Nav"
          onClose={() => setShowOpenRestrictedMetabaseConfirm(false)}
          className='w-full ax-md:w-[60rem] px-8'
          header={{ heading: "Er du sikker på at du vil åpne opp datasettet i Metabase for alle i Nav?" }}
        >
          <Modal.Body className='h-full'>
            <BodyShort className='mt-4'><strong>Dette betyr at alle ansatte i Nav vil få tilgang til:</strong></BodyShort>
            <Box marginBlock="space-16" asChild><List data-aksel-migrated-v8 as="ul">
                <List.Item>
                  Databasen i Metabase
                </List.Item>
                <List.Item>
                  Collectionen knyttet til databasen (inkludert dashboards og spørsmål)
                </List.Item>
              </List></Box>
          </Modal.Body>
          <Modal.Footer>
            <Button
              onClick={handleOpenRestrictedMetabaseDatabase}
              variant="primary"
              size="small"
              disabled={openRestrictedDataset.isPending}
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
            {openRestrictedDataset.isPending && <Loader />}
          </Modal.Footer>
        </Modal>
        <a
          className="border-l-8 border-ax-border-neutral-subtle pl-4 py-1 pr-4 w-fit"
          target="_blank"
          rel="noreferrer"
          href={dataset.metabaseDataset.URL}
        >
          Åpne i Metabase <ExternalLinkIcon />
        </a>
        {isOwner && dataset.metabaseDataset && (
          <>
            <DeleteModal
              open={showDeleteConfirm}
              onCancel={() => setShowDeleteConfirm(false)}
              onConfirm={handleDelete}
              name={dataset.name}
              error=""
              resource="metabase-remove"
              warning="Datasettet vil bli fjernet fra Metabase. Tilganger og eventuelle spørsmål som bruker denne databasen som kilde vil bli slettet."
              confirmText="Jeg forstår at datasettet fjernes fra Metabase og at dette ikke kan angres."
            />
            <div className='flex flex-col gap-1'>
              {isDeleting ? (
                <p className="border-l-8 border-ax-border-neutral-subtle py-1 px-4 flex flex-row gap-2 w-fit text-ax-text-neutral-subtle mt-2">
                  Fjerner fra Metabase
                  <Loader transparent size="small" />
                </p>
              ) : (
                <>
                  <a
                    className="border-l-8 border-ax-border-neutral-subtle pl-4 py-1 pr-4 w-fit mt-2"
                    href="#"
                    onClick={(e) => {
                      e.preventDefault()
                      setShowDeleteConfirm(true)
                    }}
                  >
                    Fjern datasettet fra Metabase
                  </a>
                  {hasAllUsers && dataset.metabaseDataset?.Type === "restricted" && (
                    isOpeningRestricted ? (
                      <p className="border-l-8 border-ax-border-neutral-subtle py-1 px-4 flex flex-row gap-2 w-fit text-ax-text-neutral-subtle">
                        Åpner for alle i Nav
                        <Loader transparent size="small" />
                      </p>
                    ) : (
                      <a
                        className="border-l-8 border-ax-border-neutral-subtle pl-4 py-1 pr-4 w-fit"
                        href="#"
                        onClick={(e) => {
                          e.preventDefault()
                          setShowOpenRestrictedMetabaseConfirm(true)
                        }}
                      >
                        Åpne opp datasett i Metabase for alle i Nav
                      </a>
                    )
                  )}
                </>
              )}
            </div>
          </>
        )}
      </div>
    );
  }

  if (isOwner) {
    if (datasetStatus.data?.isRunning) {
      return (
        <div>
          <p className="border-l-8 border-ax-border-neutral-subtle py-1 px-4 flex flex-row gap-2 w-fit text-ax-text-neutral-subtle">
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
        className="border-l-8 border-ax-border-neutral-subtle pl-4 py-1 pr-4 w-fit"
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
