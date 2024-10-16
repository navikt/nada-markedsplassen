import { SearchOptions } from "./generatedDto"

const isServer = typeof window === 'undefined'

const paramToString = (param: string | string[] | undefined) => 
  Array.isArray(param) ? param.map(it=> encodeURIComponent(it)).join(',') : param!== undefined? encodeURIComponent(param): param

const queryParamsToString = (queryParams?: Record<string, string|string[]|number|undefined>) => {
  if (!queryParams) return ''
  
  const isEmptyArray = (value: any) => Array.isArray(value) ? !(value as []).length : false

  const paramString = Object.entries(queryParams).filter(it=> it[0] && it[1] && !isEmptyArray(it[1]))
    .map(([key, value]) => `${key}=${paramToString(value?.toString())}`)
    .join('&')

  return paramString ? `?${paramString}` : ''
}

const buildUrl = (baseUrl: string) => (path: string) => (pathParam?: string) => (queryParams?: Record<string, string|string[]|number|undefined>) =>
  `${baseUrl}/${path}${pathParam ? `/${encodeURIComponent(pathParam)}` : ''}${queryParamsToString(queryParams)}`

const localBaseUrl = buildUrl('http://localhost:8080/api')
const asServerBaseUrl = buildUrl('http://nada-backend/api')
const asClientBaseUrl = buildUrl('/api')

export const buildPath = process.env.NEXT_PUBLIC_ENV === 'development' ? localBaseUrl :
  isServer ? asServerBaseUrl : asClientBaseUrl

