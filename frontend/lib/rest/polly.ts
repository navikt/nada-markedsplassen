import { useEffect, useState } from "react"
import { QueryPolly } from "./generatedDto"
import { fetchTemplate } from "./request"
import { buildPath } from "./apiUrl"

const pollyPath = buildPath('polly')
const buildSearchPollyUrl = (query?: string) => pollyPath()({query: query || ''})

const searchPolly = async (query?: string) => 
    fetchTemplate(buildSearchPollyUrl(query))

export const useSearchPolly = (query?: string) => {
    const [searchResult, setSearchResult] = useState<QueryPolly[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    useEffect(() => {
        if((query?.length??0) < 3) return
        setLoading(true);
        searchPolly(query)
            .then((searchPollyDto) => {
            setError(null);
            setSearchResult(searchPollyDto);
        })
            .catch((err) => {
            setError(err);
            setSearchResult([])
        }).finally(() => {
            setLoading(false);
        })
    }, [query])
    return { searchResult, loading, searchError: error };
}