import { useEffect, useState } from "react"
import { buildSearchPollyUrl} from "./apiUrl"
import { QueryPolly } from "./generatedDto"
import { fetchTemplate } from "./request"


const searchPolly = async (query?: string) => 
    fetchTemplate(buildSearchPollyUrl(query))

export const useSearchPolly = (query?: string) => {
    const [searchResult, setSearchResult] = useState<QueryPolly[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    useEffect(() => {
        if((query?.length??0) < 3) return
        setLoading(true);
        searchPolly(query).then((res) => res.json())
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