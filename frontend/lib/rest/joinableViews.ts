import { JoinableView, JoinableViewWithDatasource, NewJoinableViews } from "./generatedDto"
import { fetchTemplate, HttpError, postTemplate } from "./request"
import { buildUrl } from "./apiUrl"
import { useQuery } from '@tanstack/react-query'

const joinableViewPath = buildUrl('pseudo/joinable')
const buildGetJoinableViewUrl = (id: string) => joinableViewPath(id)()
const buildCreateJoinableViewsUrl = () => joinableViewPath('new')()
const buildGetJoinableViewsForUserUrl = () => joinableViewPath()()

const getJoinableView = async (id: string) => 
    fetchTemplate(buildGetJoinableViewUrl(id))

export const createJoinableViews = async (newJoinableView: NewJoinableViews) => 
    postTemplate(buildCreateJoinableViewsUrl(), newJoinableView)

const getJoinableViewsForUser = async () => 
    fetchTemplate(buildGetJoinableViewsForUserUrl())

export const useGetJoinableView = (id: string) => 
    useQuery<JoinableViewWithDatasource, HttpError>({
        queryKey: ['joinableView', id], 
        queryFn: ()=>getJoinableView(id)
    })

export const useGetJoinableViewsForUser = () => 
    useQuery<JoinableView[], HttpError>({
        queryKey: ['joinableViewsForUser'], 
        queryFn: getJoinableViewsForUser
    })
