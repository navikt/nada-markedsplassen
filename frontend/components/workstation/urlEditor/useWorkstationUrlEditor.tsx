import { ArrowUndoIcon, ExternalLinkIcon, FileTextIcon, PencilBoardIcon, PencilIcon, PlusCircleFillIcon, TrashIcon } from "@navikt/aksel-icons"
import { Label, Link, List, Stack, Table, Tabs, Textarea, TextField, VStack } from "@navikt/ds-react"
import { useWorkstationURLList } from "../queries"
import { useEffect, useState } from "react"
import GlobalAllowListSelector from "./GlobalAllowListSelector"
import { ListItem } from "@mui/material"

export interface UrlEntryProps {
    id: number
    url: string
    isEditing: boolean
    isDeleted: boolean
    isAdded: boolean
    onEdit: (id: number) => void
    onChange: (id: number, v: string) => void
    onDelete: (id: number) => void
    onRevert: (id: number) => void
}

export const textColorDeleted = "text-red-400"

const isValidUrl = (url: string) => {
    const urlPattern = /^((\*|\*?[a-zA-Z0-9-]+)\.)+[a-zA-Z0-9-]{2,}(\/(\*|[a-zA-Z0-9-._~:/?#[\]@!$&'()*+,;=]*))*$/;
    return !url ? true : urlPattern.test(url);
}

const UrlEntry = ({ id, url, isEditing, isDeleted, isAdded, onChange, onEdit, onDelete, onRevert }: UrlEntryProps) => {
    const [value, setValue] = useState(url)
    const urlStyle = isDeleted ? `${textColorDeleted} line-through` : ""
    return (isEditing || isAdded) ?
        <div className="w-96 flex items-center ml-1">
            <TextField error={isValidUrl(value) ? undefined : "ulovlig format"} className="w-[10rem] mr-2" size="small" label="" value={value}
                onChange={e => setValue(e.target.value)}
                onBlur={e => onChange(id, value)}></TextField>
            {!isAdded && <Link href="#" onClick={() => onRevert(id)}><ArrowUndoIcon title="a11y-title" fontSize="1.5rem" /></Link>}
            <Link href="#" onClick={() => onDelete(id)}><TrashIcon title="a11y-title" fontSize="1.5rem" /></Link>
        </div>
        :
        <div className="w-96 flex items-center ml-3">
            <div className={`${urlStyle} mr-2`}>{url}</div>
            {!isDeleted && <Link href="#" onClick={() => onEdit(id)}><PencilIcon title="a11y-title" fontSize="1.5rem" /></Link>}
            {isDeleted && <Link href="#" onClick={() => onRevert(id)}><ArrowUndoIcon title="a11y-title" fontSize="1.5rem" /></Link>}
            {!isDeleted && <Link href="#" onClick={() => onDelete(id)}><TrashIcon title="a11y-title" fontSize="1.5rem" /></Link>}
        </div>
}

interface UrlListPlainTextProps {
    urlText: string
    onValueChange: (urls: string) => void
}

const UrlListPlainText = ({ urlText, onValueChange }: UrlListPlainTextProps) => {
    const [value, setValue] = useState(urlText)
    const getInvalidUrls = (urltxt: string) => {
        const invalidUrls = urltxt.split("\n").filter((url) => !isValidUrl(url))
        return invalidUrls
    }
    const error = getInvalidUrls(value).length ? `ulovlig format i url:  ${getInvalidUrls(value).map(it => `"${it}"`).join(", ")}` : undefined
    const urlsChanged = value !== urlText
    return <div>
        <Textarea label={undefined} error={error} defaultValue={urlText}
            onChange={e => setValue(e.target.value)}
            onBlur={e => onValueChange(value)}></Textarea>
            {urlsChanged && <Link className="mt-1" href="#" onClick={() => setValue(urlText)}><ArrowUndoIcon title="a11y-title" fontSize="1.5rem" />tilbakestille
        </Link>}
    </div>
}

interface UrlListStandardProps {
    urlList: any[]
    listChanged: boolean
    onEditItem: (id: number) => void
    onDeleteItem: (id: number) => void
    onRevertItem: (id: number) => void
    onChangeItem: (id: number, v: string) => void
    onAddItem: () => void
    onReset: () => void
}

const UrlListStandard = ({ urlList, listChanged, onEditItem, onDeleteItem, onRevertItem, onChangeItem, onAddItem, onReset }: UrlListStandardProps) => {
    return <div>
        <div className="pl-10 w-[28rem] mt-3">
            <VStack gap="2">
                {urlList?.map((entry, index) =>
                    <Table.Row key={index}>
                        <UrlEntry id={entry.id} url={entry.url} isEditing={entry.isEditing} isDeleted={entry.isDeleted} isAdded={entry.isAdded}
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
                {listChanged && <Link href="#" onClick={onReset}><ArrowUndoIcon title="a11y-title" fontSize="1.5rem" />tilbakestille
                </Link>}
            </div>

        </div>
    </div>
}

const useWorkstationUrlEditor = () => {
    const { data, isLoading, error } = useWorkstationURLList()
    const [urlChangeList, setUrlChangeList] = useState<any[] | null>(null)
    const [urlPlainText, setUrlPlainText] = useState<string>("")
    const updateUrl = (id: number, v: string | undefined = undefined, isDeleted: boolean | undefined = undefined, isEditing: boolean | undefined = undefined, reset: boolean = false) => {
        const target = urlChangeList?.find((entry) => entry.id === id)
        if (!!target) {
            if (reset) {
                target.url = target.defaultUrl
                target.isEditing = false
                target.isDeleted = false
                console.log(reset)
            }
            else {
                target.url = v !== undefined ? v : target.url
                target.isEditing = isEditing !== undefined ? isEditing : target.isEditing
            }
            console.log(isDeleted)
            if (isDeleted !== undefined) {
                if (target.isAdded) {
                    urlChangeList?.splice(urlChangeList.indexOf(target), 1)
                } else {
                    target.isDeleted = isDeleted
                    target.isEditing = false
                }
            }
            setUrlChangeList([...urlChangeList!])
        }
    }
    const addUrl = () => {
        const newUrlList = urlChangeList?.filter(it => !it.isAdded || !!it.url) || []

        newUrlList.push({
            id: Math.max(...newUrlList.map(it => it.id), 0) + 1,
            url: "",
            isEditing: true,
            isDeleted: false,
            isAdded: true
        })
        setUrlChangeList(newUrlList)
    }

    const resetAll = () => {
        const newUrlList = urlChangeList?.filter(it => !it.isAdded)
        newUrlList?.forEach(it => updateUrl(it.id, undefined, false, false, true))
        setUrlChangeList(newUrlList || null)
    }

    const listChanged = !!urlChangeList?.some(it => it.isDeleted || (it.isAdded && !!it.url) || (!it.isAdded && it.url !== it.defaultUrl))

    useEffect(() => {
        const testUrl = ["vg.no", "facebook.com"]
        if (data?.urlAllowList) {
            const newUrlList = testUrl.map((url: string, index: number) => ({
                id: index,
                defaultUrl: url,
                url: url,
                isDeleted: false,
                isAdded: false,
                isEditing: false
            }));
            setUrlChangeList(newUrlList);

            const newUrlPlainText = testUrl.join("\n")
            setUrlPlainText(newUrlPlainText)
        }
    }, [data]);
    return {
        listChanged: listChanged,
        urlList: urlChangeList?.map(it => it.url) || [],
        isLoading: isLoading,
        isGlobalAllowListDisabled: data?.disableGlobalAllowList || false,
        urlEditor: () => (
            <div>
                {
                    <GlobalAllowListSelector value={false} urls={["vg.no", "facebook.com"]} optIn={false} onChange={function (value: boolean): void {
                        throw new Error("Function not implemented.")
                    }}></GlobalAllowListSelector>
                }
                <div className="mt-3">
                    <Label>Oppgi hvilke internett-URLer du vil åpne mot</Label>
                    <p className="mt-0 text-gray-600">
                        <br />URL-er må være inline med google cloud url-format, og jokertegn(*) er tillatt. <br />F.eks. example.com, *.example.com.
                        <Link className="ml-3" target="_blank" href="https://cloud.google.com/secure-web-proxy/docs/url-list-syntax-reference">
                            Les mer: <ExternalLinkIcon />
                        </Link>
                    </p>

                    <Tabs defaultValue="standard">
                        <Tabs.List>
                            <Tabs.Tab
                                value="standard"
                                label="Skjemainndata"
                                icon={<PencilBoardIcon aria-hidden />}
                            />
                            <Tabs.Tab
                                value="plaintext"
                                label="Klartekst"
                                icon={<FileTextIcon aria-hidden />}
                            />
                        </Tabs.List>
                        <Tabs.Panel value="standard" className="w-[30rem] bg-gray-50 p-4">
                            <UrlListStandard urlList={urlChangeList || []} listChanged={listChanged}
                                onEditItem={id => updateUrl(id, undefined, undefined, true)}
                                onDeleteItem={id => updateUrl(id, undefined, true)}
                                onRevertItem={id => updateUrl(id, undefined, undefined, undefined, true)}
                                onChangeItem={(id, v) => updateUrl(id, v)}
                                onAddItem={addUrl}
                                onReset={resetAll}>
                            </UrlListStandard>
                        </Tabs.Panel>
                        <Tabs.Panel value="plaintext" className="w-[30rem] bg-gray-50 p-4">
                            <UrlListPlainText urlText={urlPlainText} onValueChange={setUrlPlainText} ></UrlListPlainText>
                        </Tabs.Panel>
                    </Tabs>


                </div>
            </div>
        )
    }
}

export default useWorkstationUrlEditor