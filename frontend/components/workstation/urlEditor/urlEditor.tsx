import { ArrowUndoIcon, PencilIcon, PlusCircleFillIcon, TrashIcon } from "@navikt/aksel-icons";
import { Link, Table, Textarea, TextField, VStack } from "@navikt/ds-react";
import { useState } from "react";

const isValidUrl = (url: string) => {
    const urlPattern = /^((\*|\*?[a-zA-Z0-9-]+)\.)+[a-zA-Z0-9-]{2,}(\/(\*|[a-zA-Z0-9-._~:/?#[\]@!$&'()*+,;=]*))*$/;
    return !url ? true : urlPattern.test(url);
}

export type FrontendUrlListEntry = {
    id: number,
    defaultUrl: string,
    url: string,
    delete: boolean,
    edit: boolean,
}

interface UrlEntryProps {
    entry: FrontendUrlListEntry
    onEdit: (id: number) => void
    onChange: (id: number, v: string) => void
    onDelete: (id: number) => void
    onRevert: (id: number) => void
}

export const isAdded = (entry: FrontendUrlListEntry)=> !entry.defaultUrl


const UrlEntry = ({ entry, onChange, onEdit, onDelete, onRevert }: UrlEntryProps) => {
    const [value, setValue] = useState(entry.url)
    const urlStyle = entry.delete ? `line-through` : ""
    return ((entry.edit && !entry.delete)|| isAdded(entry)) ?
        <div className="w-96 flex items-center ml-1">
            <TextField error={isValidUrl(value) ? undefined : "ulovlig format"} className="w-[20rem] mr-2" size="small" label="" value={value}
                onChange={e => setValue(e.target.value)}
                onBlur={e => onChange(entry.id, value)}></TextField>
            {!isAdded && <Link href="#" onClick={() => onRevert(entry.id)}><ArrowUndoIcon title="a11y-title" fontSize="1.5rem" /></Link>}
            <Link href="#" onClick={() => onDelete(entry.id)}><TrashIcon title="a11y-title" fontSize="1.5rem" /></Link>
        </div>
        :
        <div className="w-96 flex items-center ml-3">
            <div className={`${urlStyle} mr-2`}>{entry.url}</div>
            {!entry.delete && <Link href="#" onClick={() => onEdit(entry.id)}><PencilIcon title="a11y-title" fontSize="1.5rem" /></Link>}
            {entry.delete && <Link href="#" onClick={() => onRevert(entry.id)}><ArrowUndoIcon title="a11y-title" fontSize="1.5rem" /></Link>}
            {!entry.delete && <Link href="#" onClick={() => onDelete(entry.id)}><TrashIcon title="a11y-title" fontSize="1.5rem" /></Link>}
        </div>
}

interface FormUrlEditorProps {
    urlList: FrontendUrlListEntry[]
    showReset: boolean
    onEditItem: (id: number) => void
    onDeleteItem: (id: number) => void
    onRevertItem: (id: number) => void
    onChangeItem: (id: number, v: string) => void
    onAddItem: () => void
    onReset: () => void
}

export const FormUrlEditor = ({ urlList, showReset, onEditItem, onDeleteItem, onRevertItem, onChangeItem, onAddItem, onReset }: FormUrlEditorProps) => {
    return <div>
        <div className="pl-10 w-[50rem] mt-3">
            <VStack gap="2">
                {urlList?.map((entry, index) =>
                    <Table.Row key={index}>
                        <UrlEntry entry={entry}
                            onEdit={onEditItem}
                            onDelete={onDeleteItem}
                            onRevert={onRevertItem}
                            onChange={onChangeItem}>
                        </UrlEntry>
                    </Table.Row>
                )}

            </VStack>
            <div className="flex flex-row gap-3 mt-3">
                {((urlList ?? []).length < 2500) ? <Link href="#" onClick={onAddItem}><PlusCircleFillIcon title="a11y-title" fontSize="1.5rem" />Legg til ny url
                </Link> : "Nådd varegrensen"}
                {showReset && <Link href="#" onClick={onReset}><ArrowUndoIcon title="a11y-title" fontSize="1.5rem" />tilbakestille
                </Link>}
            </div>

        </div>
    </div>
}

interface PlainTextUrlEditorProps {
    urlText: string
    showReset: boolean
    onValueChange: (urls: string) => void
    onReset: () => void
}

export const PlainTextUrlEditor = ({ urlText, onValueChange, onReset, showReset}: PlainTextUrlEditorProps) => {
    const [value, setValue] = useState(urlText)
    const getInvalidUrls = (urltxt: string) => {
        const invalidUrls = urltxt.split("\n").filter((url) => !isValidUrl(url))
        return invalidUrls
    }
    const error = getInvalidUrls(value).length ? `ulovlig format i url:  ${getInvalidUrls(value).map(it => `"${it}"`).join(", ")}` : undefined
    return <div>
        <div className="text-gray-600">En url tar én linje</div>
        <Textarea label={undefined} error={error} defaultValue={urlText}
            onChange={e => setValue(e.target.value)}
            onBlur={e => onValueChange(value)}></Textarea>
            {showReset && <Link className="mt-1" href="#" onClick={onReset}><ArrowUndoIcon title="a11y-title" fontSize="1.5rem" />tilbakestille
        </Link>}
    </div>
}



