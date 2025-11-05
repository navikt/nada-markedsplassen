import { ChevronDownIcon, ChevronUpIcon, InformationSquareFillIcon, PencilWritingIcon, PlusCircleFillIcon, TasklistIcon, TrashIcon, XMarkIcon } from "@navikt/aksel-icons";
import { Alert, Button, Link, List, Loader, Select, Switch, Table, TextField } from "@navikt/ds-react";
import { useCreateWorkstationURLListItemForIdent, useDeleteWorkstationURLListItemForIdent, useUpdateWorkstationURLListItemForIdent, useUpdateWorkstationURLListUserSettings, useWorkstationOptions, useWorkstationURLList, useWorkstationURLListForIdent } from "./queries";
import { ColorAuxText, ColorDefaultTextInvert, ColorFailed, ColorInfo, ColorSuccessful, ColorSuccessfulAlt } from "./designTokens";
import React, { useEffect, useState } from "react";
import { isValidUrl } from "./utils";
import { addHours } from "date-fns";
import { WorkstationURLListForIdent, WorkstationURLListItem } from "../../lib/rest/generatedDto";
import { set } from "lodash";
import { IconRevert } from "./widgets/knastIcons";

// Predefined description options in Norwegian
const PREDEFINED_DESCRIPTIONS = [
    'Datakilde (Nav/intern)',
    'Datakilde (Ekstern)',
    'Koderepository',
    'IDE-Extension'
];

interface UrlItemProps {
    item: any;
    onDelete: () => void;
    onRevert: () => void;
    onEdit(): void;
    onChangeUrl: (newValue: any) => void;
    onChangeDuration: (newDuration: string) => void;
    onChangeDescription: (newDescription: string) => void;
}

const backendDurationToHours = (duration: string) => {
    switch (duration) {
        case "01:00:00":
            return 1;
        case "12:00:00":
            return 12;
        default:
            return "?";
    }
}

const backendDurationToSelectOption = (duration: string) => {
    switch (duration) {
        case "01:00:00":
            return "1hour";
        case "12:00:00":
            return "12hours";
        default:
            return "?";
    }
}

const UrlItem = ({ item, onDelete, onRevert, onEdit, onChangeUrl, onChangeDuration, onChangeDescription }: UrlItemProps) => {
    return (
        <div className="pb-2"
        style={{
            backgroundColor: item.isNew || item.isEditing ? ColorSuccessfulAlt : "transparent",
            marginTop: item.isNew || item.isEditing ? "0.5rem" : "0rem",
            marginBottom: item.isNew || item.isEditing ? "0.5rem" : "0rem",
            border: item.isNew || item.isEditing ? `1px solid ${ColorSuccessful}` : "none",
        }}>
            <div className="flex flex-row justify-between items-center pt-2 pl-2 pr-2" >

                {
                    !item.isNew && !item.isEditing ? 
                    <div className="flex flex-row items-center" style={{
                        textDecoration: item.isDeleted ? "line-through" : "none"
                    }}>                <div className="mr-4 pt-1 pb-1 pl-4 pr-4 text-sm" style={{
                    backgroundColor: ColorInfo,
                    color: ColorDefaultTextInvert,
                }}>{item.description}</div>{item.url}<p className="text-sm" style={{ color: ColorAuxText }}>&nbsp;&nbsp;varighet&nbsp;&nbsp;</p> {backendDurationToHours(item.duration)} <p className="text-sm" style={{ color: ColorAuxText }}>&nbsp;&nbsp;timer</p></div>
                        : <div className="flex flex-row items-end">
                            <Select size="small" value={item.description || "- Velg en type -"} 
                            onChange={(e) => onChangeDescription(e.target.value)} label="" >
                                <option value="- Velg en type -">- Velg en type -</option>
                                {PREDEFINED_DESCRIPTIONS.map((desc, index) => (
                                    <option key={index} value={desc}>{desc}</option>
                                ))}
                            </Select>
                            <p className="text-small pl-6 pr-2 pb-1" style={{
                                color: ColorAuxText
                            }}>URL</p>
                            <TextField type="text" size="small" label="" placeholder="Skriv inn url" value={item.url} onChange={(e) => onChangeUrl(e.target.value.trim())} />
                            <p className="text-small pl-6 pr-2 pb-1" style={{
                                color: ColorAuxText
                            }}>varighet</p>
                            <Select size="small" value={item.duration || "01:00:00"} onChange={(e) => onChangeDuration(e.target.value)} label="" >
                                <option value="01:00:00">1t</option>
                                <option value="12:00:00">12t</option>
                            </Select>
                        </div>
                }
                {(item.isDeleted || item.isEditing) && <Button variant="tertiary" size="medium" onClick={onRevert} className="p-0 ml-2">
                    <IconRevert width={22} height={22} />
                </Button>}
                {!item.isNew && !item.isDeleted && !item.isEditing &&
                    <Button variant="tertiary" size="medium" onClick={onEdit} className="p-0 ml-2">
                        <PencilWritingIcon width={22} height={22} />
                    </Button>}

                {!item.isDeleted &&
                    <Button variant="tertiary" size="medium" onClick={onDelete} className="p-0 ml-2">
                        <TrashIcon width={22} height={22} />
                    </Button>}

            </div>
            {(item.isNew || item.isEditing) && (!item.isValid && !item.isEmpty || item.exist) &&
                <p className="text-small pl-6" style={{
                    color: !item.isValid || item.isEmpty || item.exist ? ColorFailed : "transparent"
                }}>{!item.isValid ? "Ugyldig URL" : item.exist ? "Allerede lagt til" : "gyldig"}</p>

            }
            {(item.isNew || item.isEditing) && (!item.isEmpty) && (!item.description || !PREDEFINED_DESCRIPTIONS.some(it=> it === item.description)) &&
                <p className="text-small pl-6" style={{
                    color: ColorFailed 
                }}>Må velge en gyldig urltype</p>

            }
        </div>
    );
}

interface InternetOpeningsFormProps {
    onSave: () => void;
    onCancel: () => void;
}

export const InternetOpeningsForm = ({ onSave, onCancel }: InternetOpeningsFormProps) => {
    const urlList = useWorkstationURLListForIdent()
    const globalSettings = useWorkstationURLList()
    const [internetSettings, setInternetSettings] = React.useState<any>(urlList.data);
    const [showCentralList, setShowCentralList] = useState(false);
    const [backendError, setBackendError] = useState<string | undefined>(undefined);
    const [showBanList, setShowBanList] = useState(false);
    const [updateFinished, setUpdateFinished] = useState<boolean>(false);
    const createUrlMutation = useCreateWorkstationURLListItemForIdent();
    const updateUrlMutation = useUpdateWorkstationURLListItemForIdent();
    const deleteUrlMutation = useDeleteWorkstationURLListItemForIdent();
    const updateGlobalURLAllowList = useUpdateWorkstationURLListUserSettings()
    const options = useWorkstationOptions()

    const [updating, setUpdating] = useState(false);


    const submitInternetSettings = async (urlList: WorkstationURLListForIdent, internetSettings: any) => {
        const centralOpeningsChanged = urlList.disableGlobalAllowList !== internetSettings.disableGlobalAllowList;
        const urlItemChanged = (item: any) => item.id
            && urlList.items?.some(it => it?.id === item.id && (it?.url != item.url) || (it?.duration != item.duration) || (it?.description != item.description));

        setUpdateFinished(false);
        setBackendError(undefined);
        setUpdating(true);
        try {
            if (centralOpeningsChanged) {
                await updateGlobalURLAllowList.mutateAsync({
                    disableGlobalURLList: internetSettings.disableGlobalAllowList
                });
            }
            for (const it of internetSettings.items.filter((it: any) => !!it.url.trim())) {
                const createdAt = new Date().toISOString();
                const duration = it.duration === "01:00:00" ? 1 : it.duration === "12:00:00" ? 12 : 1;
                const durationParam = it.duration === "01:00:00" ? "1hour" : it.duration === "12:00:00" ? "12hour" : "1hour";
                const expiresAt = addHours(createdAt, duration).toISOString();
                if (it.isNew) {
                    await createUrlMutation.mutateAsync({
                        url: it.url,
                        createdAt: createdAt,
                        expiresAt: expiresAt,
                        description: it.description || "generic url",
                        duration: durationParam,
                        selected: true,
                    } as WorkstationURLListItem)
                } else if (it.isDeleted) {
                    await deleteUrlMutation.mutateAsync(it.id)
                } else if (it.isEditing && urlItemChanged(it)) {
                    await updateUrlMutation.mutateAsync({
                        id: it.id,
                        url: it.url,
                        duration: durationParam,
                        createdAt: createdAt,
                        expiresAt: expiresAt,
                        description: it.description,
                        selected: it.selected,
                    })
                }
            }
        } catch (error) {
            console.log("Error updating internet settings:", error);
            setBackendError("Kunne ikke lagre endringer, prøv igjen senere.");
            setUpdateFinished(false);
            setUpdating(false);
            return false
        }
        setBackendError(undefined);
        setUpdateFinished(true);
        setUpdating(false);
        return true
    }

    const submitSettings = async () => {
        const success = await submitInternetSettings(urlList.data!, internetSettings);
        if (success) {
            onSave();
        } else {
            setBackendError("Kunne ikke lagre endringer, prøv igjen senere.");
        }
    };

    useEffect(() => {
        if (urlList.data) {
            setInternetSettings(urlList.data);
        }
    }, [urlList.data]);

    const isNotFinishEditingUrls = !!internetSettings?.items?.some((it: any) => it.isNew && (it.isEmpty || !it.isValid || it.exist));

    const settingsChange = () => {
        return urlList.data?.disableGlobalAllowList !== internetSettings?.disableGlobalAllowList ||
            internetSettings?.items.filter((it:any)=>!!it.url).some((it: any) => it.isDeleted || it.isNew || urlList.data?.items?.some((backendItem) =>
                backendItem?.id === it.id && (backendItem?.url !== it.url || backendItem?.duration !== it.duration || backendItem?.description !== it.description)
            ));
    }

    const settingsValid = () => {
        return !internetSettings?.items?.filter((it: any) => !!it.url).some((it: any) => it.isNew && (!it.isValid || it.exist));
    }


    return <div className="max-w-[55rem] min-w-[50rem] border-blue-100 border rounded p-4">
        <Table>
            <Table.Header>
                <Table.Row>
                    <Table.HeaderCell colSpan={2} scope="col">
                        <div className="flex flex-row justify-between items-center">
                            <h3>
                                Internettåpeninger
                            </h3>
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
                                    <Table.HeaderCell scope="row" className="align-top">
                                        <div>Sentralt administrerte åpninger</div>
                                    </Table.HeaderCell>
                                    <Table.DataCell>
                                        <div className="max-w-[30rem] text-sm" style={{
                                            color: ColorAuxText
                                        }}>Noen åpninger mot internett har mange nytte av og vi har derfor valgt å åpne disse som standard for alle brukere. Men, du står fritt til å ikke åpne for disse.</div>
                                        <div className="flex flex-row justify-between">
                                            <Switch defaultChecked={!urlList.data.disableGlobalAllowList} onChange={e => {
                                                setInternetSettings({
                                                    ...internetSettings,
                                                    disableGlobalAllowList: !e.target.checked
                                                })
                                                setBackendError(undefined);
                                            }}><div className="flex flex-row items-start"><p>{internetSettings?.disableGlobalAllowList ? "Deaktiver" : "Aktiver"}</p>{
                                                !internetSettings?.disableGlobalAllowList && <p className="text-sm pl-1" style={{color: ColorSuccessful}}>Anbefalt</p>
                                            }</div></Switch>
                                            <Link
                                                type="button"
                                                onClick={() => setShowCentralList(!showCentralList)}
                                                className="flex items-center gap-2 text-blue-600 hover:text-blue-800 text-sm font-medium transition-colors"
                                            >
                                                <TasklistIcon className="w-4 h-4" />
                                                <span>{showCentralList ? "Skjul" : "Vis"} URL-listen ({options.data?.globalURLAllowList?.length} URL-er)</span>
                                                {showCentralList ? <ChevronUpIcon className="w-4 h-4" /> : <ChevronDownIcon className="w-4 h-4" />}
                                            </Link>
                                        </div>
                                        <div className="pt-4">
                                            {showCentralList && (
                                                <div className="max-w-140 max-h-60 overflow-y-auto overflow-x-auto">
                                                    {options.data?.globalURLAllowList?.map((url, index) => (
                                                        <div key={index} className="text-sm">
                                                            {`- ${url}`}
                                                        </div>
                                                    ))}
                                                </div>
                                            )}
                                        </div>
                                    </Table.DataCell>
                                </Table.Row>
                                <Table.Row>
                                    <Table.HeaderCell scope="row" className="align-top">
                                        <div>Globalt blokkerte URL-er</div>
                                    </Table.HeaderCell>
                                    <Table.DataCell>
                                        <div className="max-w-[30rem] text-sm" style={{
                                            color: ColorAuxText
                                        }}>URL-er som er permanent blokkert av systemadministrator.</div>
                                        <div className="flex flex-row justify-between">
                                            <Switch checked disabled><div className="flex flex-row items-start"><p>Må være aktivert</p></div></Switch>
                                            <Link
                                                type="button"
                                                onClick={() => setShowBanList(!showBanList)}
                                                className="flex items-center gap-2 text-blue-600 hover:text-blue-800 text-sm font-medium transition-colors"
                                            >
                                                <TasklistIcon className="w-4 h-4" />
                                                <span>{showBanList ? "Skjul" : "Vis"} URL-listen ({globalSettings.data?.globalDenyList.length} URL-er)</span>
                                                {showBanList ? <ChevronUpIcon className="w-4 h-4" /> : <ChevronDownIcon className="w-4 h-4" />}
                                            </Link>
                                        </div>
                                        <div className="pt-4">
                                            {showBanList && (
                                                <div className="max-w-140 max-h-60 overflow-y-auto overflow-x-auto">
                                                    {globalSettings.data?.globalDenyList?.map((url, index) => (
                                                        <div key={index} className="text-sm">
                                                            {`- ${url}`}
                                                        </div>
                                                    ))}
                                                </div>
                                            )}
                                        </div>
                                    </Table.DataCell>
                                </Table.Row>

                                <Table.Row>
                                    <Table.HeaderCell scope="row" className="align-top">
                                        <p className="pt-2">Tidsbegrensede åpninger</p>
                                    </Table.HeaderCell>
                                    <Table.DataCell>
                                        {
                                            <div className="flex flex-col">{
                                                internetSettings?.items?.map((item: any, index: number) => (
                                                    <div key={index} className="flex flex-row justify-between items-center">
                                                        <UrlItem
                                                            item={item}
                                                            onDelete={() => {
                                                                const newList = internetSettings.items[index].isNew ?
                                                                    internetSettings.items.filter((_: any, i: number) => i !== index) :
                                                                    internetSettings.items.map((item: any, i: number) => i === index
                                                                        ? { ...item, isDeleted: true, isEditing: false } : item);
                                                                setInternetSettings({
                                                                    ...internetSettings,
                                                                    items: newList
                                                                });
                                                                setBackendError(undefined);
                                                            }}
                                                            onEdit={() => {
                                                                const newList = internetSettings.items.map((item: any, i: number) => i === index
                                                                    ? { ...item, isEditing: true, isValid: item.isValid===undefined?true:item.isValid } : item);
                                                                setInternetSettings({
                                                                    ...internetSettings,
                                                                    items: newList
                                                                });
                                                                setBackendError(undefined);
                                                            }}
                                                            onRevert={() => {
                                                                const reverted = {
                                                                    ...urlList.data?.items.find(urlItem => urlItem?.id === item.id),
                                                                    isEditing: false,
                                                                    isDeleted: false
                                                                };
                                                                const newList = internetSettings.items.map((it: any, i: number) => i === index && !it.isNew
                                                                    ? reverted : it)
                                                                setInternetSettings({
                                                                    ...internetSettings,
                                                                    items: newList
                                                                });
                                                                setBackendError(undefined);
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
                                                                setBackendError(undefined);
                                                            }}
                                                            onChangeDuration={(newDuration: string) => {
                                                                const newList = internetSettings.items.map((item: any, i: number) => i === index
                                                                    ? { ...item, duration: newDuration } : item);
                                                                setInternetSettings({
                                                                    ...internetSettings,
                                                                    items: newList
                                                                });
                                                                setBackendError(undefined);
                                                            }}
                                                            onChangeDescription={(newDescription: string) => {
                                                                const newList = internetSettings.items.map((item: any, i: number) => i === index
                                                                    ? { ...item, description: newDescription } : item);
                                                                setInternetSettings({
                                                                    ...internetSettings,
                                                                    items: newList
                                                                });
                                                                setBackendError(undefined);
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
                                                                { url: "", duration: "1hour", isNew: true, isValid: false, isEmpty: true }
                                                            ]
                                                        })
                                                    }}><div className="flex flex-row space-x-1 items-center"><p>Legg til</p><PlusCircleFillIcon /></div></Button>
                                                </div>
                                            </div>
                                        }
                                    </Table.DataCell>
                                </Table.Row>
                            </Table.Body>

                        )
            }
        </Table>
        <div>
            <div className="flex flex-row pt-6">
                <Button variant="primary" disabled={updating || !settingsChange() || !settingsValid()} onClick={() => {
                    submitSettings();
                }}>Lagre{updating && <Loader />}</Button>
                <Button variant="secondary" className="ml-6" onClick={onCancel}>Tilbake</Button>
            </div>
            <div className="pt-4">
                {!updateFinished && !settingsChange() && <p className="text-sm" style={{ color: ColorAuxText }}>Ingen endringer å lagre</p>}
                {!settingsValid() && <p style={{ color: ColorFailed }}>Fiks ugyldige url før lagring</p>}
            </div>
            {backendError && <div className="pt-4">
                <Alert variant="error">{backendError}</Alert>
            </div>}
            {
                updateFinished && !settingsChange() && <div className="pt-4">
                    <Alert variant="success">Endringer lagret</Alert>
                </div>
            }
        </div>
    </div>
}