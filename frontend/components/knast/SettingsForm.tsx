import { ExclamationmarkTriangleIcon } from "@navikt/aksel-icons";
import { Alert, Button, Link, Loader, Switch, Table } from "@navikt/ds-react";
import React, { useEffect, useState } from "react";
import { JobStateRunning, WorkstationInput, WorkstationJob, WorkstationOptions } from "../../lib/rest/generatedDto";
import { configWorkstationSSH } from "../../lib/rest/workstation";
import { ColorAuxText } from "./designTokens";
import { useCreateWorkstationJob, useWorkstationJobs, useWorkstationSSHJob } from "./queries";
import { useAutoCloseAlert } from "./widgets/autoCloseAlert";
import ContainerImageSelector from "./widgets/containerImageSelector";
import useMachineTypeSelector from "./widgets/machineTypeSelector";

type SettingsFormProps = {
    knastInfo?: any;
    options?: WorkstationOptions;
}

export const SettingsForm = ({ knastInfo, options}: SettingsFormProps) => {
    const [ssh, setSSH] = React.useState(knastInfo?.allowSSH || false);
    const { MachineTypeSelector, selectedMachineType } = useMachineTypeSelector(knastInfo?.config?.machineType || options?.machineTypes?.find(type => type !== undefined)?.machineType || "", options)
    const [selectedContainerImage, setSelectedContainerImage] = useState<string>(knastInfo?.config?.image || options?.containerImages?.find(image => image !== undefined)?.image || "");
    const createWorkstationJob = useCreateWorkstationJob();
    const [showRestartAlert, setShowRestartAlert] = useState(false);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const workstationJobs = useWorkstationJobs();
    const workstationSSHJobs = useWorkstationSSHJob();
    const hasRunningJobs = !!workstationJobs.data?.jobs?.filter((job): job is WorkstationJob =>
        job !== undefined && job.state === JobStateRunning).length || !!workstationSSHJobs.data && workstationSSHJobs.data.state === JobStateRunning;
    const workstationOnpremMapping = knastInfo?.workstationOnpremMapping || [];
    const hasDVHSource = workstationOnpremMapping.some((it: any) => it.isDVHSource);
    const needCreateWorkstationJob = selectedMachineType && (knastInfo?.config?.machineType !== selectedMachineType) 
    || selectedContainerImage &&(knastInfo?.config?.image !== selectedContainerImage);
    const needUpdateSSH = knastInfo?.allowSSH !== ssh;
    const {showAlert, AutoHideAlert} = useAutoCloseAlert(5000);

    useEffect(() => {
        if (!hasRunningJobs && saving) {
            showAlert();
            setSaving(false);
        }        
    }, [hasRunningJobs]);

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

            if (needCreateWorkstationJob) {
                await createWorkstationJob.mutateAsync(input)
            }

            if (needUpdateSSH) {
                configWorkstationSSH(ssh);
            }
        } catch (error) {
            setSaving(false);
            setShowRestartAlert(false);
            setError("Noe gikk galt ved lagring av innstillinger. Vennligst prøv igjen.");
        }
    }

    return (
        <div className="w-180 border-gray-300 border-l pl-6">
            <Table>
                <Table.Body>
                    <Table.Row>
                        <Table.HeaderCell scope="row" className="align-text-top"><div className="mt-4">Image</div></Table.HeaderCell>
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
                            <div className="flex flex-row gap-2 items-center"><Switch checked={ssh} onChange={() => setSSH(!ssh)} >Local dev (SSH)</Switch>                            {hasDVHSource && ssh && <p className="flex flex-row mt-1 items-center"><ExclamationmarkTriangleIcon /><p className="text-sm italic" style={{ color: ColorAuxText }}> DVH kilder er ikke tiltat nå SSH er aktivert</p></p>}</div>

                            {ssh && <div className="mt-2 flex flex-rol">For instruksjoner og restriksjoner for lokal utvikling, se <Link href="#" className="flex flex-rol ml-2">Dokumentasjon</Link></div>}
                        </Table.DataCell>
                    </Table.Row>
                </Table.Body>
            </Table>
            {(needCreateWorkstationJob || needUpdateSSH) && !saving ? <Alert className="mt-4" variant="warning">Endringer ikke lagret</Alert> : null}
            <div className="flex mt-6 flex-rol gap-4">
                <Button variant="primary" disabled={saving || hasRunningJobs || !needCreateWorkstationJob && !needUpdateSSH} onClick={handleSave}>Lagre endringer{(saving || hasRunningJobs) && <Loader className="ml-2" />}</Button>
            </div>
            {showRestartAlert && <div className="mt-4 italic">Vennligst start knast på nytt for at endringene skal tre i kraft.</div>}
            {error && <div className="mt-4 text-red-600">{error}</div>}
            <AutoHideAlert variant="success" className="mt-4">Endringer lagret.</AutoHideAlert>
        </div>
    )
}
