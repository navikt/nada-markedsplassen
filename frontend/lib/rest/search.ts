import { useEffect, useState } from "react"
import { SearchOptions } from "./generatedDto";
import { fetchTemplate } from "./request";
import { buildPath, QueryParams } from "./apiUrl";
import { useQuery } from "react-query";

const searchPath = buildPath('search')
const buildSearchUrl = (options: SearchOptions) => searchPath()(options as any as QueryParams)

export interface SearchResult{
    results: any[]| undefined
}

const search = async (o: SearchOptions)=>
    fetchTemplate(buildSearchUrl(o))

export const useSearch = (o: SearchOptions)=> useQuery<SearchResult, any>(['search', o], ()=>search(o))
