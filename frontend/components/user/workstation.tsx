import { Button, Heading, Modal, Select } from "@navikt/ds-react"
import { ensureWorkstation, startWorkstation, stopWorkstation, useGetWorkstation } from "../../lib/rest/userData"
import LoaderSpinner from "../lib/spinner"
import { useState } from "react"

interface WorkstationStateProps {
    workstationData?: any
    handleOnStart: () => void
    handleOnStop: () => void
  }

const WorkstationState = ({workstationData, handleOnStart, handleOnStop}: WorkstationStateProps) => {
    return (
        <>
            {workstationData.state === 1 ?
                <div className="flex">
                    Starter workstation <LoaderSpinner/>
                </div> :
            workstationData.state === 2 ?          
                <div>
                    <Button variant='secondary' onClick={handleOnStop}>Stop</Button>
                </div> :
            workstationData.state === 3 ?
                <div>
                    Stopper workstation <LoaderSpinner/>
                </div> :
            <div>
                <Button onClick={handleOnStart}>Start</Button>
            </div>
            }
        </>
    )
}

export const Workstation = () => {
    const { workstation, loading } = useGetWorkstation()
    const [showCreateOrUpdateModal, setShowCreateOrUpdateModal] = useState(false)

    if (loading) return <LoaderSpinner />

    const handleOnCreateOrUpdate = (event: any) => {
        event.preventDefault()
        ensureWorkstation(
            {
                "machineType": event.target[0].value, 
                "containerImage": event.target[1].value
            }
        ).then(() => {
            setShowCreateOrUpdateModal(false)
        }).catch((e: any) => {
            console.log(e)
        })
    }

    const handleOnStart = () => {
        startWorkstation().then(() => {
            console.log("ok")
        }).catch((e: any) => {
            console.log(e)
        })
    }
    
    const handleOnStop = () => {
        stopWorkstation().then(() => {
            console.log("ok")
        }).catch((e: any) => {
            console.log(e)
        })
    }

    return (
    <div>
        <Modal
                open={showCreateOrUpdateModal}
                aria-label="Opprett eller oppdater workstation"
                onClose={() => setShowCreateOrUpdateModal(false)}
                className="max-w-full md:max-w-3xl px-8 h-[25rem]"
              >
                <Modal.Body className="h-full">
                <form
                    onSubmit={handleOnCreateOrUpdate}>
                  <div className="flex flex-col gap-8">
                    {workstation === null ?
                        <Heading level="1" size="medium">Opprett workstation</Heading> :
                        <Heading level="1" size="medium">Endre workstation</Heading>
                    }
                    <Select defaultValue={workstation?.config.machineType} label="Velg maskintype">
                        <option value={"n2d-standard-2"}>n2d-standard-2</option>
                        <option value={"n2d-standard-4"}>n2d-standard-4</option>
                        <option value={"n2d-standard-8"}>n2d-standard-8</option>
                        <option value={"n2d-standard-16"}>n2d-standard-16</option>
                        <option value={"n2d-standard-32"}>n2d-standard-32</option>
                    </Select>
                    <Select defaultValue={workstation?.config.containerImage} label="Velg containerImage">
                        <option value={"us-central1-docker.pkg.dev/cloud-workstations-images/predefined/code-oss:latest"}>VS code</option>
                        <option value={"us-central1-docker.pkg.dev/cloud-workstations-images/predefined/intellij-ultimate:latest"}>Intellij</option>
                        <option value={"us-central1-docker.pkg.dev/posit-images/cloud-workstations/workbench:latest"}>Posit</option>
                    </Select>
                    <div className="flex flex-row gap-3">
                        <Button variant="secondary" onClick={() => {setShowCreateOrUpdateModal(false)}}>
                            Avbryt
                        </Button>
                        {workstation === null ?
                            <Button type="submit">Opprett</Button> :
                            <Button type="submit">Endre</Button>
                        }
                    </div>
                  </div>
                </form>
                </Modal.Body>
        </Modal>
        {workstation !== null ? 
            <div className="flex flex-col">
                <Button onClick={() => {setShowCreateOrUpdateModal(true)}}>Endre</Button>
                <WorkstationState workstationData={workstation} handleOnStart={handleOnStart} handleOnStop={handleOnStop}/>
            </div> :
            <Button onClick={() => {setShowCreateOrUpdateModal(true)}}>Opprett</Button>
        }
    </div>
    )
}