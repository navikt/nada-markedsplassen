import { PublicMetabaseDashboardInput, PublicMetabaseDashboardEditInput, PublicMetabaseDashboardOutput } from "./generatedDto"
import { fetchTemplate, deleteTemplate, postTemplate, HttpError, putTemplate } from "./request"
import { buildUrl } from "./apiUrl"
import { useQuery } from '@tanstack/react-query'

const metabaseDashboardPath = buildUrl('metabaseDashboards')
const buildGetMetabaseDashboardUrl = (id: string) => metabaseDashboardPath(id)()
const buildCreateMetabaseDashboardUrl = () => metabaseDashboardPath('new')()
const buildEditMetabaseDashboardUrl = (id: string) => metabaseDashboardPath(id)()
const buildDeleteMetabaseDashboardUrl = (id: string) => metabaseDashboardPath(id)()

export const createMetabaseDashboard = async (input: PublicMetabaseDashboardInput) =>
	postTemplate(buildCreateMetabaseDashboardUrl(), input)

export const updateMetabaseDashboard = async (id: string, input: PublicMetabaseDashboardEditInput) => 
  putTemplate(buildEditMetabaseDashboardUrl(id), input)

export const deleteMetabaseDashboard = async (id: string) => 
	deleteTemplate(buildDeleteMetabaseDashboardUrl(id))

export const getMetabaseDashboard = async (id: string) =>
  fetchTemplate(buildGetMetabaseDashboardUrl(id))

export const useGetMetabaseDashboardUrl= (id: string)=>
    useQuery<PublicMetabaseDashboardOutput, HttpError>({
        queryKey: ['metabaseDashboard', id], 
        queryFn: ()=> getMetabaseDashboard(id)
    })
