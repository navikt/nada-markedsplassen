import { useEffect, useState } from "react"
import { KeywordsList, UpdateKeywordsDto } from "./generatedDto"
import { fetchTemplate, postTemplate } from "./request"
import { buildFetchKeywordsUrl, buildUpdateKeywordsUrl } from "./apiUrl"

const fetchKeywords = async () => 
    fetchTemplate(buildFetchKeywordsUrl())

export const updateKeywords = async (updateKeywordsDto: UpdateKeywordsDto) => 
    postTemplate(buildUpdateKeywordsUrl(), updateKeywordsDto)

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

