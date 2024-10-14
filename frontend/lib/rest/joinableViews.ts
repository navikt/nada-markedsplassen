import { useEffect, useState } from "react"
import { createJoinableViewsUrl, fetchTemplate, getJoinableViewUrl, getJoinableViewsForUserUrl, postTemplate } from "./restApi"
import { JoinableView, JoinableViewWithDatasource, NewJoinableViews } from "./generatedDto"

const getJoinableView = async (id: string) => {
    const url = getJoinableViewUrl(id)
    return fetchTemplate(url)
}

export const createJoinableViews = async (newJoinableView: NewJoinableViews) => {
    const url = createJoinableViewsUrl()
    return postTemplate(url, newJoinableView)
}

export const getJoinableViewsForUser = async () => {
    const url = getJoinableViewsForUserUrl()
    return fetchTemplate(url)
}

export const useGetJoinableView = (id: string) => {
    const [joinableView, setJoinableView] = useState<JoinableViewWithDatasource|null>(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)

    useEffect(() => {
        if (!id) return
        getJoinableView(id).then((res) => res.json())
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
        getJoinableViewsForUser().then((res) => res.json())
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