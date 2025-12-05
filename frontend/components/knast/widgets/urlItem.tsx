import { useEffect, useRef, useState } from "react";
import { ColorAuxText, ColorDefaultText, ColorDisabled, ColorFailed, ColorInfoText, ColorSuccessful, ColorSuccessfulAlt } from "../designTokens";
import { Button, Checkbox, Link, Loader, Popover, Select, TextField, Tooltip, UNSAFE_Combobox } from "@navikt/ds-react";
import { ArrowCirclepathReverseIcon, ChevronUpDoubleIcon, ClockIcon, ExternalLinkIcon, FloppydiskIcon, PencilWritingIcon, QuestionmarkCircleIcon, TrashIcon } from "@navikt/aksel-icons";
import { IconConnected, IconConnectLightGray, IconDisconnected } from "./knastIcons";

// Predefined description options in Norwegian
const PREDEFINED_DESCRIPTIONS = [
    'Datakilde (Nav/intern)',
    'Datakilde (Ekstern)',
    'Koderepository',
    'IDE-Extension'
];

type UrlItemStyle = "view" | "edit" | "status" | "pick"

export interface UrlItemProps {
    item: any;
    style?: UrlItemStyle;
    status?: "connected" | "expired" | "disabled" | "unavailable";
    className?: string;
    selectedItems?: string[];
    onDelete?: () => void;
    onEdit?: () => void;
    onSave?: () => void;
    onRevert?: () => void;
    onToggle?: () => void;
    onChangeUrl?: (newValue: any) => void;
    onChangeDuration?: (newDuration: string) => void;
    onChangeDescription?: (newDescription: string) => void;
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

export interface UrlTextDisplayProps {
    url: string;
    className?: string;
    lengthLimitHead?: number;
    lengthLimitTail?: number;
}

export const UrlTextDisplay = ({ url, className, lengthLimitHead = 10, lengthLimitTail = 10 }: UrlTextDisplayProps) => {
    const [showFullUrl, setShowFullUrl] = useState(false);
    const totalLengthLimit = lengthLimitHead + lengthLimitTail + 3; // 3 for the ellipsis
    return <>
    <Tooltip content={showFullUrl ? "Klikk for å folde sammen" : "Klikk for å folde ut"}>
        <div className={className} onClick={() => setShowFullUrl(!showFullUrl)}>
            {(url?.length > totalLengthLimit && !showFullUrl)
                ? <div className="text-nowrap">
                    {url.substring(0, lengthLimitHead)}
                    ...{url.substring(url.length - lengthLimitTail)}
                </div> : <div className="wrap-break-word min-w-20 max-w-60" >{url}</div>}

        </div>
    </Tooltip>
    </>
}

const UrlItemViewStyle = ({ item, onEdit, onDelete }: UrlItemProps) => {
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
                <div className="urlItem flex flex-row justify-between w-full items-center pt-2 pb-2 pl-4 pr-4" >

                    <div className="flex flex-row items-center">
                        <UrlTextDisplay url={item.url} />
                        <Tooltip content="URL-en vil bli deaktivert etter at den har vært aktivert i den angitte tiden">
                            <div className="flex flex-row items-center ml-4">
                                <p className="text-sm ml-2" style={{ color: ColorAuxText }}><ClockIcon width={16} height={16} color={ColorSuccessful} className="m-1" /></p>
                                {backendDurationToHours(item.duration)} <p className="text-sm" style={{ color: ColorAuxText }}>&nbsp;&nbsp;timer</p>
                            </div>
                        </Tooltip>

                    </div>

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
                </div >
            }
        </>
    );
}

const UrlItemEditStyle = ({ item, onChangeUrl, onChangeDuration, onChangeDescription, onSave, onRevert }: UrlItemProps) => {
    const urlInputRef = useRef<HTMLDivElement>(null);
    const expireInputRef = useRef<HTMLDivElement>(null);
    const [showUrlHelpText, setShowUrlHelpText] = useState(false);
    const [showExpireHelpText, setShowExpireHelpText] = useState(false);
    const getSaveButtonTooltip = () => {
        if (item.isEmpty) { return "Url-en er tom"; }
        if (!item.isValid) { return "Ugyldig url"; }
        if (item.exist) { return "Denne url-en er allerede lagt til"; }
        if (!item.description) { return "Må velge en type"; }
        if (!item.isChanged) { return "Ingen endringer å lagre"; }
        return "Lagre url-en";
    }

    return (
        <>
            {
                <div className="pb-2 min-w-full"
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
                                        <p className="p-2 text-sm">Url kan være domene eller inkludere en stikomponent, ikke inkluder protokoll, f.eks. «http://».
                                            <Link className="ml-2" onClick={() => window.open("https://cloud.google.com/secure-web-proxy/docs/url-list-syntax-reference?_gl=1*1w3xzaf*_ga*Mjk4ODA4Mjg1LjE3MzY4NDE5MDU.*_ga_WH2QY8WWF5*czE3NjI3Nzc5NjYkbzM0JGcxJHQxNzYyNzc4NzQzJGoxNiRsMCRoMA..")} href="">
                                                syntaks<ExternalLinkIcon />
                                            </Link></p>
                                    </Popover>
                                    <div ref={urlInputRef} className="flex h-8 items-center" onMouseEnter={() => setShowUrlHelpText(true)} >
                                        <QuestionmarkCircleIcon color={ColorInfoText} width={16} height={16} />
                                    </div>
                                    <TextField type="text" size="small" label="" className="ml-2 w-[80%]" placeholder="Skriv inn url" value={item.url}
                                        onChange={(e) => onChangeUrl?.(e.target.value.trim())} />
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
                                    <Popover placement="top" content={"url format"} anchorEl={expireInputRef.current} open={showExpireHelpText} onClose={() => setShowExpireHelpText(false)}>
                                        <p className="p-2 text-sm">URL-en vil bli deaktivert etter at den har vært aktivert i den angitte tiden</p>
                                    </Popover>
                                    <div className="flex flex-row items-center ml-6 mr-2">
                                        <p className="text-small" style={{
                                            color: ColorAuxText
                                        }}>varighet </p>
                                        <div ref={expireInputRef} className="flex h-8 items-center" onMouseEnter={() => setShowExpireHelpText(true)} onMouseLeave={() => setShowExpireHelpText(false)}>
                                            <QuestionmarkCircleIcon color={ColorInfoText} width={16} height={16} />
                                        </div>

                                    </div>
                                    <Select size="small" value={item.duration || "01:00:00"} onChange={(e) => onChangeDuration?.(e.target.value)} label="" >
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
                                        onChangeDescription?.(option.trim())
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
            }
        </>
    );
}

const UrlItemStatusStyle = ({ item, status }: UrlItemProps) => {
    const expires = new Date(item.expiresAt)
    const expiresIn = expires.getTime() - Date.now()
    const hours = Math.floor(expiresIn / (1000 * 60 * 60));
    const minutes = Math.floor((expiresIn % (1000 * 60 * 60)) / (1000 * 60));
    const durationText = hours > 0 ? `${hours}t ${minutes}m` : `${minutes}m`;

    switch (status) {
        case "unavailable":
            return <Tooltip content="Du kan ikke aktivere internett når knast ikke er startet">
                <div className="grid grid-cols-[20px_1fr] items-center">
                    <IconConnectLightGray />
                    <div className="flex flex-row gap-x-2 items-center"><p style={{
                        color: ColorDefaultText
                    }}><UrlTextDisplay url={item.url} /></p>
                    </div>
                </div>
            </Tooltip>
        case "expired":
            return <>
                <div className="grid grid-cols-[20px_1fr] items-center">
                    <IconDisconnected width={12} />
                    <div className="flex flex-row gap-x-2 items-center">
                        <p><UrlTextDisplay url={item.url} /></p>
                        <p className="text-sm" style={{
                            color: ColorFailed
                        }}>Utløpt</p>
                    </div>
                </div>
            </>
        case "connected":
            return <>
                <div className="grid grid-cols-[20px_1fr] items-center">
                    <IconConnected width={12} />
                    <div className="flex flex-row items-center">
                        <p><UrlTextDisplay url={item.url} /></p>
                        <ClockIcon width={16} height={16} color={ColorSuccessful} className="ml-2" />
                        <p className="text-sm" style={{
                            color: ColorSuccessful
                        }}>{durationText}</p>
                    </div>
                </div>
            </>

        case "disabled":
            return <>
                <div className="grid grid-cols-[20px_1fr] items-center">
                    <IconConnectLightGray />
                    <div className="flex flex-row gap-x-2 items-center">
                        <p style={{
                            color: ColorDisabled
                        }}><UrlTextDisplay url={item.url} /></p>
                    </div >
                </div>
            </>
    }
}

const UrlItemPickStyle = ({ item, selectedItems, onToggle }: UrlItemProps) => {
    return <div className="pt-2 flex flex-row items-center">
        <Checkbox checked={selectedItems?.includes(item.id)} size="small"
            onChange={onToggle}
        >
            {""}
        </Checkbox>
        <div className="flex flex-row gap-x-2 items-center">
            <p><UrlTextDisplay url={item.url} /></p>
            <p style={{
                color: ColorAuxText
            }}>{item.duration === "01:00:00" ? "1t" : item.duration === "12:00:00" ? "12t" : "?t"}</p>
            {item.selected !== selectedItems?.includes(item.id) && <Loader size="small" className="ml-2" />}
        </div>
    </div>

}

export const UrlItem = ({ item, style, status, selectedItems, onDelete, onEdit, onSave, onRevert, onChangeUrl, onChangeDuration, onToggle, onChangeDescription }: UrlItemProps) => {
    switch (style) {
        case "view":
            return <UrlItemViewStyle item={item} onDelete={onDelete} onEdit={onEdit} />;
        case "edit":
            return <UrlItemEditStyle item={item} onChangeUrl={onChangeUrl} onChangeDuration={onChangeDuration} onChangeDescription={onChangeDescription} onSave={onSave} onRevert={onRevert} />;
        case "status":
            return <UrlItemStatusStyle item={item} status={status} />;
        case "pick":
            return <UrlItemPickStyle item={item} selectedItems={selectedItems} onToggle={onToggle} />;
        default:
            return null;
    }
}
