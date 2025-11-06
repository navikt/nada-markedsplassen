import { ArrowLeftIcon, XMarkIcon } from "@navikt/aksel-icons";
import { Alert, Button, Link, Loader, Select, Switch, Table } from "@navikt/ds-react"
import React, { use, useEffect, useState } from "react"
import MachineTypeSelector from "./widgets/machineTypeSelector";
import useMachineTypeSelector from "./widgets/machineTypeSelector";
import { useCreateWorkstationJob, useWorkstationJobs, useWorkstationMine, useWorkstationOptions, useWorkstationSSHJob } from "./queries";
import { JobStateRunning, OnpremHostTypeTNS, WorkstationInput, WorkstationJob, WorkstationOnpremAllowList, WorkstationOptions } from "../../lib/rest/generatedDto";
import ContainerImageSelector from "./widgets/containerImageSelector";
import machineTypeSelector from "./widgets/machineTypeSelector";
import { configWorkstationSSH, createWorkstationJob } from "../../lib/rest/workstation";
import { set } from "lodash";
import { ColorAuxText } from "./designTokens";
import { useOnpremMapping } from "../onpremmapping/queries";

type SettingsFormProps = {
    knastInfo?: any;
    options?: WorkstationOptions;
    onSave: () => void;
    onCancel: () => void;
    onConfigureDatasources: () => void;
    onConfigureInternet: () => void;
}

export const SettingsForm = ({ knastInfo, options, onSave, onCancel, onConfigureDatasources, onConfigureInternet }: SettingsFormProps) => {
    const [ssh, setSSH] = React.useState(knastInfo?.allowSSH || false);
    const { MachineTypeSelector, selectedMachineType } = useMachineTypeSelector(knastInfo.config?.machineType || options?.machineTypes?.find(type => type !== undefined)?.machineType || "")
    const [selectedContainerImage, setSelectedContainerImage] = useState<string>(knastInfo?.config?.image || options?.containerImages?.find(image => image !== undefined)?.image || "");
    const createWorkstationJob = useCreateWorkstationJob();
    const [showRestartAlert, setShowRestartAlert] = useState(false);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const workstationJobs = useWorkstationJobs();
    const workstationSSHJobs = useWorkstationSSHJob();
    const hasRunningJobs = !!workstationJobs.data?.jobs?.filter((job): job is WorkstationJob => 
    job !== undefined && job.state === JobStateRunning).length || !!workstationSSHJobs.data && workstationSSHJobs.data.state === JobStateRunning;
    const workstationOnpremMapping = knastInfo?.workstationOnpremMapping as WorkstationOnpremAllowList || {};
    const onpremMapping = useOnpremMapping();
    const hasDVHSource = workstationOnpremMapping.hosts.some(h =>
        Object.entries(onpremMapping.data?.hosts ?? {}).find(([type, _]) => type === OnpremHostTypeTNS)?.[1].some(host => host?.Host === h));
    const needCreateWorkstationJob = (knastInfo.config?.machineType !== selectedMachineType) || (knastInfo.config?.image !== selectedContainerImage);
    const needUpdateSSH = knastInfo.allowSSH !== ssh;
    const handleSave = async () => {

        const input: WorkstationInput = {
            machineType: selectedMachineType,
            containerImage: selectedContainerImage ?? "",
        };

        try {
            setSaving(true);
            if (knastInfo.operationalStatus === "started" || knastInfo.operationalStatus === "starting") {
                setShowRestartAlert(true);
            }           
            setError(null);

            if (needCreateWorkstationJob){
                await createWorkstationJob.mutateAsync(input)
            }
            
            if (needUpdateSSH){
                configWorkstationSSH(ssh);
            }

            const t= setTimeout(() => {
                setSaving(false);
                clearTimeout(t);
            }, 5000);
        } catch (error) {
            setSaving(false);
            setShowRestartAlert(false);
            setError("Noe gikk galt ved lagring av innstillinger. Vennligst prøv igjen.");
        }

        onSave();
    }

    return (
        <div className="max-w-200 border-blue-100 border rounded p-4">
            <Table>
                <Table.Header>
                    <Table.Row>
                        <Table.HeaderCell colSpan={2} scope="col">
                        <div className="flex flex-row justify-between items-center">

                            <h3>Innstillinger</h3>                            <Button variant="tertiary" size="small" onClick={onCancel}>
                                <XMarkIcon width={20} height={20} />
                            </Button>
                        </div>
                        </Table.HeaderCell>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    <Table.Row>
                        <Table.HeaderCell scope="row">Utviklingsmiljø / Image</Table.HeaderCell>
                        <Table.DataCell>
                            <ContainerImageSelector initialContainerImage={selectedContainerImage} handleSetContainerImage={setSelectedContainerImage} />
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.HeaderCell scope="row">Maskintype</Table.HeaderCell>
                        <Table.DataCell>
                            <MachineTypeSelector /> 
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.DataCell colSpan={2}>
                            <Switch checked={ssh} onChange={() => setSSH(!ssh)} >Local dev (SSH)</Switch>
                            {ssh && <div className="mt-2 flex flex-rol">For instruksjoner og restriksjoner for lokal utvikling, se <Link href="#" className="flex flex-rol ml-2">Dokumentasjon</Link></div>}
                            {hasDVHSource && <div className="mt-2 italic">Av sikkerhetshensyn kan ikke aktiver SSH (lokal IDE-tilgang) når du bruker DVH-kilden.</div>}
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.DataCell colSpan={2}>
                            <Link href="#" className="flex flex-rol ml-2" onClick={onConfigureDatasources}>Konfigure NAV datakilder</Link>
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.DataCell colSpan={2}>
                            <Link href="#" className="flex flex-rol ml-2" onClick={onConfigureInternet}>Konfigure internettåpninger</Link>
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.DataCell colSpan={2}>
                            <div className="flex flex-rol gap-4">
                                <Button variant="primary" disabled={saving || hasRunningJobs} onClick={handleSave}>Lagre{(saving || hasRunningJobs) && <Loader className="ml-2"/>}</Button>
                                <div>
                                    <Button variant="secondary" className="ml-6" onClick={onCancel}>Tilbake</Button>
                                </div>

                            </div>
                        </Table.DataCell>
                    </Table.Row>
                </Table.Body>
            </Table>
            {showRestartAlert && <div className="mt-4 italic">Vennligst start arbeidsstasjonen på nytt for at endringene skal tre i kraft.</div>}
            {error && <div className="mt-4 text-red-600">{error}</div>}
        </div>
    )
}
