import { PublicMetabaseDashboardInput } from "./generatedDto"
import { postTemplate } from "./request"
import { buildUrl } from "./apiUrl"
const metabaseDashboardPath = buildUrl('metabaseDashboards')
const buildCreateMetabaseDashboardUrl = () => metabaseDashboardPath('new')()

export const createMetabaseDashboard = async (input: PublicMetabaseDashboardInput) =>
	postTemplate(buildCreateMetabaseDashboardUrl(), input)

//export const deleteMetabaseDashboard = async (id: string) => 
//    deleteTemplate(buildDeleteMetabaseDashboardUrlUrl(id))

