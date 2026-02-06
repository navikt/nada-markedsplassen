import { ArrowCirclepathReverseIcon, ChevronDownIcon, ChevronUpIcon, ExternalLinkIcon, FloppydiskIcon, PencilWritingIcon, PlusCircleFillIcon, QuestionmarkCircleIcon, TasklistIcon, TrashIcon, XMarkIcon } from "@navikt/aksel-icons";
import { Alert, Button, Link, Loader, Popover, Select, Switch, Table, TextField, Tooltip, UNSAFE_Combobox } from "@navikt/ds-react";
import { useCreateWorkstationURLListItemForIdent, useDeleteWorkstationURLListItemForIdent, useUpdateWorkstationURLListItemForIdent, useUpdateWorkstationURLListUserSettings, useWorkstationOptions, useWorkstationURLList, useWorkstationURLListForIdent, useWorkstationURLListGlobalAllow } from "./queries";
import { ColorAuxText, ColorFailed, ColorInfoText, ColorSuccessful, ColorSuccessfulAlt } from "./designTokens";
import { useEffect, useRef, useState } from "react";
import { isValidUrl } from "./utils";
import { addHours } from "date-fns";
import { WorkstationURLListItem } from "../../lib/rest/generatedDto";
import { IconConnected, IconDisconnected } from "./widgets/knastIcons";
import { UrlItem } from "./widgets/urlItem";
import { useAutoCloseAlert } from "./widgets/autoCloseAlert";

export const InternetOpeningsForm = () => {
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
    const { showAlert, AutoHideAlert } = useAutoCloseAlert(5000);
    const [showSavingToggleGlobalAllowList, setShowSavingToggleGlobalAllowList] = useState<boolean>(false);

    useEffect(() => {
        if (workstationInternetSettings.data && disableGlobalAllowList === undefined) {
            setDisableGlobalAllowList(workstationInternetSettings.data.disableGlobalAllowList);
        }
    }, [workstationInternetSettings.data, disableGlobalAllowList]);

    const toggleGlobalAllowList = async (enable: boolean) => {
        setBackendError(undefined);
        setDisableGlobalAllowList(!enable);
        setShowSavingToggleGlobalAllowList(true);
        try {
            await updateGlobalURLAllowList.mutateAsync({
                disableGlobalURLList: !enable
            });
            showAlert();
        } catch (error) {
            setBackendError("Kunne ikke oppdatere sentrale innstillinger, prøv igjen senere.");
            setDisableGlobalAllowList(workstationInternetSettings.data?.disableGlobalAllowList);
        } finally {
            setShowSavingToggleGlobalAllowList(false);
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

    return <div className="w-180 border-gray-300 border-l pl-6">
        <Table>
            {
                workstationInternetSettings.isLoading ? (<div className="text-center" style={{ color: ColorAuxText }}>Henter konfigurasjon<Loader /></div>) :
                    (workstationInternetSettings.isError || !workstationInternetSettings.data) ? (
                        <div className="mt-4">
                            <Alert variant="error">Kunne ikke hente internettåpninger</Alert>
                        </div>)
                        : (
                            <Table.Body>
                                <Table.Row>
                                    <Table.DataCell>
                                        <h4 className="pb-2">Sentralt administrerte åpninger</h4>
                                        <div className="max-w-120 text-sm" style={{
                                            color: ColorAuxText
                                        }}>Noen åpninger mot internett har mange nytte av og vi har derfor valgt å åpne disse som standard for alle brukere. Men, du står fritt til å ikke åpne for disse.</div>
                                        <div className="flex flex-row justify-between">
                                            <Switch checked={!disableGlobalAllowList} onChange={e => {
                                                toggleGlobalAllowList(e.target.checked);
                                            }}><div className="flex flex-row items-start"><p>{disableGlobalAllowList ? "Deaktivert" : "Aktivert"}</p>{
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
                                        <AutoHideAlert variant="success">Innstillinger lagret</AutoHideAlert>
                                        {showSavingToggleGlobalAllowList && <div className="text-sm" style={{ color: ColorAuxText }}>Lagrer<Loader /></div>}
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
                                    <Table.DataCell>
                                        <h4 className="pb-2">Globalt blokkerte URL-er</h4>
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
                                    <Table.DataCell>
                                        <div className="flex flex-row justify-between items-center">
                                            <h4 className="pt-2 pb-2">Tidsbegrensede åpninger</h4>
                                            <Button variant="tertiary" disabled={editingUrls?.some((it: any) => !it.id && it.isEditing)}
                                                onClick={() => {
                                                    setEditingUrls([...editingUrls, { id: undefined, url: "", duration: "1hour", isEditing: true, isValid: false, isEmpty: true, isChanged: true }]);
                                                }} >
                                                <div className="flex flex-row space-x-1 items-center"><p>Legg til</p><PlusCircleFillIcon /></div></Button>

                                        </div>
                                        {
                                            <div className="flex flex-col ml-4">{
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
                                                                style={item.isEditing ? "edit" : "view"}
                                                                onDelete={() => {
                                                                    deleteUrlItem(item);
                                                                }}
                                                                onEdit={() => {
                                                                    setEditingUrls([...editingUrls.filter(url => url.id !== item.id), item]);
                                                                }}
                                                                onSave={() => {
                                                                    saveUrlItem(item);
                                                                }}
                                                                onRevert={() => {
                                                                    reverUrlItem(item);
                                                                }}
                                                                onChangeUrl={(newValue: any) => {
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
                                            </div>
                                        }
                                        {!workstationInternetSettings.data?.items?.length && !editingUrls.length && <div style={{
                                            color: ColorAuxText
                                        }}>Ingen åpninger konfigurert</div>}
                                    </Table.DataCell>

                                </Table.Row>

                            </Table.Body>

                        )
            }
        </Table>
        <div>
            {backendError && <div className="pt-4">
                <Alert variant="error">{backendError}</Alert>
            </div>}
        </div>
    </div>
}