import { PublicMetabaseDashboardInput } from "./generatedDto"
import { deleteTemplate, postTemplate } from "./request"
import { buildUrl } from "./apiUrl"
const metabaseDashboardPath = buildUrl('metabaseDashboards')
const buildCreateMetabaseDashboardUrl = () => metabaseDashboardPath('new')()
const buildDeleteMetabaseDashboardUrl = (id: string) => metabaseDashboardPath(id)()

export const createMetabaseDashboard = async (input: PublicMetabaseDashboardInput) =>
	postTemplate(buildCreateMetabaseDashboardUrl(), input)

export const deleteMetabaseDashboard = async (id: string) => 
	deleteTemplate(buildDeleteMetabaseDashboardUrl(id))

