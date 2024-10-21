import { QueryPolly } from "./generatedDto"
import { fetchTemplate, HttpError } from "./request"
import { buildPath } from "./apiUrl"
import { useQuery } from "react-query"

const pollyPath = buildPath('polly')
const buildSearchPollyUrl = (query?: string) => pollyPath()({query: query || ''})

const searchPolly = async (query?: string) => 
    fetchTemplate(buildSearchPollyUrl(query))


export const useSearchPolly = (query?: string) => useQuery<QueryPolly[], HttpError>(['polly', query], ()=>searchPolly(query))
