import { Alert, Button, Heading, Select, UNSAFE_Combobox, Textarea, Label, Link } from "@navikt/ds-react"
import {
    ensureWorkstation,
    startWorkstation,
    stopWorkstation,
    useGetWorkstation,
    useGetWorkstationOptions
} from "../../lib/rest/userData"
import LoaderSpinner from "../lib/spinner"
import { Fragment, useState } from "react";
import {
    FirewallTag,
    Workstation_STATE_RUNNING,
    Workstation_STATE_STARTING, Workstation_STATE_STOPPED,
    Workstation_STATE_STOPPING, WorkstationContainer, WorkstationMachineType
} from "../../lib/rest/generatedDto";
import TagsSelector from "../lib/tagsSelector";
import TagPill from "../lib/tagPill";
import { Header } from "@navikt/ds-react-internal";

interface WorkstationStateProps {
    workstationData?: any
    handleOnStart: () => void
    handleOnStop: () => void
}

const WorkstationState = ({ workstationData, handleOnStart, handleOnStop }: WorkstationStateProps) => {
    
    const startStopButtons = (startButtonDisabled: boolean, stopButtonDisabled:boolean) => {
        return (
            <div className="flex gap-2">
                <Button disabled={startButtonDisabled} onClick={handleOnStart}>Start</Button>
                <Button disabled={stopButtonDisabled} onClick={handleOnStop}>Stopp</Button>
            </div>
        )
    }
    
    if (workstationData === null) {
        return (
            <div className="flex flex-col gap-4 pt-4">
            <Alert variant={'warning'}>Du har ikke opprettet en arbeidsstasjon</Alert>
            {startStopButtons(true,true)}
            </div>
        )
    }

    switch (workstationData.state) {
        case Workstation_STATE_STARTING:
            return (
                <div className="flex">
                    Starter arbeidsstasjon <LoaderSpinner />
                    { startStopButtons(true,true) }
                </div>
            )
        case Workstation_STATE_RUNNING:
            return (
                <div>
                    { startStopButtons(true,false) }
                </div>
            )
        case Workstation_STATE_STOPPING:
            return (
                <div>
                    Stopper arbeidsstasjon <LoaderSpinner />
                    { startStopButtons(true,true) }
                </div>
            )
        case Workstation_STATE_STOPPED:
            return (
                <div>
                    { startStopButtons(false,true) }
                </div>
            )
    }
}

export const Workstation = () => {
    const { workstation, loading } = useGetWorkstation()
    const { workstationOptions, loadingOptions } = useGetWorkstationOptions()
    const [selectedFirewallTags, setSelectedFirewallTags] = useState(new Set<string>())

    if (loading) return <LoaderSpinner />
    if (loadingOptions) return <LoaderSpinner />

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


    const handleFirewallTagChange = (tagValue: string, isSelected: boolean) => {
        if (isSelected) {
            setSelectedFirewallTags(new Set(selectedFirewallTags.add(tagValue)))
            return
        }
        selectedFirewallTags.delete(tagValue)

        setSelectedFirewallTags(new Set(selectedFirewallTags))
    }

    return (
        <div className="flex flex-col gap-8">
            <p>Her kan du opprette og gjøre endringer på din personlige arbeidsstasjon</p>
            <div className="flex">
            <form className="basis-1/2 border-x p-4"
                onSubmit={handleOnCreateOrUpdate}>
                <div className="flex flex-col gap-8">
                    {workstation === null ?
                        <Heading level="1" size="medium">Opprett arbeidsstasjon</Heading> :
                        <Heading level="1" size="medium">Endre arbeidsstasjon</Heading>
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
                    <UNSAFE_Combobox
                        label="Velg hvilke onprem-kilder du trenger åpninger mot"
                        options={workstationOptions ? workstationOptions.firewallTags?.map((o: FirewallTag | undefined) => (o ? {
                            label: `${o?.name}`,
                            value: o?.secureTag,
                        } : { label: "Could not load firewall tag", value: "Could not load firewall tag" })) : []}
                        isMultiSelect
                        onToggleSelected={handleFirewallTagChange}
                    />
                    <div className="flex gap-2 flex-col">
                        <Label>Oppgi hvilke internett-URL-er du vil åpne mot</Label>
                        <p className="pt-0">Du kan legge til opptil 2500 oppføringer i en URL-liste. Hver oppføring må stå på en egen linje uten mellomrom eller skilletegn. Oppføringer kan være kun domenenavn (som matcher alle stier) eller inkludere en sti-komponent. <Link href="https://cloud.google.com/secure-web-proxy/docs/url-list-syntax-reference">Les mer om syntax her.</Link></p>

                        <Textarea size="medium" maxRows={2500} hideLabel label="Hvilke URL-er vil du åpne mot" resize />
                    </div>
                    <div className="flex flex-row gap-3">
                        {workstation === null ?
                            <Button type="submit">Opprett</Button> :
                            <Button type="submit">Endre</Button>
                        }
                    </div>
                </div>
            </form>
            <div className="flex flex-col gap-4 basis-1/2">
                <div className="border-1 p-4">
                <Heading level="1" size="medium">Status for arbeidsstasjon</Heading>
                <WorkstationState workstationData={workstation} handleOnStart={handleOnStart} handleOnStop={handleOnStop} />
                </div>
            </div>
        </div>
            </div>
    )
}
