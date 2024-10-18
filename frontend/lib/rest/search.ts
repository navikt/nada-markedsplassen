import { useEffect, useState } from "react"
import { SearchOptions } from "./generatedDto";
import { fetchTemplate } from "./request";
import { buildPath, QueryParams } from "./apiUrl";

const searchPath = buildPath('search')
const buildSearchUrl = (options: SearchOptions) => searchPath()(options as any as QueryParams)

export interface SearchResult{
    results: any[]| undefined
}

const search = async (o: SearchOptions)=>
    fetchTemplate(buildSearchUrl(o))

export const useSearch = (o: SearchOptions)=>{
    const [data, setData] = useState<SearchResult|undefined>(undefined)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)


    useEffect(()=>{
        search(o)
        .then((data)=>
        {
            setError(null)
            setData(data)
        })
        .catch((err)=>{
            console.log(err)
            setError(err)
            setData(undefined)            
        }).finally(()=>{
            setLoading(false)
        })
    }, [o?JSON.stringify(o):""])

    return {data, loading, error}
}
