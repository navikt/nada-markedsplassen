import { ExternalLinkIcon, FileTextIcon, PencilBoardIcon } from "@navikt/aksel-icons"
import { Label, Link, Tabs } from "@navikt/ds-react"
import { useWorkstationOptions, useWorkstationURLList } from "../queries"
import { useEffect, useState } from "react"
import { FormUrlEditor, FrontendUrlListEntry, isAdded, PlainTextUrlEditor } from "./urlEditor"
import GlobalAllowListField from "./globalAllowListField"


export const textColorDeleted = "text-red-400"


const useWorkstationUrlEditor = () => {
    const { data: backendUrlList, isLoading, error } = useWorkstationURLList()
    const [frontendUrlList, setFrontendUrlList] = useState<FrontendUrlListEntry[] | null>(null)
    const [keepGlobalAllowList, setKeepGlobalAllowList] = useState<boolean>(true)
    const [currentEditor, setCurrentEditor] = useState<"form" | "plaintext">("form")
    const options = useWorkstationOptions()

    const addUrlEntry = () => {
        //only have one empty entry at most
        const newUrlList = frontendUrlList?.filter(it => !isAdded(it) || !!it.url) || []
        newUrlList.push({
            id: Math.max(...newUrlList.map(it => it.id), 0) + 1,
            url: "",
            defaultUrl: "",
            edit: true,
            delete: false,
        })
        setFrontendUrlList(newUrlList)
    }

    const deleteUrlEntry = (id: number) => {
        const target = frontendUrlList?.find((entry) => entry.id === id)
        if (!!target) {
            if (isAdded(target)) {
                frontendUrlList?.splice(frontendUrlList.indexOf(target), 1)
            } else {
                target.delete = true
                target.edit = false
            }
            setFrontendUrlList([...frontendUrlList!])
        }
    }

    const editUrlEntry = (id: number) => {
        const target = frontendUrlList?.find((entry) => entry.id === id)
        if (!!target) {
            target.edit = true
            setFrontendUrlList([...frontendUrlList!])
        }
    }

    const setUrlEntry = (id: number, v: string) => {
        const target = frontendUrlList?.find((entry) => entry.id === id)
        if (!!target) {
            target.url = v
            setFrontendUrlList([...frontendUrlList!])
        }
    }

    const revertUrlEntry = (id: number) => {
        const target = frontendUrlList?.find((entry) => entry.id === id)
        if (!!target) {
            if (isAdded(target)) {
                frontendUrlList?.splice(frontendUrlList.indexOf(target), 1)
            } else {
                target.url = target.defaultUrl
                target.edit = false
                target.delete = false
            }
            setFrontendUrlList([...frontendUrlList!])
        }
    }

    const updateUrlListWithPlainText = (text: string) => {
        console.log(text)
        const plainTextUrlList = text.split("\n").map(url => url.trim());

        //keep entries that in plain text
        let newUrlList = frontendUrlList?.filter(it => !isAdded(it) || plainTextUrlList.includes(it.url)).map(it => {
            const deleted = !plainTextUrlList.includes(it.url)
            return {
                ...it,
                delete: deleted,
                edit: deleted ? false : it.edit
            }
        })

        console.log(newUrlList)

        const assignid = Math.max(...newUrlList?.map(it => it.id) || [], 0) + 1
        //concat new entries
        newUrlList = newUrlList?.concat(
            plainTextUrlList
                .filter(url => !frontendUrlList?.some(it => it.url === url))
                .map((url, index) => ({
                    id: assignid + index,
                    defaultUrl: "",
                    url: url,
                    edit: true,
                    delete: false,
                })))

        setFrontendUrlList(newUrlList || []);
    }

    const resetEditor = () => {
        const newUrlList = frontendUrlList?.filter(it => !isAdded(it))
        newUrlList?.forEach(it => revertUrlEntry(it.id))
        setFrontendUrlList(newUrlList || null)
    }

    const listChanged = (() => {
        if (isLoading || !backendUrlList) {
            return false
        }

        console.log(backendUrlList.disableGlobalAllowList, keepGlobalAllowList)
        let changed = backendUrlList.disableGlobalAllowList !== !keepGlobalAllowList
        changed ||= !!frontendUrlList?.some(it => it.delete || (isAdded(it) && !!it.url) || (!isAdded(it) && it.url !== it.defaultUrl))
        return changed
    })()

    useEffect(() => {
        if (backendUrlList?.urlAllowList) {
            const newUrlList = backendUrlList.urlAllowList.map((url: string, index: number) => ({
                id: index,
                defaultUrl: url,
                url: url,
                delete: false,
                edit: false,
            }));
            setFrontendUrlList(newUrlList);
            setKeepGlobalAllowList(!backendUrlList.disableGlobalAllowList)
        }
    }, [backendUrlList]);

    return {
        listChanged: listChanged,
        urlList: frontendUrlList?.filter(it=>!it.delete).map(it => it.url)|| [],
        isLoading: isLoading,
        keepGlobalAllowList: keepGlobalAllowList,
        urlEditor: () => (
            <div>
                {
                    <GlobalAllowListField optIn={keepGlobalAllowList} onChange={(value: boolean) => {
                        setKeepGlobalAllowList(value)
                    }} urls={options.data?.globalURLAllowList || []}></GlobalAllowListField>
                }
                <div className="mt-3">
                    <Label>Oppgi hvilke internett-URLer du vil åpne mot</Label>
                    <p className="mt-0 text-gray-600">
                        <br />URL-er må være inline med google cloud url-format, og jokertegn(*) er tillatt. <br />F.eks. example.com, *.example.com.
                        <Link className="ml-3" target="_blank" href="https://cloud.google.com/secure-web-proxy/docs/url-list-syntax-reference">
                            Les mer: <ExternalLinkIcon />
                        </Link>
                    </p>

                    <Tabs defaultValue="form" value={currentEditor} onChange={v => setCurrentEditor(v as "form" | "plaintext")}>
                        <Tabs.List>
                            <Tabs.Tab
                                value="form"
                                label="Skjemainndata"
                                icon={<PencilBoardIcon aria-hidden />}
                            />
                            <Tabs.Tab
                                value="plaintext"
                                label="Klartekst"
                                icon={<FileTextIcon aria-hidden />}
                            />
                        </Tabs.List>
                        <Tabs.Panel value="form" className="w-[50rem] bg-gray-50 p-4">
                            <FormUrlEditor urlList={frontendUrlList || []} showReset={listChanged}
                                onEditItem={editUrlEntry}
                                onDeleteItem={deleteUrlEntry}
                                onRevertItem={resetEditor}
                                onChangeItem={setUrlEntry}
                                onAddItem={addUrlEntry}
                                onReset={resetEditor}>
                            </FormUrlEditor>
                        </Tabs.Panel>
                        <Tabs.Panel value="plaintext" className="w-[50rem] bg-gray-50 p-4">
                            <PlainTextUrlEditor urlText={frontendUrlList?.filter(it => !it.delete).
                                map(it => it.url).join("\n") || ""} onValueChange={updateUrlListWithPlainText}
                                onReset={resetEditor}
                                showReset={listChanged} ></PlainTextUrlEditor>
                        </Tabs.Panel>
                    </Tabs>


                </div>
            </div>
        )
    }
}

export default useWorkstationUrlEditor