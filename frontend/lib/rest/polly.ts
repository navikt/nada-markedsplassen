import { useEffect, useState } from "react"
import { apiUrl, fetchTemplate } from "./restApi"


export const searchPolly = async (query?: string) => {
    const url= `${apiUrl()}/polly?query=${query}`
    return fetchTemplate(url)
}

export const useSearchPolly = (query?: string) => {
    const [searchResult, setSearchResult] = useState<any[]>([]);
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