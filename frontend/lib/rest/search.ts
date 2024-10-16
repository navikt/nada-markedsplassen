import { useEffect, useState } from "react"
import { SearchOptions } from "./generatedDto";
import { fetchTemplate } from "./request";
import { buildPath } from "./apiUrl";

const searchPath = buildPath('search')
const buildSearchUrl = (options: SearchOptions) => searchPath()({
  keywords: options.keywords,
  groups: options.groups,
  teamIDs: options.teamIDs,
  services: options.services,
  types: options.types,
  text: options.text,
  limit: options.limit,
  offset: options.offset,
})

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
