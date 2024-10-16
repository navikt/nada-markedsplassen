import { useEffect, useState } from "react"
import { SearchOptions } from "./generatedDto";
import { fetchTemplate } from "./request";
import { buildSearchUrl } from "./apiUrl";

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
        search(o).then((res)=> 
        {
            return res.json()
        }
    )
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
