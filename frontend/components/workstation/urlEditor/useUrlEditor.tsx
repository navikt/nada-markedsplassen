import { ExternalLinkIcon } from "@navikt/aksel-icons"
import { Label, Link } from "@navikt/ds-react"
import { useWorkstationOptions, useWorkstationURLList } from "../queries"
import { useEffect, useState } from "react"
import { FormUrlEditor, FrontendUrlListEntry, isAdded, PlainTextUrlEditor } from "./urlEditor"
import GlobalAllowListField from "./globalAllowListField"
import GlobalDenyListField from "./globalDenyListField"

export const textColorDeleted = "text-red-400"

const useWorkstationUrlEditor = () => {
    const { data: backendUrlList, isLoading, error } = useWorkstationURLList()
    const [frontendUrlList, setFrontendUrlList] = useState<FrontendUrlListEntry[] | null>(null)
    const [keepGlobalAllowList, setKeepGlobalAllowList] = useState<boolean>(true)
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

    const setUrlEntry = (id: number, url: string) => {
        const target = frontendUrlList?.find((entry) => entry.id === id)
        if (!!target) {
            target.url = url
            target.edit = false
            setFrontendUrlList([...frontendUrlList!])
        }
    }

    const revertUrlEntry = (id: number) => {
        const target = frontendUrlList?.find((entry) => entry.id === id)
        if (!!target) {
            target.url = target.defaultUrl
            target.edit = false
            target.delete = false
            setFrontendUrlList([...frontendUrlList!])
        }
    }

    const updateUrlListWithPlainText = (plainTextUrls: string) => {
        // Remove empty lines and trim whitespace
        const urlsFromText = plainTextUrls.split('\n')
            .map(url => url.trim())
            .filter(url => url.length > 0);

        // Create new list based on text input
        const newUrlList = urlsFromText.map((url, index) => ({
                id: index,
                defaultUrl: "",
                url: url,
                edit: true,
                delete: false,
            }))

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
        addUrlEntry,
        deleteUrlEntry,
        editUrlEntry,
        setUrlEntry,
        revertUrlEntry,
        updateUrlListWithPlainText,
        resetEditor,
        urlEditor: () => (
            <div>
                <GlobalAllowListField optIn={keepGlobalAllowList} onChange={(value: boolean) => {
                    setKeepGlobalAllowList(value)
                }} urls={options.data?.globalURLAllowList || []}></GlobalAllowListField>
                <GlobalDenyListField urls={backendUrlList?.globalDenyList || []} />
            </div>
        )
    }
}

export default useWorkstationUrlEditor