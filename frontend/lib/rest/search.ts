import { SearchOptions } from "./generatedDto";
import { fetchTemplate } from "./request";
import { buildUrl, QueryParams } from "./apiUrl";
import { useQuery } from '@tanstack/react-query';

const searchPath = buildUrl('search')
const buildSearchUrl = (options: SearchOptions) => searchPath()(options as any as QueryParams)

export interface SearchResult{
    results: any[]| undefined
}

const search = async (o: SearchOptions)=>
    fetchTemplate(buildSearchUrl(o))

export const useSearch = (o: SearchOptions)=> useQuery<SearchResult, any>({
    queryKey: ['search', o], 
    queryFn: ()=>search(o)
})
