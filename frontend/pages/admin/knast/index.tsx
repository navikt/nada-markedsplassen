import { Alert, Button, Checkbox, Heading, Link, Modal, Table } from "@navikt/ds-react"
import { deleteWorkstation, useListWorkstationsPeriodically } from "../../../lib/rest/workstation"
import LoaderSpinner from "../../../components/lib/spinner"
import ErrorStripe from "../../../components/lib/errorStripe"
import { TrashIcon } from "@navikt/aksel-icons"
import { useEffect, useState } from "react"
import { WorkstationOutput } from "../../../lib/rest/generatedDto"
import { useRouter } from "next/router"

const KnastPasge = () => {
  const { data: workstations, isLoading, error } = useListWorkstationsPeriodically()
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [selectedKNAST, setSelectedKNAST] = useState<WorkstationOutput|null>(null)
  const [confirmDeleteKnast, setConfirmDeleteKnast] = useState(false)
  const [deleteError, setDeleteError] = useState('')
  const [showDeleteInfo, setShowDeleteInfo] = useState(false)

  const router = useRouter()

  const closeDeleteModal = () => setShowDeleteModal(false)
  
  const deleteKnast = (slug: string) => {
    setDeleteError('')
    deleteWorkstation(slug).then(
      () => {
        setShowDeleteInfo(false)
      }
    ).catch(error => {
      setDeleteError(error.message)
    })
    setShowDeleteModal(false)
    setShowDeleteInfo(true)
  }

  const onTrash= (knast: WorkstationOutput)=>{
    setSelectedKNAST(knast)
    setConfirmDeleteKnast(false)
    setShowDeleteModal(true)
  }

  if (isLoading) return <div><LoaderSpinner />Laster...</div>
  if (error) return <ErrorStripe error={error} />
  return <div>
    <Modal open={showDeleteModal && !!selectedKNAST} onClose={closeDeleteModal} header={{ heading: "Slett KNAST" }}>
      <Modal.Body className="flex flex-col gap-4">
        <p>Denne operasjonen vil permanent slette KNAST eid av: {selectedKNAST?.displayName}</p>
        <Checkbox className='mt-2' checked={confirmDeleteKnast} onClick={() => setConfirmDeleteKnast(!confirmDeleteKnast)}>
          Jeg forstår at operasjonen vil slette dataproduktet samt datasettene ovenfor, og at dette ikke kan angres.
        </Checkbox>
        <div className="flex flex-row gap-3">
          <Button variant="secondary" onClick={closeDeleteModal}>
            Avbryt
          </Button>
          <Button onClick={()=>deleteKnast(selectedKNAST!!.slug)} disabled={!confirmDeleteKnast || !selectedKNAST?.slug}>Slett</Button>
        </div>
      </Modal.Body>
    </Modal>
    <Heading size="large">Knast</Heading>
    {deleteError && <Alert variant={'error'}>{deleteError}</Alert>}
    {showDeleteInfo && <Alert variant={'info'}>Sletting av KNAST kan ta flere minutter å fullføre</Alert>}
    <Table>
      <Table.Header>
        <Table.Row className="border-none border-transparent">
          <Table.HeaderCell>Bruker</Table.HeaderCell>
          <Table.HeaderCell>Maskintype</Table.HeaderCell>
          <Table.HeaderCell>Opprettelse</Table.HeaderCell>
          <Table.HeaderCell>Oppdatering</Table.HeaderCell>
          <Table.HeaderCell>Handlinger</Table.HeaderCell>
          <Table.HeaderCell />
          <Table.HeaderCell />
        </Table.Row>
      </Table.Header>
      {workstations?.map((w, i) => (
        <>
          <Table.Row
            className={i % 2 === 0 ? 'bg-[#f7f7f7]' : ''}
            key={i + '-access'}
          >
            <Table.DataCell className="w-72">{w.displayName}</Table.DataCell>
            <Table.DataCell className="w-36">
              {w.config?.machineType}
            </Table.DataCell>
            <Table.DataCell className="w-48">
              {new Date(w.createTime).toString()}
            </Table.DataCell>
            <Table.DataCell className="w-48">
              {new Date(w.updateTime || '').toString()}
            </Table.DataCell>
            <Table.DataCell className="w-[207px]">
              <Link href="#" onClick={()=>onTrash(w)}><TrashIcon title="a11y-title" fontSize="2rem"/></Link>
            </Table.DataCell>
          </Table.Row>
        </>
      ))}
    </Table>
  </div>
}

export default KnastPasge