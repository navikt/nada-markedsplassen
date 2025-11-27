import { XMarkIcon } from "@navikt/aksel-icons";
import { Alert, Button, Checkbox, CheckboxGroup, Loader, Select, Table, UNSAFE_Combobox } from "@navikt/ds-react";
import React, { use, useState } from "react";
import { useUpdateWorkstationOnpremMapping, useWorkstationOnpremMapping } from "./queries";
import { useOnpremMapping } from "../onpremmapping/queries";
import { ColorAuxText, ColorDisabled, ColorFailed } from "./designTokens";
import { sub } from "date-fns";

type DatasourcesFormProps = {
    knastInfo: any;
    onCancel: () => void;
}

export const DatasourcesForm = ({ knastInfo, onCancel}: DatasourcesFormProps) => {
    const onpremMapping = useOnpremMapping()
    const [selectedOnpremMapping, setSelectedOnpremMapping] = useState<string[]>(
        knastInfo.workstationOnpremMapping ? knastInfo.workstationOnpremMapping.map((h: any) => h.host) : [])
    const [expandedGroups, setExpandedGroups] = useState<string[]>([]);
    const [backendError, setBackendError] = useState<string | undefined>(undefined);
    const updateOnpremMapping = useUpdateWorkstationOnpremMapping();


    const isExpanded = (group: string) => expandedGroups.includes(group) || selectedOnpremMapping.some(selectedHost =>
        onpremMapping.data?.hosts?.[group]?.some(host => host?.Host === selectedHost));

    const settingsChange = () => {
        return selectedOnpremMapping.some(it => !knastInfo.workstationOnpremMapping?.hosts?.includes(it)) ||
            (knastInfo.workstationOnpremMapping?.hosts?.some((it: string) => !selectedOnpremMapping.includes(it)));
    }


    const submitSettings = async (selectedHosts: string[]) => {
        setBackendError(undefined);
        try {
            await updateOnpremMapping.mutateAsync({ hosts: selectedHosts  });
        } catch (error: any) {
            setBackendError("Feil ved lagring av datakilder");
        }
    }


    const updateSelectedOnpremMapping = (newSelection: string[]) => {
        setSelectedOnpremMapping(newSelection);
        submitSettings(newSelection);
    }
    return (
        <div className="max-w-220 border-blue-100 border rounded p-4">
            <Table>
                <Table.Header>
                    <Table.Row>
                        <Table.HeaderCell colSpan={2} scope="col">
                            <div className="flex flex-row justify-between items-center">
                                <h3>
                                    Nav datakilder
                                </h3>
                                <Button variant="tertiary" size="small" onClick={onCancel}>
                                    <XMarkIcon width={20} height={20} />
                                </Button>
                            </div>
                        </Table.HeaderCell>
                    </Table.Row>
                </Table.Header>
                {
                    onpremMapping.isLoading ? <Loader /> : onpremMapping.isError || !onpremMapping?.data?.hosts ? <div className="pt-4"><Alert variant="error">Feil ved lasting av datakilder</Alert></div> :
                        <Table.Body>
                            {
                                Object.entries(onpremMapping.data?.hosts!!).map(it =>
                                    <Table.Row key={it[0]}>
                                        <Table.HeaderCell scope="row" className="align-top">
                                            <Checkbox checked={isExpanded(it[0])} onChange={e => {
                                                if (e.target.checked) {
                                                    setExpandedGroups([...expandedGroups, it[0]])
                                                } else {
                                                    setExpandedGroups(expandedGroups.filter(h => h !== it[0]))
                                                    updateSelectedOnpremMapping(selectedOnpremMapping.filter(h =>
                                                        !it[1].some(host => host?.Host === h)))
                                                }
                                            }}>
                                                {it[0] === "tns" ? "DVH" : it[0].toUpperCase()}
                                            </Checkbox>
                                        </Table.HeaderCell>
                                        <Table.DataCell>
                                            {   it[0]==="tns" && knastInfo?.allowSSH &&                                               <div className="italic mb-2" style={{
                                                    color: ColorFailed
                                                }}>Av sikkerhetshensyn kan ikke Knast åpne DVH-kilder når SSH (lokal IDE-tilgang) er aktivert.</div>}
                                                <div className={it[0] === "tns" ? "flex flex-col gap-2" : "flex flex-wrap gap-2"}>{
                                                    isExpanded(it[0]) && it[1].map((host: any, index: number) => (
                                                        <div key={index}>
                                                            <Checkbox size="medium" checked={selectedOnpremMapping.includes(host?.Host)} onChange={e => {
                                                                if (e.target.checked) {
                                                                    updateSelectedOnpremMapping([...selectedOnpremMapping, host?.Host])
                                                                } else {
                                                                    updateSelectedOnpremMapping(selectedOnpremMapping.filter(h => h !== host?.Host))
                                                                }
                                                            }}>
                                                                {host.Name} {host.Name !== host.Host ? "(" + host.Host + ")" : ""}
                                                            </Checkbox>
                                                            {it[0] === "tns" && <div className="text-sm text-gray-600 ml-6">{host?.Description}</div>}
                                                        </div>
                                                    ))}
                                                </div>
                                        </Table.DataCell>
                                    </Table.Row>
                                )
                            }
                        </Table.Body>
                }
            </Table>
            <div>
                <div className="flex flex-row pt-6">
                    <Button variant="secondary" onClick={onCancel}>Tilbake</Button>
                </div>
                {backendError && <div className="pt-4">
                    <Alert variant="error">{backendError}</Alert>
                </div>}
            </div>

        </div>
    )
}