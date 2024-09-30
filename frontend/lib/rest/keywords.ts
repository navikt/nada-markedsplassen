import { useEffect, useState } from "react"
import { fetchKeywordsUrl, postTemplate, updateKeywordsUrl } from "./restApi"
import { KeywordsList, UpdateKeywordsDto } from "./generatedDto"

const fetchKeywords = async () => {
    const url = fetchKeywordsUrl()
    return fetch(url)
}

export const useFetchKeywords = () => {
    const [keywordsList, setKeywordsList] = useState<KeywordsList>({
        keywordItems: [],
    })
    useEffect(() => {
        fetchKeywords().then((res) => res.json())
            .then((keywordsList) => {
            setKeywordsList(keywordsList)
        })
            .catch((err) => {
            setKeywordsList({
                keywordItems: [],
            })
        })
    }, [])
    return keywordsList
}

export const updateKeywords = async (updateKeywordsDto: UpdateKeywordsDto) => {
    const url = updateKeywordsUrl()
    return postTemplate(url, updateKeywordsDto)
}