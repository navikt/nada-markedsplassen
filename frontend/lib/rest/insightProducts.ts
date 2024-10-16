import { useEffect, useState } from "react"
import { InsightProduct, NewInsightProduct, UpdateInsightProductDto } from "./generatedDto"
import { deleteTemplate, fetchTemplate, postTemplate, putTemplate } from "./request"
import { buildCreateInsightProductUrl, buildDeleteInsightProductUrl, buildGetInsightProductUrl, buildUpdateInsightProductUrl } from "./apiUrl"

const getInsightProduct = async (id: string) => 
    fetchTemplate(buildGetInsightProductUrl(id))

export const createInsightProduct = async (insp: NewInsightProduct) => 
    postTemplate(buildCreateInsightProductUrl(), insp)

export const updateInsightProduct = async (id: string, insp: UpdateInsightProductDto) => 
    putTemplate(buildUpdateInsightProductUrl(id), insp)

export const deleteInsightProduct= async (id: string) => 
    deleteTemplate(buildDeleteInsightProductUrl(id))

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

