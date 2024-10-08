import { useEffect, useState } from "react"
import { createInsightProductUrl, deleteTemplate, fetchTemplate, getInsightProductUrl, postTemplate, putTemplate, updateInsightProductUrl } from "./restApi"
import { InsightProduct, NewInsightProduct, UpdateInsightProductDto } from "./generatedDto"

const getInsightProduct = async (id: string) => {
    const url = getInsightProductUrl(id)
    return fetchTemplate(url)
}

export const useGetInsightProduct = (id: string)=>{
    const [insightProduct, setInsightProduct] = useState<InsightProduct|null>(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<Error| undefined>(undefined)


    useEffect(()=>{
        if(!id) return
        getInsightProduct(id).then((res)=> res.json())
        .then((data)=>
        {
            setError(undefined)
            setInsightProduct(data)
        })
        .catch((err)=>{
            setError(err)
            setInsightProduct(null)            
        }).finally(()=>{
            setLoading(false)
        })
    }, [id])

    return {insightProduct, loading, error}
}

export const createInsightProduct = async (insp: NewInsightProduct) => {
    const url = createInsightProductUrl()
    return postTemplate(url, insp).then((res)=>res.json())
}

export const updateInsightProduct = async (id: string, insp: UpdateInsightProductDto) => {
    const url = updateInsightProductUrl(id)
    return putTemplate(url, insp).then((res)=>res.json())
}

export const deleteInsightProduct= async (id: string) => {
    const url = updateInsightProductUrl(id)
    return deleteTemplate(url).then((res)=>res.json())
}