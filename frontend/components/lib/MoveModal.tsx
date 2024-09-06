import * as React from 'react'
import {
  Modal,
  Button,
  Heading,
  Select,
} from '@navikt/ds-react'
import LoaderSpinner from './spinner'
import { useState } from 'react'
import ErrorMessage from './error'
import { useSearch } from '../../lib/rest/search'

interface MoveModalProps {
  open: boolean
  onCancel: () => void
  onConfirm: (dataproductID: string) => void
  group: string
  currentDataproductID: string
}
export const MoveModal = ({
  open,
  onCancel,
  onConfirm,
  group,
  currentDataproductID,
}: MoveModalProps) => {
  const [dataproductID, setDataproductID] = useState<string | undefined>(
    undefined
  )
  const [error, setError] = useState<string | undefined>(undefined)
  const search = useSearch({ types: ['dataproduct'], groups: [group] })

  if (search.error) return <ErrorMessage error={search.error} />
  if (search.loading || !search.data?.results) return <LoaderSpinner />

  return (
    <Modal open={open} onClose={onCancel} header={{heading: "Flytt datasett"}}>
      <Modal.Body className="flex flex-col gap-4">
        <p>Velg hvilket dataprodukt datasett skal flyttes til.</p>
        <p>
          (Du kan kun flytte datasett til dataprodukter som hører til samme
          team)
        </p>
        <Select
          className="w-full"
          error={error}
          label={undefined}
          onChange={(event) => setDataproductID(event.target.value)}
        >
          <option value="">Velg dataprodukt</option>
          {search.data?.results?.map(
            (dp) =>
              currentDataproductID !== dp.result.id && (
                <option value={dp.result.id} key={dp.result.id}>
                  {dp.result.name}
                </option>
              )
          )}
        </Select>
        <div className="flex flex-row gap-3">
          <Button variant="secondary" onClick={onCancel}>
            Avbryt
          </Button>
          <Button
            onClick={() =>
              dataproductID
                ? onConfirm(dataproductID)
                : setError(
                    'Du må velge hvilket dataprodukt datasettet skal flyttes til'
                  )
            }
          >
            Flytt
          </Button>
        </div>
      </Modal.Body>
    </Modal>
  )
}
export default MoveModal
