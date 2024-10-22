import { InsightProduct, NewInsightProduct, UpdateInsightProductDto } from "./generatedDto"
import { deleteTemplate, fetchTemplate, HttpError, postTemplate, putTemplate } from "./request"
import { buildUrl } from "./apiUrl"
import { useQuery } from "react-query"

const insightProductPath = buildUrl('insightProducts')
const buildGetInsightProductUrl = (id: string) => insightProductPath(id)()
const buildCreateInsightProductUrl = () => insightProductPath('new')()
const buildUpdateInsightProductUrl = (id: string) => insightProductPath(id)()
const buildDeleteInsightProductUrl = (id: string) => insightProductPath(id)()

const getInsightProduct = async (id: string) => 
    fetchTemplate(buildGetInsightProductUrl(id))

export const createInsightProduct = async (insp: NewInsightProduct) => 
    postTemplate(buildCreateInsightProductUrl(), insp)

export const updateInsightProduct = async (id: string, insp: UpdateInsightProductDto) => 
    putTemplate(buildUpdateInsightProductUrl(id), insp)

export const deleteInsightProduct= async (id: string) => 
    deleteTemplate(buildDeleteInsightProductUrl(id))

export const useGetInsightProduct = (id: string)=>
    useQuery<InsightProduct, HttpError>(['insightProduct', id], ()=>getInsightProduct(id))

