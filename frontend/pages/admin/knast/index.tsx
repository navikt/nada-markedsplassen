import { Alert, Button, Checkbox, Heading, Link, Loader, Modal, Table } from "@navikt/ds-react"
import { deleteWorkstation, useListWorkstationsPeriodically, resyncWorkstation, createWorkstationResyncAllWorkflow } from "../../../lib/rest/workstation"
import LoaderSpinner from "../../../components/lib/spinner"
import ErrorStripe from "../../../components/lib/errorStripe"
import { TrashIcon, ArrowCirclepathIcon } from "@navikt/aksel-icons"
import { useEffect, useState } from "react"
import { WorkstationOutput } from "../../../lib/rest/generatedDto"
import { useRouter } from "next/router"
import { set } from "lodash"
import Head from "next/head"

const KnastPasge = () => {
  const { data: workstations, isLoading, error } = useListWorkstationsPeriodically()
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [showResyncModal, setShowResyncModal] = useState(false)
  const [showResyncAllModal, setShowResyncAllModal] = useState(false)
  const [deleteError, setDeleteError] = useState('')
  const [resyncError, setResyncError] = useState('')
  const [resyncAllError, setResyncAllError] = useState('')
  const [selectedKNAST, setSelectedKNAST] = useState<WorkstationOutput|null>(null)
  const [confirmDeleteKnast, setConfirmDeleteKnast] = useState(false)
  const [showDeleteInfo, setShowDeleteInfo] = useState(false)
  const [deleting, setDeleting] = useState(false)

  const closeDeleteModal = () => setShowDeleteModal(false)
  
  const deleteKnast = (slug: string) => {
    setDeleteError('')
    setDeleting(true)
    deleteWorkstation(slug).then(
      () => {
        setShowDeleteModal(false)
        setShowDeleteInfo(true)
      }
    ).catch(error => {
      setDeleteError(error.message)
    }).finally(() => {
      setDeleting(false)
    })
  }

  const onTrash= (knast: WorkstationOutput)=>{
    setSelectedKNAST(knast)
    setConfirmDeleteKnast(false)
    setShowDeleteModal(true)
  }

  const resyncKnast = (slug: string) => {
    resyncWorkstation(slug).then(
      () => {
        setShowResyncModal(false)
      }
    ).catch(error => {
      setResyncError(error.message)
    })
  }

  const resyncAllKnasts = () => {
    createWorkstationResyncAllWorkflow({
      "slugs": workstations?.map(w => w.slug) || []
    }).then(
      () => {
        setShowResyncAllModal(false)
      }
    ).catch(error => {
      setResyncAllError(error.message)
    })
  }

  const onResync = (knast: WorkstationOutput) => {
    setSelectedKNAST(knast)
    setShowResyncModal(true)
  }

  if (isLoading) return <div><LoaderSpinner />Laster...</div>
  if (error) return <ErrorStripe error={error} />
  return <div>
    <Modal open={showDeleteModal && !!selectedKNAST} onClose={closeDeleteModal} header={{ heading: "Slett KNAST" }}>
      <Modal.Body className="flex flex-col gap-4">
        <p>Denne operasjonen vil permanent slette KNAST eid av: {selectedKNAST?.displayName}</p>
        <Checkbox className='mt-2' checked={confirmDeleteKnast} onClick={() => setConfirmDeleteKnast(!confirmDeleteKnast)}>
          Jeg forstår at operasjonen vil slette KNAST eid av {selectedKNAST?.displayName}, og at dette ikke kan angres.
        </Checkbox>
        {deleteError && <Alert variant={'error'}>{deleteError}</Alert>}
        <div className="flex flex-row gap-3">
          <Button variant="secondary" onClick={closeDeleteModal} disabled={deleting}>
            Avbryt
          </Button>
          <Button onClick={()=>deleteKnast(selectedKNAST!!.slug)} disabled={!confirmDeleteKnast || deleting || !selectedKNAST?.slug}>Slett</Button>
        </div>
      </Modal.Body>
    </Modal>
    <Modal open={showResyncModal && !!selectedKNAST} onClose={() => setShowResyncModal(false)} header={{ heading: "Resynkroniser KNAST konfigurasjon" }}>
      <Modal.Body className="flex flex-col gap-4">
        <p>Denne operasjonen vil resynkronisere KNAST eid av: {selectedKNAST?.displayName}</p>
        {resyncError && <Alert variant={'error'}>{resyncError}</Alert>}
        <div className="flex flex-row gap-3">
          <Button variant="secondary" onClick={() => setShowResyncModal(false)}>
            Avbryt
          </Button>
          <Button onClick={()=>resyncKnast(selectedKNAST!!.slug)}>Resynkroniser</Button>
        </div>
      </Modal.Body>
    </Modal>
    <Modal open={showResyncAllModal} onClose={() => setShowResyncAllModal(false)} header={{ heading: "Resynkroniser alle Knast konfigurasjoner" }}>
      <Modal.Body className="flex flex-col gap-4">
        <p>Denne operasjonen vil resynkronisere alle Knast konfigurasjoner</p>
        {resyncAllError && <Alert variant={'error'}>{resyncAllError}</Alert>}
        <div className="flex flex-row gap-3">
          <Button variant="secondary" onClick={() => setShowResyncModal(false)}>
            Avbryt
          </Button>
          <Button onClick={resyncAllKnasts}>Resynkroniser alle</Button>
        </div>
      </Modal.Body>
    </Modal>
    <Head>
        <title>Admin verktøy - Knast administrasjon</title>
    </Head>
    <div className="flex flex-col gap-8 pt-4">
      <Heading size="large">Knast</Heading>
      <Button variant="primary" size="medium" className="w-1/4" onClick={() => setShowResyncAllModal(true)}>Resynkroniser alle Knaster</Button>
      {showDeleteInfo && <Alert variant={'info'}>Sletting av KNAST kan ta flere minutter å fullføre, men du kan forlate siden og komme tilbake senere</Alert>}
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
                {w.config?.reconciling ? <Loader title="oppdatering-pågår" size="large"/> : (
                  <div>
                    <Link href="#" onClick={()=>onResync(w)}><ArrowCirclepathIcon title="resync-knast" fontSize="2rem" /></Link>
                    <Link href="#" className="text-red-300" onClick={()=>onTrash(w)}><TrashIcon title="slett-knast" fontSize="2rem"/></Link>
                  </div>
                )
                }
              </Table.DataCell>
            </Table.Row>
          </>
        ))}
      </Table>
    </div>
    </div>
}

export default KnastPasge