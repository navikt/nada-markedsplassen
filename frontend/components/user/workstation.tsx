import {Alert, Button, Heading, Select} from "@navikt/ds-react"
import {
    ensureWorkstation,
    startWorkstation,
    stopWorkstation,
    useGetWorkstation,
    useGetWorkstationOptions
} from "../../lib/rest/userData"
import LoaderSpinner from "../lib/spinner"
import {Fragment, useState} from "react";
import {
    FirewallTag,
    Workstation_STATE_RUNNING,
    Workstation_STATE_STARTING, Workstation_STATE_STOPPED,
    Workstation_STATE_STOPPING, WorkstationContainer, WorkstationMachineType
} from "../../lib/rest/generatedDto";
import TagsSelector from "../lib/tagsSelector";
import TagPill from "../lib/tagPill";

interface WorkstationStateProps {
    workstationData?: any
    handleOnStart: () => void
    handleOnStop: () => void
}

const WorkstationState = ({workstationData, handleOnStart, handleOnStop}: WorkstationStateProps) => {
    if (workstationData === null) {
        return
        // return <Alert variant={'warning'}>No running workstation</Alert>
    }

    switch (workstationData.state) {
        case Workstation_STATE_STARTING:
            return (
                <div className="flex">
                    Starter workstation <LoaderSpinner/>
                </div>
            )
        case Workstation_STATE_RUNNING:
            return (
                <div>
                    <Button variant='secondary' onClick={handleOnStop}>Stop</Button>
                </div>
            )
        case Workstation_STATE_STOPPING:
            return (
                <div>
                    Stopper workstation <LoaderSpinner/>
                </div>
            )
        case Workstation_STATE_STOPPED:
            return (
                <div>
                    <Button onClick={handleOnStart}>Start</Button>
                </div>
            )
    }
}

export const Workstation = () => {
    const {workstation, loading} = useGetWorkstation()
    const {workstationOptions, loadingOptions} = useGetWorkstationOptions()
    const [selectedFirewallTags, setSelectedFirewallTags] = useState<string[]>([])

    if (loading) return <LoaderSpinner/>
    if (loadingOptions) return <LoaderSpinner/>

    const handleOnCreateOrUpdate = (event: any) => {
        event.preventDefault()
        ensureWorkstation(
            {
                "machineType": event.target[0].value,
                "containerImage": event.target[1].value,
                "firewallTags": selectedFirewallTags,
            }
        ).then(() => {
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

    const handleFirewallTagChange = (event: any) => {
        const options = event.target.options
        const selectedTags: string[] = []
        for (const option of options) {
            if (option.selected) {
                selectedTags.push(option.value)
            }
        }
        setSelectedFirewallTags(selectedTags)
    }

    return (
        <div>
            <form
                onSubmit={handleOnCreateOrUpdate}>
                <div className="flex flex-col gap-8">
                    {workstation === null ?
                        <Heading level="1" size="medium">Opprett workstation</Heading> :
                        <Heading level="1" size="medium">Endre workstation</Heading>
                    }
                    <Select defaultValue={workstation?.config?.machineType} label="Velg maskintype">
                        {workstationOptions?.machineTypes.map((type: WorkstationMachineType | undefined) => (
                            type ? <option key={type.machineType} value={type.machineType}>{type.machineType} (vCPU: {type.vCPU}, memoryGB: {type.memoryGB})</option> :
                                "Could not load machine type"
                        ))}                    </Select>
                    <Select defaultValue={workstation?.config?.image} label="Velg containerImage">
                        {workstationOptions?.containerImages.map((image: WorkstationContainer | undefined) => (
                            image ? <option key={image.image} value={image.image}>{image.description}</option> :
                                "Could not load container image"
                        ))}                    </Select>
                    <Select multiple value={selectedFirewallTags} onChange={handleFirewallTagChange} label="Velg firewall tags" >
                        {workstationOptions?.firewallTags.map((tag: FirewallTag | undefined) => (
                            <option key={tag?.name} value={tag?.secureTag}>
                                {tag?.name}
                            </option>
                        ))}
                    </Select>
                    <div className="flex flex-row gap-3">
                        <Button variant="secondary" onClick={() => {
                        }}>
                            Avbryt
                        </Button>
                        {workstation === null ?
                            <Button type="submit">Opprett</Button> :
                            <Button type="submit">Endre</Button>
                        }
                    </div>
                </div>
            </form>
            <div className="flex flex-col">
                <Button onClick={() => {
                }}>Endre</Button>
                <WorkstationState workstationData={workstation} handleOnStart={handleOnStart}
                                  handleOnStop={handleOnStop}/>
            </div>
        </div>
    )
}
