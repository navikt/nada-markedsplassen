import { Button, Checkbox, CheckboxGroup, Select, Table, UNSAFE_Combobox } from "@navikt/ds-react";
import React from "react";

type DatasourcesFormProps = {
    onSave?: () => void;
    onCancel?: () => void;
}

export const DatasourcesForm = ({ onSave, onCancel }: DatasourcesFormProps) => {
    const [accessDVH, setAccessDVH] = React.useState(false)
    const [accessPostgres, setAccessPostgres] = React.useState(false)
    const [accessHttps, setAccessHttps] = React.useState(false)
    const [accessSFTP, setAccessSFTP] = React.useState(false)
    const [accessCloudSQL, setAccessCloudSQL] = React.useState(false)
    const [accessInformatica, setAccessInformatica] = React.useState(false)
    const [accessOracle, setAccessOracle] = React.useState(false)
    const [accessSMTP, setAccessSMTP] = React.useState(false)

    return (
        <div className="max-w-[35rem] border-blue-100 border rounded p-4">
            <Table>
                <Table.Header>
                    <Table.Row>
                        <Table.HeaderCell colSpan={2} scope="col">
                            Configure on-prem data sources
                        </Table.HeaderCell>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    <Table.Row>
                        <Table.HeaderCell scope="row">
                            <Checkbox checked={accessDVH} onChange={() => setAccessDVH(!accessDVH)}>Datavarehus</Checkbox>
                        </Table.HeaderCell>
                        <Table.DataCell>
                            {
                                accessDVH && <CheckboxGroup legend={undefined}>
                                    <Checkbox value={"dvhi"} description="Live datamart mot DVH-P som benyttes til konsum av dataobjekter gjort tilgjengelig i IAPP-sonen.">DVH-I (a01dbfl036.adeo.no)</Checkbox>
                                    <Checkbox value={"dvhq"} description="Produksjonsdatabasen til Nav-datavarehus. Den benyttes til ETL/lagring og konsum av dataobjekter i fagsystemsonen.">DVH-P (dm08-scan.adeo.no)</Checkbox>
                                    <Checkbox value={"dvhr"} description="Kopi av produksjon (synkroniseres hver natt) som benyttes til test og utvikling.">DVH-R (dmv34-scan.adeo.no)</Checkbox>
                                    <Checkbox value={"dvhu"} description="Kopi av produksjon (manuelt oppdatert) som benyttes til utvikling.">DVH-U (dmv34-scan.adeo.no)</Checkbox>
                                    <Checkbox value={"dvhq"} description="Kopi av produksjon (manuelt oppdatert) som benyttes til test av deploy.">DDVH-Q (dmv38-scan.adeo.no)</Checkbox>
                                </CheckboxGroup>
                            }
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.HeaderCell scope="row">
                            <Checkbox checked={accessPostgres} onChange={() => setAccessPostgres(!accessPostgres)}>Postgres</Checkbox>
                        </Table.HeaderCell>
                        <Table.DataCell>
                            {
                                accessPostgres && <UNSAFE_Combobox
                                    isMultiSelect
                                    hideLabel
                                    options={[{ label: "pg01.adeo.no", value: "pg01.adeo.no" }, { label: "pg02.adeo.no", value: "pg02.adeo.no" }, { label: "pg03.adeo.no", value: "pg03.adeo.no" }]} label={undefined} />
                            }
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.HeaderCell scope="row">
                            <Checkbox checked={accessHttps} onChange={() => setAccessHttps(!accessHttps)}>HTTPS</Checkbox>
                        </Table.HeaderCell>
                        <Table.DataCell>
                            {
                                accessHttps && <CheckboxGroup legend={undefined}>
                                    <Checkbox value={"https1"} description="Firewall rule for host *.dev.intern.nav.no">dev.intern.nav.no</Checkbox>
                                    <Checkbox value={"https2"} description="Firewall rule for host *.intern.dev.nav.no">intern.dev.nav.no</Checkbox>
                                    <Checkbox value={"https3"} description="Firewall rule for host *.intern.nav.no">intern.nav.no</Checkbox>
                                    <Checkbox value={"https4"} description="Firewall rule for host *.nais.adeo.no">nais.adeo.no</Checkbox>
                                </CheckboxGroup>
                            }
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.HeaderCell scope="row">
                            <Checkbox checked={accessSFTP} onChange={() => setAccessSFTP(!accessSFTP)}>SFTP</Checkbox>
                        </Table.HeaderCell>
                        <Table.DataCell>
                            {
                                accessSFTP && <UNSAFE_Combobox
                                    isMultiSelect
                                    hideLabel
                                    options={[{ label: "sftp1.adeo.no", value: "sftp1.adeo.no" }, { label: "sftp2.adeo.no", value: "sftp2.adeo.no" }, { label: "sftp3.adeo.no", value: "sftp3.adeo.no" }]} label={undefined} />
                            }
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.HeaderCell scope="row">
                            <Checkbox checked={accessCloudSQL} onChange={() => setAccessCloudSQL(!accessCloudSQL)}>Cloud SQL</Checkbox>
                        </Table.HeaderCell>
                        <Table.DataCell>
                            {
                                accessCloudSQL && <UNSAFE_Combobox
                                    isMultiSelect
                                    hideLabel
                                    options={[{ label: "cloudsql1.adeo.no", value: "cloudsql1.adeo.no" }, { label: "cloudsql2.adeo.no", value: "cloudsql2.adeo.no" }, { label: "cloudsql3.adeo.no", value: "cloudsql3.adeo.no" }]} label={undefined} />
                            }
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.HeaderCell scope="row">
                            <Checkbox checked={accessInformatica} onChange={() => setAccessInformatica(!accessInformatica)}>Informatica</Checkbox>
                        </Table.HeaderCell>
                        <Table.DataCell>
                            {
                                accessInformatica && <UNSAFE_Combobox
                                    isMultiSelect
                                    hideLabel
                                    options={[{ label: "informatica1.adeo.no", value: "informatica1.adeo.no" }, { label: "informatica2.adeo.no", value: "informatica2.adeo.no" }, { label: "informatica3.adeo.no", value: "informatica3.adeo.no" }]} label={undefined} />
                            }
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.HeaderCell scope="row">
                            <Checkbox checked={accessOracle} onChange={() => setAccessOracle(!accessOracle)}>Oracle</Checkbox>
                        </Table.HeaderCell>
                        <Table.DataCell>
                            {
                                accessOracle && <UNSAFE_Combobox
                                    isMultiSelect
                                    hideLabel
                                    options={[{ label: "oracle1.adeo.no", value: "oracle1.adeo.no" }, { label: "oracle2.adeo.no", value: "oracle2.adeo.no" }, { label: "oracle3.adeo.no", value: "oracle3.adeo.no" }]} label={undefined} />
                            }
                        </Table.DataCell>
                    </Table.Row>
                    <Table.Row>
                        <Table.HeaderCell scope="row">
                            <Checkbox checked={accessSMTP} onChange={() => setAccessSMTP(!accessSMTP)}>SMTP</Checkbox>
                        </Table.HeaderCell>
                        <Table.DataCell>
                            {
                                accessSMTP && <UNSAFE_Combobox
                                    isMultiSelect
                                    hideLabel
                                    options={[{ label: "smtp1.adeo.no", value: "smtp1.adeo.no" }, { label: "smtp2.adeo.no", value: "smtp2.adeo.no" }, { label: "smtp3.adeo.no", value: "smtp3.adeo.no" }]} label={undefined} />
                            }
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