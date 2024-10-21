import { KeywordsList, UpdateKeywordsDto } from "./generatedDto"
import { fetchTemplate, HttpError, postTemplate } from "./request"
import { buildPath } from "./apiUrl"
import { useQuery } from "react-query"

const keywordsPath = buildPath('keywords')
const buildFetchKeywordsUrl = () => keywordsPath()()
const buildUpdateKeywordsUrl = () => keywordsPath()()

const fetchKeywords = async () =>
    fetchTemplate(buildFetchKeywordsUrl())

export const updateKeywords = async (updateKeywordsDto: UpdateKeywordsDto) =>
    postTemplate(buildUpdateKeywordsUrl(), updateKeywordsDto)

export const useFetchKeywords = () =>
    useQuery<KeywordsList, HttpError>(['keywords'], fetchKeywords)
