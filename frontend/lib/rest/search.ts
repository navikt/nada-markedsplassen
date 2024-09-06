import { useEffect, useState } from "react"
import { fetchTemplate, searchUrl } from "./restApi"

export interface SearchOptions {
    // Freetext search
    text?: string
    // Filter on keyword
    keywords?: string[]
    // Filter on group
    groups?: string[]
    //Filter on team_id
    teamIDs?: string[]
    // Filter on enabled services
    services?: string[]
    // Filter on types
    types?: string[]

    limit?: number
    offset?: number
}

export interface SearchResult{
    results: any[]
}

const search = async (o: SearchOptions)=>{
    const url = searchUrl(o);
    return fetchTemplate(url)
}

export const useSearch = (o: SearchOptions)=>{
    const [data, setData] = useState<SearchResult|undefined>(undefined)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)


    useEffect(()=>{
        search(o).then((res)=> res.json())
        .then((data)=>
        {
            setError(null)
            setData(data)
        })
        .catch((err)=>{
            setError(err)
            setData(undefined)            
        }).finally(()=>{
            setLoading(false)
        })
    }, [o?JSON.stringify(o):""])

    return {data, loading, error}
}
