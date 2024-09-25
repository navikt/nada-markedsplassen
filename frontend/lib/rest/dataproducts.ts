import { useEffect, useState } from "react"
import { createDataproductUrl, createDatasetUrl, deleteDataproductUrl, deleteDatasetUrl, deleteTemplate, fetchTemplate, getAccessiblePseudoDatasetsUrl, getDataproductUrl, getDatasetUrl, mapDatasetToServicesUrl, postTemplate, putTemplate, updateDataproductUrl, updateDatasetUrl } from "./restApi"
import { Dataproduct, DataproductWithDataset } from "./generatedDto"
import { da } from "date-fns/locale"

const getDataproduct = async (id: string) => {
    const url = getDataproductUrl(id)
    return fetchTemplate(url)
}

const getDataset = async (id: string) => {
    const url = getDatasetUrl(id)
    return fetchTemplate(url)
}

export const useGetDataproduct = (id: string, activeDataSetID?: string)=>{
    const [dataproduct, setDataproduct] = useState<DataproductWithDataset|null>(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)


    useEffect(()=>{
        if(!id) return
        getDataproduct(id).then((res)=> res.json())
        .then((dataproduct: DataproductWithDataset)=>
        {
            setError(null)
            setDataproduct(dataproduct)
        })
        .catch((err)=>{
            setError(err)
            setDataproduct(null)            
        }).finally(()=>{
            setLoading(false)
        })
    }, [id, activeDataSetID])

    return {dataproduct, loading, error}
}

export const useGetDataset = (id: string)=>{
    const [dataset, setDataset] = useState<any>(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)

    useEffect(()=>{
        if(!id) return
        getDataset(id).then((res)=> res.json())
        .then((dataset)=>
        {
            setError(null)
            setDataset(dataset)
        })
        .catch((err)=>{
            setError(err)
            setDataset(null)            
        }).finally(()=>{
            setLoading(false)
        })
    }, [id])

    return {dataset, loading, error}
}

export const createDataproduct = async (dp: any) => {
    const url = createDataproductUrl()
    return postTemplate(url, dp).then((res)=>res.json())
}

export const updateDataproduct = async (id: string, dp: any) => {
    const url = updateDataproductUrl(id)
    return putTemplate(url, dp).then((res)=>res.json())
}

export const deleteDataproduct = async (id: string) => {
    const url = deleteDataproductUrl(id)
    return deleteTemplate(url).then((res)=>res.json())
}

export const mapDatasetToServices = (id: string, services: string[])=>{
    const url = mapDatasetToServicesUrl(id)
    return postTemplate(url, {
        services}).then((res)=>res.json())
}

export const createDataset = async (dataset: any) => {
    const url = createDatasetUrl()
    return postTemplate(url, dataset).then((res)=>res.json())
}

export const deleteDataset = async (id: string) => {
    const url = deleteDatasetUrl(id)
    return deleteTemplate(url).then((res)=>res.json())
}

export const updateDataset = async (id: string, dataset: any) => {
    const url = updateDatasetUrl(id)
    return putTemplate(url, dataset).then((res)=>res.json())
}

export const getAccessiblePseudoDatasets = async () => {
    const url = getAccessiblePseudoDatasetsUrl()
    return fetchTemplate(url)
}

export const useGetAccessiblePseudoDatasets = ()=>{
    const [accessiblePseudoDatasets, setAccessiblePseudoDatasets] = useState<any[]>([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)


    useEffect(()=>{
        getAccessiblePseudoDatasets().then((res)=> res.json())
        .then((accessibleds)=>
        {
            setError(null)
            setAccessiblePseudoDatasets(accessibleds)
        })
        .catch((err)=>{
            setError(err)
            setAccessiblePseudoDatasets([])
        }).finally(()=>{
            setLoading(false)
        })
    }, [])

    return {accessiblePseudoDatasets: accessiblePseudoDatasets, loading, error}
}
