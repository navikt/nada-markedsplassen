import { useEffect, useState } from "react"
import { fetchTemplate, searchUrl } from "./restApi"
import { SearchOptions } from "./generatedDto";

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
