import { useRouter } from "next/router"
import { useState } from "react"
import { deleteDataproduct } from "../../lib/rest/dataproducts"
import amplitudeLog from "../../lib/amplitude"
import { Alert, Button, Checkbox, Dropdown, Modal } from "@navikt/ds-react"
import { MenuElipsisHorizontalCircleIcon } from "@navikt/aksel-icons"
import DeleteModal from "../lib/deleteModal"

interface IDataproductOwnerMenuProps {
    dataproduct: any
    className?: string
}

const DataproductOwnerMenu = ({
    dataproduct,
    className,
}: IDataproductOwnerMenuProps) => {
    const [showDeleteEmpty, setShowDeleteEmpty] = useState(false)
    const [showDeleteNonEmpty, setShowDeleteNonEmpty] = useState(false)
    const [confirmDeleteNonEmpty, setConfirmDeleteNonEmpty] = useState(false)
    const [deleteError, setDeleteError] = useState('')
    const router = useRouter()

    const onDelete = async () => {
        if (!dataproduct) return
        deleteDataproduct(dataproduct.id).then(() => {
            amplitudeLog('slett dataprodukt', { name: dataproduct.name })
            router.push('/')
        }).catch(error => {
            amplitudeLog('slett dataprodukt feilet', { name: dataproduct.name })
            setDeleteError(error)
        })
    }

    const showDeleteModal = (empty: boolean)=>{
        if (empty){
            setShowDeleteEmpty(true)
        }else{
            setShowDeleteNonEmpty(true)
        }
    }

    const closeDeleteModal = ()=>{
        setShowDeleteEmpty(false)
        setShowDeleteNonEmpty(false)
        setConfirmDeleteNonEmpty(false)
    }

    return (
        <div className={className}>
            <Dropdown>
                <Button
                    as={Dropdown.Toggle}
                    className="p-0 w-8 h-8"
                    variant="tertiary"
                >
                    <MenuElipsisHorizontalCircleIcon className="w-6 h-6" />
                </Button>
                <Dropdown.Menu>
                    <Dropdown.Menu.GroupedList>
                        <Dropdown.Menu.GroupedList.Item onClick={() => router.push(`/dataproduct/${dataproduct.id}/${dataproduct.slug}/edit`)}>
                            Endre dataprodukt
                        </Dropdown.Menu.GroupedList.Item>
                        <Dropdown.Menu.GroupedList.Item onClick={() => 
                            showDeleteModal(!dataproduct?.datasets?.length)
                        }>
                            Slett dataprodukt
                        </Dropdown.Menu.GroupedList.Item>
                    </Dropdown.Menu.GroupedList>
                </Dropdown.Menu>
            </Dropdown>
            <DeleteModal
                name={dataproduct?.name}
                resource="dataprodukt"
                error={deleteError}
                open={showDeleteEmpty}
                onCancel={closeDeleteModal}
                onConfirm={onDelete}
            ></DeleteModal>
            <Modal open={showDeleteNonEmpty} onClose={closeDeleteModal} header={{ heading: "Slett" }}>
                <Modal.Body className="flex flex-col gap-4">
                    <p>Dataproduktet <strong>{dataproduct.name}</strong> er ikke tomt. Sletting av dataproduktet vil også slette følgende datasett:</p>
                    <ul>
                        {dataproduct.datasets.map((dataset: any) => (
                            <li key={dataset.id}><strong>{dataset.name}</strong></li>
                        ))}
                    </ul>
                    <Checkbox className='mt-2' checked={confirmDeleteNonEmpty} onClick={()=> setConfirmDeleteNonEmpty(!confirmDeleteNonEmpty)}>
                        Jeg forstår at operasjonen vil slette dataproduktet samt datasettene ovenfor, og at dette ikke kan angres.
                    </Checkbox>
                    <div className="flex flex-row gap-3">
                        <Button variant="secondary" onClick={closeDeleteModal}>
                            Avbryt
                        </Button>
                        <Button onClick={onDelete} disabled={!confirmDeleteNonEmpty}>Slett</Button>
                    </div>
                    {deleteError && <Alert variant={'error'}>{deleteError}</Alert>}
                </Modal.Body>
            </Modal>

        </div>
    )
}

export default DataproductOwnerMenu