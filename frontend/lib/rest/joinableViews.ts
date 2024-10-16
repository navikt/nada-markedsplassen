import { useEffect, useState } from "react"
import { JoinableView, JoinableViewWithDatasource, NewJoinableViews } from "./generatedDto"
import { fetchTemplate, postTemplate } from "./request"
import { buildPath } from "./apiUrl"

const joinableViewPath = buildPath('pseudo/joinable')
const buildGetJoinableViewUrl = (id: string) => joinableViewPath(id)()
const buildCreateJoinableViewsUrl = () => joinableViewPath('new')()
const buildGetJoinableViewsForUserUrl = () => joinableViewPath()()

const getJoinableView = async (id: string) => 
    fetchTemplate(buildGetJoinableViewUrl(id))

export const createJoinableViews = async (newJoinableView: NewJoinableViews) => 
    postTemplate(buildCreateJoinableViewsUrl(), newJoinableView)

const getJoinableViewsForUser = async () => 
    fetchTemplate(buildGetJoinableViewsForUserUrl())

export const useGetJoinableView = (id: string) => {
    const [joinableView, setJoinableView] = useState<JoinableViewWithDatasource|null>(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)

    useEffect(() => {
        if (!id) return
        getJoinableView(id)
            .then((joinableView) => {
                setError(null)
                setJoinableView(joinableView)
            })
            .catch((err) => {
                setError(err)
                setJoinableView(null)
            }).finally(() => {
                setLoading(false)
            })
    }, [id])

    return { data: joinableView, loading, error }
}

export const useGetJoinableViewsForUser = () => {
    const [joinableViews, setJoinableViews] = useState<JoinableView[]>([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)

    useEffect(() => {
        getJoinableViewsForUser()
            .then((joinableViews) => {
                setError(null)
                setJoinableViews(joinableViews)
            })
            .catch((err) => {
                setError(err)
                setJoinableViews([])
            }).finally(() => {
                setLoading(false)
            })
    }, [])

    return { data: joinableViews, loading, error }
}