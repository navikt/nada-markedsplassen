import { SearchOptions } from "./search"

const isServer = typeof window === 'undefined'

export const apiUrl = () => {
  if (process.env.NEXT_PUBLIC_ENV === 'development') {
    return 'http://localhost:8080/api'
  }
  return isServer ? 'http://nada-backend/api' : '/api'
}

export const getDataproductUrl = (id: string) => `${apiUrl()}/dataproducts/${id}`
export const createDataproductUrl = () => `${apiUrl()}/dataproducts/new`
export const updateDataproductUrl = (id: string) => `${apiUrl()}/dataproducts/${id}`
export const deleteDataproductUrl = (id: string) => `${apiUrl()}/dataproducts/${id}`

export const getDatasetUrl = (id: string) => `${apiUrl()}/datasets/${id}`
export const mapDatasetToServicesUrl = (datasetId: string) => `${apiUrl()}/datasets/${datasetId}/map`
export const createDatasetUrl = () => `${apiUrl()}/datasets/new`
export const deleteDatasetUrl = (id: string) => `${apiUrl()}/datasets/${id}`
export const updateDatasetUrl = (id: string) => `${apiUrl()}/datasets/${id}`
export const getAccessiblePseudoDatasetsUrl = () => `${apiUrl()}/datasets/pseudo/accessible`

export const createStoryUrl = () => `${apiUrl()}/stories/new`
export const updateStoryUrl = (id: string) => `${apiUrl()}/stories/${id}`
export const deleteStoryUrl = (id: string) => `${apiUrl()}/stories/${id}`

export const getJoinableViewUrl = (id: string) => `${apiUrl()}/pseudo/joinable/${id}`
export const createJoinableViewsUrl = () => `${apiUrl()}/pseudo/joinable/new`
export const getJoinableViewsForUserUrl = () => `${apiUrl()}/pseudo/joinable`

export const getProductAreasUrl = () => `${apiUrl()}/productareas`
export const getProductAreaUrl = (id: string) => `${apiUrl()}/productareas/${id}`
export const fetchUserDataUrl = () => `${apiUrl()}/userData`

export const isValidSlackChannelUrl = (channel: string) => `${apiUrl()}/slack/isValid?channel=${channel}`

export const getInsightProductUrl = (id: string) => `${apiUrl()}/insightProducts/${id}`
export const createInsightProductUrl = () => `${apiUrl()}/insightProducts/new`
export const updateInsightProductUrl = (id: string) => `${apiUrl()}/insightProducts/${id}`
export const deleteInsightProductUrl = (id: string) => `${apiUrl()}/insightProducts/${id}`

export const fetchKeywordsUrl = () => `${apiUrl()}/keywords`
export const updateKeywordsUrl = () => `${apiUrl()}/keywords`

export const fetchAccessRequestUrl = (datasetId: string) => `${apiUrl()}/accessRequests?datasetId=${datasetId}`
export const fetchBQDatasetsUrl = (projectId: string) => `${apiUrl()}/bigquery/datasets?projectId=${projectId}`
export const fetchBQTablesUrl = (projectId: string, datasetId: string) => `${apiUrl()}/bigquery/tables?projectId=${projectId}&datasetId=${datasetId}`
export const fetchBQColumnsUrl = (projectId: string, datasetId: string, tableId: string) => `${apiUrl()}/bigquery/columns?projectId=${projectId}&datasetId=${datasetId}&tableId=${tableId}`
export const fetchStoryMetadataURL = (id: string) => `${apiUrl()}/stories/${id}`
export const searchTeamKatalogenUrl = (gcpGroups?: string[]) => {
  const parameters = gcpGroups?.length ? gcpGroups.map(group => `gcpGroups=${encodeURIComponent(group)}`).join('&') : ''
  const query = parameters ? `?${parameters}` : ''
  return `${apiUrl()}/teamkatalogen${query}`
}
export const searchPollyUrl = (query?: string) => {
  return `${apiUrl()}/polly?query=${query}`
}

export const createAccessRequestUrl = () => `${apiUrl()}/accessRequests/new`
export const deleteAccessRequestUrl = (id: string) => `${apiUrl()}/accessRequests/${id}`
export const updateAccessRequestUrl = (id: string) => `${apiUrl()}/accessRequests/${id}`
export const approveAccessRequestUrl = (accessRequestId: string) => `${apiUrl()}/accessRequests/process/${accessRequestId}?action=approve`
export const denyAccessRequestUrl = (accessRequestId: string, reason: string) => `${apiUrl()}/accessRequests/process/${accessRequestId}?action=deny&reason=${reason}`
export const grantAccessUrl = () => `${apiUrl()}/accesses/grant`
export const revokeAccessUrl = (accessId: string) => `${apiUrl()}/accesses/revoke?id=${accessId}`

export const fetchTemplate = (url: string) => fetch(url, {
  method: 'GET',
  credentials: 'include',
  headers: {
    'Content-Type': 'application/json',
  },
}).then(res => {
  if (!res.ok) {
    throw new Error(res.statusText)
  }
  return res
})

export const postTemplate = (url: string, body?: any) => fetch(url, {
  method: 'POST',
  credentials: 'include',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify(body),
}).then(async res => {
  if (!res.ok) {
    const errorMessage = await res.text()
    throw new Error(`${res.statusText}${errorMessage&&":"}${errorMessage}`)
  }
  return res
})

export const putTemplate = (url: string, body?: any) => fetch(url, {
  method: 'PUT',
  credentials: 'include',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify(body),
}).then(res => {
  if (!res.ok) {
    throw new Error(res.statusText)
  }
  return res
})

export const deleteTemplate = (url: string, body?: any) => fetch(url, {
  method: 'DELETE',
  credentials: 'include',
  headers: {
    'Content-Type': 'application/json',
  },
}).then(res => {
  if (!res.ok) {
    throw new Error(res.statusText)
  }
  return res
})

export const searchUrl = (options: SearchOptions) => {
  let queryParams: string[] = [];

  // Helper function to add array-based options
  const addArrayOptions = (optionArray: string[] | undefined, paramName: string) => {
    if (optionArray && optionArray.length) {
      queryParams.push(`${paramName}=${optionArray.reduce((s, p) => `${s ? `${s},` : ""}${encodeURIComponent(p)}`)}`)
    }
  };

  // Adding array-based options
  addArrayOptions(options.keywords, 'keywords');
  addArrayOptions(options.groups, 'groups');
  addArrayOptions(options.teamIDs, 'teamIDs');
  addArrayOptions(options.services, 'services');
  addArrayOptions(options.types, 'types');

  // Adding single-value options
  if (options.text) queryParams.push(`text=${encodeURIComponent(options.text)}`);
  if (options.limit !== undefined) queryParams.push(`limit=${options.limit}`);
  if (options.offset !== undefined) queryParams.push(`offset=${options.offset}`);

  const query = queryParams.length ? `?${queryParams.join('&')}` : '';
  return `${apiUrl()}/search${query}`;
};

