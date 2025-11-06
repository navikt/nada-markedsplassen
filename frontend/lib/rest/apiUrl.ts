
const isServer = typeof window === 'undefined'

const paramToString = (param: string | string[] | undefined) => 
  Array.isArray(param) ? param.map(it=> encodeURIComponent(it)).join(',') : param!== undefined? encodeURIComponent(param): param

export type QueryParams = Record<string, string|string[]|number|undefined>
const queryParamsToString = (queryParams?: QueryParams) => {
  if (!queryParams) return ''
  
  const isEmptyArray = (value: any) => Array.isArray(value) ? !(value as []).length : false

  const paramString = Object.entries(queryParams).filter(it=> it[0] && it[1] && !isEmptyArray(it[1]))
    .map(([key, value]) => `${key}=${paramToString(value?.toString())}`)
    .join('&')

  return paramString ? `?${paramString}` : ''
}

const curriedBuildUrl = (baseUrl: string) => (path: string) => (...pathParams: string[]) => (queryParams?: Record<string, string|string[]|number|undefined>) =>
  `${baseUrl}/${path}${pathParams.length ? `/${pathParams.map(encodeURIComponent).join('/')}` : ''}${queryParamsToString(queryParams)}`

const buildUrlAsServer = curriedBuildUrl('http://nada-backend/api')
const buildUrlAsClient = curriedBuildUrl('/proxy')

export const buildUrl = isServer ? buildUrlAsServer : buildUrlAsClient
export const buildUrlAsClientWithoutProxy = curriedBuildUrl('/api')