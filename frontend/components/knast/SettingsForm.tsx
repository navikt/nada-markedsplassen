import { ArrowLeftIcon } from "@navikt/aksel-icons";
import { Button, Link, Select, Switch, Table } from "@navikt/ds-react"
import React from "react"

type SettingsFormProps = {
    onSave?: () => void;
    onCancel?: () => void;
}

export const SettingsForm = ({ onSave, onCancel }: SettingsFormProps) => {

    const [selectedMachine, setSelectedMachine] = React.useState("n2ds2")
    const [ssh, setSSH] = React.useState(true)

    const machineDetails = {
        "n2ds2": "2vCPU, 8GB RAM, kr 22/day",
        "n2ds4": "4vCPU, 16GB RAM, kr 45/day",
        "n2ds8": "8vCPU, 32GB RAM, kr 90/day",
    }

    return (

        <div className="max-w-[35rem] border-blue-100 border rounded p-4">
            <Table>
                <Table.Header>
                    <Table.Row>
                        <Table.HeaderCell colSpan={2} scope="col">
                            Settings
                        </Table.HeaderCell>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    <Table.Row>
                        <Table.HeaderCell scope="row">Environment</Table.HeaderCell>
                        <Table.DataCell>
                            <Select label="">
                                <option value="vscode">VS Code</option>
                                <option value="jupyter">Jupyter Notebook</option>
                                <option value="steam">Steam</option>
                                <option value="xbox">XBox with Gamepass</option>
                            </Select>
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.HeaderCell scope="row">Machine Type</Table.HeaderCell>
                        <Table.DataCell>
                            <Select label="" value={selectedMachine} onChange={(e) => setSelectedMachine(e.target.value)}>
                                <option value="n2ds2">n2d-standard-2</option>
                                <option value="n2ds4">n2d-standard-4</option>
                                <option value="n2ds8">n2d-standard-8</option>
                            </Select>
                            <div className="mt-4">{machineDetails[selectedMachine as keyof typeof machineDetails]}</div>
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.DataCell colSpan={2}>
                            <Switch checked={ssh} onChange={() => setSSH(!ssh)}>Local dev (SSH)</Switch>
                            {ssh && <div className="mt-2 flex flex-rol">For instructions and restrictions on Local dev, see <Link href="#" className="flex flex-rol ml-2">Documentation</Link></div>}
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.DataCell colSpan={2}>
                            <Link href="#" className="flex flex-rol ml-2">Configure access to on-prem data sources</Link>
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.DataCell colSpan={2}>
                            <Link href="#" className="flex flex-rol ml-2">Configure internet gateway</Link>
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.DataCell colSpan={2}>
                            <div className="flex flex-rol gap-4">
                                <Button variant="primary" onClick={onSave}>Save</Button>
                                <div>
                                    <Button variant="secondary" className="ml-6" onClick={onCancel}>Cancel</Button>
                                </div>

                            </div>
                        </Table.DataCell>
                    </Table.Row>
                </Table.Body>
            </Table>
        </div>
    )
}
