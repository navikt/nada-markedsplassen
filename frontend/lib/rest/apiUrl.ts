import { SearchOptions } from "./generatedDto"

const isServer = typeof window === 'undefined'

const paramToString = (param: string | string[] | undefined) => 
  Array.isArray(param) ? param.map(it=> encodeURIComponent(it)).join(',') : param!== undefined? encodeURIComponent(param): param

const queryParamsToString = (queryParams?: Record<string, string|string[]|number|undefined>) => {
  if (!queryParams) return ''
  
  const paramString = Object.entries(queryParams).filter(it=> it[0] && it[1])
    .map(([key, value]) => `${key}=${paramToString(value?.toString())}`)
    .join('&')

  return paramString ? `?${paramString}` : ''
}

const buildUrl = (baseUrl: string) => (path: string) => (pathParam?: string) => (queryParams?: Record<string, string|string[]|number|undefined>) =>
  `${baseUrl}/${path}${pathParam ? `/${encodeURIComponent(pathParam)}` : ''}${queryParamsToString(queryParams)}`

const localBaseUrl = buildUrl('http://localhost:8080/api')
const asServerBaseUrl = buildUrl('http://nada-backend/api')
const asClientBaseUrl = buildUrl('/api')

const buildPath = process.env.NEXT_PUBLIC_ENV === 'development' ? localBaseUrl :
  isServer ? asServerBaseUrl : asClientBaseUrl

const dataproductPath = buildPath('dataproducts')
export const buildFetchDataproductUrl = (id: string) => dataproductPath(id)()
export const buildCreateDataproductUrl = () => dataproductPath('new')()
export const buildUpdateDataproductUrl = (id: string) => dataproductPath(id)()
export const buildDeleteDataproductUrl = (id: string) => dataproductPath(id)()


const datasetPath = buildPath('datasets')
export const buildFetchDatasetUrl = (id: string) => datasetPath(id)()
export const buildMapDatasetToServicesUrl = (datasetId: string) => `${datasetPath(datasetId)()}/map`
export const buildCreateDatasetUrl = () => datasetPath('new')()
export const buildDeleteDatasetUrl = (id: string) => datasetPath(id)()
export const buildUpdateDatasetUrl = (id: string) => datasetPath(id)()
export const buildGetAccessiblePseudoDatasetsUrl = () => datasetPath('pseudo/accessible')()

const storyPath = buildPath('stories')
export const buildCreateStoryUrl = () => storyPath('new')()
export const buildUpdateStoryUrl = (id: string) => storyPath(id)()
export const buildDeleteStoryUrl = (id: string) => storyPath(id)()
export const buildFetchStoryMetadataUrl = (id: string) => storyPath(id)()


const joinableViewPath = buildPath('pseudo/joinable')
export const buildGetJoinableViewUrl = (id: string) => joinableViewPath(id)()
export const buildCreateJoinableViewsUrl = () => joinableViewPath('new')()
export const buildGetJoinableViewsForUserUrl = () => joinableViewPath()()

const productAreasPath = buildPath('productareas')
export const buildGetProductAreasUrl = () => productAreasPath()()
export const buildGetProductAreaUrl = (id: string) => productAreasPath(id)()

const userDataPath = buildPath('userData')
export const buildFetchUserDataUrl = () => userDataPath()()

const slackPath = buildPath('slack')
export const buildIsValidSlackChannelUrl = (channel: string) => slackPath('isValidChannel')({channel: channel})

const insightProductPath = buildPath('insightProducts')
export const buildGetInsightProductUrl = (id: string) => insightProductPath(id)()
export const buildCreateInsightProductUrl = () => insightProductPath('new')()
export const buildUpdateInsightProductUrl = (id: string) => insightProductPath(id)()
export const buildDeleteInsightProductUrl = (id: string) => insightProductPath(id)()

const keywordsPath = buildPath('keywords')
export const buildFetchKeywordsUrl = () => keywordsPath()()
export const buildUpdateKeywordsUrl = () => keywordsPath()()

const accessRequestsPath = buildPath('accessRequests')
export const buildFetchAccessRequestUrl = (datasetId: string) => accessRequestsPath()({datasetId: datasetId})
export const buildCreateAccessRequestUrl = () => accessRequestsPath('new')()
export const buildDeleteAccessRequestUrl = (id: string) => accessRequestsPath(id)()
export const buildUpdateAccessRequestUrl = (id: string) => accessRequestsPath(id)()

const processAccessRequestsPath = buildPath('accessRequests/process')
export const buildApproveAccessRequestUrl = (accessRequestId: string) => processAccessRequestsPath(accessRequestId)({action: 'approve'})
export const buildDenyAccessRequestUrl = (accessRequestId: string, reason: string) => processAccessRequestsPath(accessRequestId)({action: 'deny', reason: reason})


const bigqueryPath = buildPath('bigquery')
export const buildFetchBQDatasetsUrl = (projectId: string) => bigqueryPath('datasets')({projectId: projectId})
export const buildFetchBQTablesUrl = (projectId: string, datasetId: string) => bigqueryPath('tables')({projectId: projectId, datasetId: datasetId})
export const buildFetchBQColumnsUrl = (projectId: string, datasetId: string, tableId: string) =>  bigqueryPath('columns')({projectId: projectId, datasetId: datasetId, tableId: tableId})

const teamKatalogenPath = buildPath('teamkatalogen')
export const buildSearchTeamKatalogenUrl = (gcpGroups?: string[]) => 
  teamKatalogenPath()({gcpGroups: gcpGroups?.length ? gcpGroups.map(group => `gcpGroups=${encodeURIComponent(group)}`).join('&') : ''})

const pollyPath = buildPath('polly')
export const buildSearchPollyUrl = (query?: string) => pollyPath()({query: query || ''})

const accessPath = buildPath('accesses')
export const buildGrantAccessUrl = () => accessPath('grant')()
export const buildRevokeAccessUrl = (accessId: string) => accessPath('revoke')({accessId: accessId})

const workstationsPath = buildPath('workstations')
export const buildGetWorkstationUrl = () => workstationsPath()()
export const buildEnsureWorkstationUrl = () => workstationsPath()()
export const buildStartWorkstationUrl = () => workstationsPath('start')()
export const buildStopWorkstationUrl = () => workstationsPath('stop')()

const searchPath = buildPath('search')
export const buildSearchUrl = (options: SearchOptions) => searchPath()({
  keywords: options.keywords,
  groups: options.groups,
  teamIDs: options.teamIDs,
  services: options.services,
  types: options.types,
  text: options.text,
  limit: options.limit,
  offset: options.offset,
})

const teamTokenPath = buildPath('user/token')
export const buildUpdateTeamTokenUrl = (team: string) => teamTokenPath()({team: team})