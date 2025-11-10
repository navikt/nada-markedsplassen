import { ArrowCirclepathReverseIcon, ChevronDownIcon, ChevronUpIcon, ExternalLinkIcon, FloppydiskIcon, PencilWritingIcon, PlusCircleFillIcon, QuestionmarkCircleIcon, TasklistIcon, TrashIcon, XMarkIcon } from "@navikt/aksel-icons";
import { Alert, Button, Link, Loader, Popover, Select, Switch, Table, TextField, Tooltip, UNSAFE_Combobox } from "@navikt/ds-react";
import { useCreateWorkstationURLListItemForIdent, useDeleteWorkstationURLListItemForIdent, useUpdateWorkstationURLListItemForIdent, useUpdateWorkstationURLListUserSettings, useWorkstationOptions, useWorkstationURLList, useWorkstationURLListForIdent, useWorkstationURLListGlobalAllow } from "./queries";
import { ColorAuxText, ColorFailed, ColorInfoText, ColorSuccessful, ColorSuccessfulAlt } from "./designTokens";
import { useEffect, useRef, useState } from "react";
import { isValidUrl } from "./utils";
import { addHours } from "date-fns";
import { WorkstationURLListItem } from "../../lib/rest/generatedDto";
import { IconConnected, IconDisconnected } from "./widgets/knastIcons";

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
    onEdit(): void;
    onSave(): void;
    onRevert(): void;
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


const UrlItem = ({ item, onDelete, onEdit, onSave, onRevert, onChangeUrl, onChangeDuration, onChangeDescription }: UrlItemProps) => {
    const urlInputRef = useRef<HTMLInputElement>(null);
    const [showUrlHelpText, setShowUrlHelpText] = useState(false);
    const getSaveButtonTooltip = () => {
        if (item.isEmpty) { return "Url-en er tom"; }
        if (!item.isValid) { return "Ugyldig url"; }
        if (item.exist) { return "Denne url-en er allerede lagt til"; }
        if (!item.description) { return "Må velge en type"; }
        if (!item.isChanged) { return "Ingen endringer å lagre"; }
        return "Lagre url-en";
    }
    console.log(showUrlHelpText)

    return (
        <>
            <style>
                {`
          .urlItem {
            background-color: transparent;
          }

          .urlItem:hover {
            background-color: #CCF1D6;
            border: 1px solid #2AA758;
            transition: background-color 0.1s ease;
          }
        `}
            </style>
            {
                item.isEditing ? <div className="pb-2 min-w-full"
                    style={{
                        backgroundColor: ColorSuccessfulAlt,
                        marginTop: "0.5rem",
                        marginBottom: "0.5rem",
                        border: `1px solid ${ColorSuccessful}`,
                    }}>
                    <div className="flex flex-row justify-between items-center pt-2 pl-2 pr-2" >
                        <div className="flex flex-col">
                            <div className="flex flex-row items-end">
                                <div className="min-w-80 flex flex-row items-end">
                                    <p className="text-small pl-6 pb-1" style={{
                                        color: ColorAuxText
                                    }}>URL</p>
                                    <Popover placement="top" content={"url format"} anchorEl={urlInputRef.current} open={showUrlHelpText} onClose={() => setShowUrlHelpText(false)}>
                                        <p className="p-2 text-sm">Url kan være domene eller inkludere en stikomponent.
                                            <Link className="ml-2" onClick={() => window.open("https://cloud.google.com/secure-web-proxy/docs/url-list-syntax-reference?_gl=1*1w3xzaf*_ga*Mjk4ODA4Mjg1LjE3MzY4NDE5MDU.*_ga_WH2QY8WWF5*czE3NjI3Nzc5NjYkbzM0JGcxJHQxNzYyNzc4NzQzJGoxNiRsMCRoMA..")} href="">
                                                syntaks<ExternalLinkIcon />
                                            </Link></p>
                                    </Popover>
                                    <div ref={urlInputRef} className="flex h-8 items-center" onMouseEnter={() => setShowUrlHelpText(true)} >
                                        <QuestionmarkCircleIcon color={ColorInfoText} width={16} height={16} />
                                    </div>
                                    <TextField type="text" size="small" label="" className="ml-2 w-[80%]" placeholder="Skriv inn url" value={item.url}
                                        onChange={(e) => onChangeUrl(e.target.value.trim())} />
                                    <Tooltip content={item.isEmpty ? "Url er tom" : item.isValid ? "Gyldig url" : "Ugyldig URL"}>
                                        <div className="ml-2 min-h-8 flex items-center">
                                            {item.isEmpty ? null :
                                                item.isValid
                                                    ? <IconConnected width={14} height={14} />
                                                    : <div><IconDisconnected width={14} height={14} /></div>}
                                        </div>
                                    </Tooltip>
                                </div>
                                <div className="flex flex-row items-end">
                                    <p className="text-small pl-6 pr-2 pb-1" style={{
                                        color: ColorAuxText
                                    }}>varighet</p>
                                    <Select size="small" value={item.duration || "01:00:00"} onChange={(e) => onChangeDuration(e.target.value)} label="" >
                                        <option value="01:00:00">1t</option>
                                        <option value="12:00:00">12t</option>
                                    </Select>
                                </div>
                            </div>
                            <div className="flex flex-row items-end">
                                <p className="text-small pl-6 pr-2 pb-1" style={{
                                    color: ColorAuxText
                                }}>
                                    Type
                                </p>
                                <UNSAFE_Combobox allowNewValues={true} size="small" className="w-60" label="" options={PREDEFINED_DESCRIPTIONS}
                                    defaultValue={item.description || ""}
                                    onToggleSelected={option => {
                                        onChangeDescription(option.trim())
                                    }} />
                                {
                                    !item.isUpdating ?
                                        <div className="flex flex-row">
                                            <Tooltip content={getSaveButtonTooltip()}>
                                                <div className="flex items-center">
                                                    <Button variant="tertiary" size="medium" onClick={onSave} className="p-0 ml-2"
                                                        disabled={!item.isValid || item.isEmpty || item.exist || !item.description || !item.isChanged}>
                                                        <FloppydiskIcon width={28} height={28} />
                                                    </Button>
                                                </div>
                                            </Tooltip>
                                            <Tooltip content={"Reverter endringer"}>
                                                <div className="flex items-center">
                                                    <Button variant="tertiary" size="medium" onClick={onRevert} className="p-0 ml-2">
                                                        <ArrowCirclepathReverseIcon width={28} height={28} />
                                                    </Button>
                                                </div>
                                            </Tooltip>
                                        </div> :
                                        <div><Loader size="small" className="ml-2" /></div>
                                }
                            </div>
                        </div>
                    </div>
                </div>
                    : <div className="urlItem flex flex-row justify-between w-full items-center pt-2 pb-2 pl-4 pr-4" >

                        <div className="flex flex-row items-center" style={{
                            textDecoration: item.isDeleted ? "line-through" : "none"
                        }}> {item.url}<p className="text-sm" style={{ color: ColorAuxText }}>&nbsp;&nbsp;varighet&nbsp;&nbsp;</p> {backendDurationToHours(item.duration)} <p className="text-sm" style={{ color: ColorAuxText }}>&nbsp;&nbsp;timer</p></div>
                        <Tooltip content="Rediger">
                            <Button variant="tertiary" size="medium" onClick={onEdit} className="p-0 ml-2">
                                <PencilWritingIcon width={22} height={22} />
                            </Button>
                        </Tooltip>

                        <Tooltip content="Slett">
                            <Button variant="tertiary" size="medium" onClick={onDelete} className="p-0 ml-2">
                                <TrashIcon width={22} height={22} />
                            </Button>
                        </Tooltip>
                        <div className="ml-auto pt-1 pb-1 pl-4 pr-4 text-sm" style={{
                            color: ColorAuxText,
                        }}>{item.description}</div>
                    </div>
            }
        </>
    );
}

interface InternetOpeningsFormProps {
    onSave: () => void;
    onCancel: () => void;

}

export const InternetOpeningsForm = ({ onSave, onCancel }: InternetOpeningsFormProps) => {
    const workstationInternetSettings = useWorkstationURLListForIdent()
    const globalSettings = useWorkstationURLList()
    const [showCentralList, setShowCentralList] = useState(false);
    const [backendError, setBackendError] = useState<string | undefined>(undefined);
    const [showBanList, setShowBanList] = useState(false);
    const createUrlMutation = useCreateWorkstationURLListItemForIdent();
    const updateUrlMutation = useUpdateWorkstationURLListItemForIdent();
    const deleteUrlMutation = useDeleteWorkstationURLListItemForIdent();
    const updateGlobalURLAllowList = useUpdateWorkstationURLListUserSettings()
    const options = useWorkstationOptions()
    const [editingUrls, setEditingUrls] = useState<any[]>([]);
    const [updatingUrlIDs, setUpdatingUrlIDs] = useState<Set<string>>(new Set());
    const [disableGlobalAllowList, setDisableGlobalAllowList] = useState<boolean | undefined>(workstationInternetSettings.data?.disableGlobalAllowList);

    useEffect(() => {
        if (workstationInternetSettings.data && disableGlobalAllowList === undefined) {
            setDisableGlobalAllowList(workstationInternetSettings.data.disableGlobalAllowList);
        }
    }, [workstationInternetSettings.data, disableGlobalAllowList]);

    const toggleGlobalAllowList = async (enable: boolean) => {
        setBackendError(undefined);
        setDisableGlobalAllowList(!enable);
        try {
            await updateGlobalURLAllowList.mutateAsync({
                disableGlobalURLList: !enable
            });
        } catch (error) {
            setBackendError("Kunne ikke oppdatere sentrale innstillinger, prøv igjen senere.");
            setDisableGlobalAllowList(workstationInternetSettings.data?.disableGlobalAllowList);
        }
    }

    const saveUrlItem = async (urlListItem: any) => {
        setBackendError(undefined);
        const createdAt = new Date().toISOString();
        const duration = urlListItem.duration === "01:00:00" ? 1 : urlListItem.duration === "12:00:00" ? 12 : 1;
        const durationParam = urlListItem.duration === "01:00:00" ? "1hour" : urlListItem.duration === "12:00:00" ? "12hour" : "1hour";
        const expiresAt = addHours(createdAt, duration).toISOString();
        setUpdatingUrlIDs(new Set(updatingUrlIDs).add(urlListItem.id));
        try {
            if (!urlListItem.id) {
                await createUrlMutation.mutateAsync({
                    url: urlListItem.url,
                    createdAt: createdAt,
                    expiresAt: expiresAt,
                    description: urlListItem.description || "generic url",
                    duration: durationParam,
                    selected: true,
                } as WorkstationURLListItem)
            } else {
                await updateUrlMutation.mutateAsync({
                    id: urlListItem.id,
                    url: urlListItem.url,
                    duration: durationParam,
                    createdAt: createdAt,
                    expiresAt: expiresAt,
                    description: urlListItem.description,
                    selected: urlListItem.selected,
                })
            }
            setEditingUrls(editingUrls.filter(url => url.id !== urlListItem.id));
        } catch (error) {
            setBackendError("Kunne ikke lagre endringer, prøv igjen senere.");
        } finally {
            setUpdatingUrlIDs(new Set([...updatingUrlIDs].filter(id => id !== urlListItem.id)));
        }
    }

    const deleteUrlItem = async (urlListItem: any) => {
        if (!urlListItem.id) {
            setEditingUrls(editingUrls.filter(url => url.id !== urlListItem.id));
            return
        }

        setBackendError(undefined);
        setUpdatingUrlIDs(new Set([...updatingUrlIDs, urlListItem.id]));
        try {
            await deleteUrlMutation.mutateAsync(urlListItem.id)
            setEditingUrls(editingUrls.filter(url => url.id !== urlListItem.id));
        } catch (error) {
            setBackendError("Kunne ikke slette url-en, prøv igjen senere.");
        } finally {
            setUpdatingUrlIDs(new Set([...updatingUrlIDs].filter(id => id !== urlListItem.id)));
        }
    }

    const reverUrlItem = (urlListItem: any) => {
        setEditingUrls(editingUrls.filter(url => url.id !== urlListItem.id));
    }

    return <div className="max-w-220 min-w-220 border-blue-100 border rounded p-4">
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
                workstationInternetSettings.isLoading ? (<div className="text-center" style={{ color: ColorAuxText }}>Henter konfigurasjon<Loader /></div>) :
                    (workstationInternetSettings.isError || !workstationInternetSettings.data) ? (
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
                                        <div className="max-w-120 text-sm" style={{
                                            color: ColorAuxText
                                        }}>Noen åpninger mot internett har mange nytte av og vi har derfor valgt å åpne disse som standard for alle brukere. Men, du står fritt til å ikke åpne for disse.</div>
                                        <div className="flex flex-row justify-between">
                                            <Switch checked={!disableGlobalAllowList} onChange={e => {
                                                toggleGlobalAllowList(e.target.checked);
                                            }}><div className="flex flex-row items-start"><p>{disableGlobalAllowList ? "Deaktiver" : "Aktiver"}</p>{
                                                !disableGlobalAllowList && <p className="text-sm pl-1" style={{ color: ColorSuccessful }}>Anbefalt</p>
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
                                        <div className="max-w-120 text-sm" style={{
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
                                                [...workstationInternetSettings?.data?.items.map(it => {
                                                    const editingItem = editingUrls.find(url => url.id === it?.id);
                                                    return editingItem ? ({
                                                        ...editingItem,
                                                        isEditing: true,
                                                        isValid: isValidUrl(editingItem.url),
                                                        isUpating: updatingUrlIDs.has(editingItem.id),
                                                        isEmpty: editingItem.url.trim() === "",
                                                        isChanged: editingItem.url !== it?.url || editingItem.duration !== it?.duration || editingItem.description !== it?.description,
                                                        exist: workstationInternetSettings.data?.items?.some((item: any) => item.url === editingItem.url && item.id !== editingItem.id)
                                                    }) : it;
                                                })
                                                    , ...editingUrls.filter(it => !it.id)].map((item: any, index: number) => (
                                                        <div key={index} className="flex flex-row justify-between items-center">
                                                            <UrlItem
                                                                item={item}
                                                                onDelete={() => {
                                                                    deleteUrlItem(item);
                                                                }}
                                                                onEdit={() => {
                                                                    console.log("edit", item);
                                                                    console.log("before edit", editingUrls);
                                                                    setEditingUrls([...editingUrls.filter(url => url.id !== item.id), item]);
                                                                }}
                                                                onSave={() => {
                                                                    saveUrlItem(item);
                                                                }}
                                                                onRevert={() => {
                                                                    reverUrlItem(item);
                                                                }}
                                                                onChangeUrl={(newValue: any) => {
                                                                    console.log("change url", newValue);
                                                                    setEditingUrls([...editingUrls.filter(url => url.id !== item.id), {
                                                                        ...item, url: newValue
                                                                        , isEmpty: newValue.trim() === ""
                                                                        , isValid: isValidUrl(newValue)
                                                                        , exist: workstationInternetSettings.data?.items?.some((it: any) => it.url === newValue && it.id !== item.id)
                                                                    }]);
                                                                }}
                                                                onChangeDuration={(newDuration: string) => {
                                                                    setEditingUrls([...editingUrls.filter(url => url.id !== item.id), { ...item, duration: newDuration }]);
                                                                }}
                                                                onChangeDescription={(newDescription: string) => {
                                                                    setEditingUrls([...editingUrls.filter(url => url.id !== item.id), { ...item, description: newDescription }]);
                                                                }}
                                                            />
                                                        </div>))
                                            }
                                                <div className="flex flex-row justify-end items-center">
                                                    {!workstationInternetSettings.data?.items?.length && !editingUrls.length && <div style={{
                                                        color: ColorAuxText
                                                    }}>Ingen åpninger konfigurert</div>}
                                                    <Button variant="tertiary" disabled={editingUrls?.some((it: any) => !it.id && it.isEditing)}
                                                        onClick={() => {
                                                            setEditingUrls([...editingUrls, { id: undefined, url: "", duration: "1hour", isEditing: true, isValid: false, isEmpty: true, isChanged: true }]);
                                                        }} >
                                                        <div className="flex flex-row space-x-1 items-center"><p>Legg til</p><PlusCircleFillIcon /></div></Button>
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
                <Button variant="secondary" className="ml-6" onClick={onCancel}>Tilbake</Button>
            </div>
            {backendError && <div className="pt-4">
                <Alert variant="error">{backendError}</Alert>
            </div>}
        </div>
    </div>
}