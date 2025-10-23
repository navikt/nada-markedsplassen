import { InformationSquareFillIcon, PlusCircleFillIcon, TrashIcon, XMarkIcon } from "@navikt/aksel-icons";
import { Alert, Button, Link, Loader, Select, Switch, Table, TextField } from "@navikt/ds-react";
import { useWorkstationURLList } from "./queries";
import { ColorAuxText, ColorFailed } from "./designTokens";
import React, { useEffect } from "react";
import { isValidUrl } from "./utils";

interface UrlItemProps {
    item: any;
    onDelete: () => void;
    onChangeUrl: (newValue: any) => void;
    onChangeDuration: (newDuration: string) => void;
}

const UrlItem = ({ item, onDelete, onChangeUrl, onChangeDuration }: UrlItemProps) => {
    return (
        <div style={{
            backgroundColor: item.isNew ? "#E6F0FF" : "transparent",
        }}>
            <div className="flex flex-row justify-between items-center pt-2 pl-2 pr-2" >
                {
                    item.isNew ?
                        <TextField type="text" size="small" label="" placeholder="Skriv inn url" value={item.url} onChange={(e) => onChangeUrl(e.target.value.trim())} />
                        :
                        <div>{item.url}</div>
                }
                <p className="text-small p-2" style={{
                    color: ColorAuxText
                }}>varighet</p>
                {
                    item.isNew ? <Select size="small" value={item.duration || "01:00:00"} onChange={(e) => onChangeDuration(e.target.value)} label="" >
                        <option value="01:00:00">1</option>
                        <option value="12:00:00">12</option>
                    </Select>
                        : <div>{item.expiration}</div>
                }
                <p className="text-small p-2" style={{
                    color: ColorAuxText
                }}>timer</p>
                <Button variant="tertiary" size="medium" onClick={onDelete} className="p-0 ml-2">
                    <TrashIcon width={22} height={22} />
                </Button>

            </div>
            <p className="text-small pl-6" style={{
                color: !item.isValid && !item.isEmpty || item.exist ? ColorFailed : "transparent"
            }}>{!item.isValid ? "Ugyldig URL" : item.exist ? "Allerede lagt til" : "gyldig"}</p>
        </div>
    );
}

interface InternetOpeningsFormProps {
    onSave: () => void;
    onCancel: () => void;
}

export const InternetOpeningsForm = ({ onSave, onCancel }: InternetOpeningsFormProps) => {
    const urlList = useWorkstationURLList()
    const [internetSettings, setInternetSettings] = React.useState<any>(urlList.data);
    useEffect(() => {
        if (urlList.data) {
            setInternetSettings(urlList.data);
        }
    }, [urlList.data]);

    const isNotFinishEditingUrls = !!internetSettings?.items?.some((it: any) => it.isNew && (it.isEmpty || !it.isValid || it.exist));

    const settingsChange = () => {
        return urlList.data?.disableGlobalAllowList !== internetSettings?.disableGlobalAllowList ||
            internetSettings?.items.some((it: any) => it.isDeleted || it.isNew)
    }

    const settingsValid = () => {
        return !internetSettings?.items?.filter((it: any) => !it.isEmpty).some((it: any) => it.isNew && (!it.isValid || it.exist));
    }


    return <div className="max-w-[45rem] min-w-[45rem] border-blue-100 border rounded p-4">
        <Table>
            <Table.Header>
                <Table.Row>
                    <Table.HeaderCell colSpan={2} scope="col">
                        <div className="flex flex-row justify-between items-center">
                            <div>
                                Internettåpeninger
                            </div>
                            <Button variant="tertiary" size="small" onClick={onCancel}>
                                <XMarkIcon width={20} height={20} />
                            </Button>
                        </div>
                    </Table.HeaderCell>
                </Table.Row>
            </Table.Header>
            {
                urlList.isLoading ? (<div className="text-center" style={{ color: ColorAuxText }}>Henter konfigurasjon<Loader /></div>) :
                    (urlList.isError || !urlList.data) ? (
                        <div className="mt-4">
                            <Alert variant="error">Kunne ikke hente internettåpninger</Alert>
                        </div>)
                        : (
                            <Table.Body>
                                <Table.Row>
                                    <Table.HeaderCell scope="row">
                                        <div className="flex flex-row gap-2 items-center">
                                            <div>Sentralt administrerte åpninger</div>
                                            <Link><InformationSquareFillIcon width={20} height={20} /></Link>
                                        </div>
                                    </Table.HeaderCell>
                                    <Table.DataCell>
                                        <Switch defaultChecked={!urlList.data.disableGlobalAllowList}> </Switch>
                                    </Table.DataCell>
                                </Table.Row>

                                <Table.Row>
                                    <Table.HeaderCell scope="row" className="align-top">
                                        <p className="pt-2">Tidsbegrensede åpninger</p>
                                    </Table.HeaderCell>
                                    <Table.DataCell>
                                        {
                                            <div className="flex flex-col">{
                                                internetSettings?.items?.map((url: string, index: number) => (
                                                    <div key={index} className="flex flex-row justify-between items-center">
                                                        <UrlItem
                                                            item={url}
                                                            onDelete={() => {
                                                                const newList = internetSettings.items.filter((_: any, i: number) => i !== index);
                                                                setInternetSettings({
                                                                    ...internetSettings,
                                                                    items: newList
                                                                });
                                                            }}
                                                            onChangeUrl={(newValue: any) => {
                                                                const newList = internetSettings.items.map((item: any, i: number) => i === index
                                                                    ? {
                                                                        ...item, url: newValue, isEmpty: !newValue, isValid: newValue && isValidUrl(newValue), exist:
                                                                            internetSettings.items.some((otherItem: any, otherIndex: number) => otherIndex !== index && otherItem.url === newValue)
                                                                    } : item);
                                                                setInternetSettings({
                                                                    ...internetSettings,
                                                                    items: newList
                                                                });
                                                            }}
                                                            onChangeDuration={(newDuration: string) => {
                                                                const newList = internetSettings.items.map((item: any, i: number) => i === index
                                                                    ? { ...item, duration: newDuration } : item);
                                                                setInternetSettings({
                                                                    ...internetSettings,
                                                                    items: newList
                                                                });
                                                            }}
                                                        />
                                                    </div>))
                                            }
                                                <div className="flex flex-row justify-end mt-4 items-center">
                                                    {!internetSettings?.items?.length && <div style={{
                                                        color: ColorAuxText
                                                    }}>Ingen åpninger konfigurert</div>}
                                                    <Button disabled={isNotFinishEditingUrls} variant="tertiary" onClick={() => {
                                                        setInternetSettings({
                                                            ...internetSettings,
                                                            items: [
                                                                ...internetSettings.items || [],
                                                                { url: "", isNew: true, isValid: false, isEmpty: true }
                                                            ]
                                                        })
                                                    }}><div className="flex flex-row space-x-1 items-center"><p>Legg til</p><PlusCircleFillIcon /></div></Button>
                                                </div>
                                            </div>
                                        }
                                    </Table.DataCell>
                                </Table.Row>
                                <Table.Row>
                                    <div className="pt-4">
                                        {!settingsChange() && <p className="text-sm" style={{ color: ColorAuxText }}>Ingen endringer å lagre</p>}
                                        {!settingsValid() && <p style={{ color: ColorFailed }}>Fiks ugyldige url før lagring</p>}
                                    </div>

                                    <div className="flex flex-row pt-6">
                                        <Button variant="primary" disabled={!settingsChange() || !settingsValid()} onClick={() => {
                                            setInternetSettings({
                                                ...internetSettings,
                                                urlAllowList: internetSettings.items.filter((item: any) => !item.isEmpty)
                                            })
                                        }}>Lagre</Button>
                                        <Button variant="secondary" className="ml-6" onClick={onCancel}>Avbryt</Button>
                                    </div>
                                </Table.Row>
                            </Table.Body>

                        )
            }
        </Table>
    </div>
}