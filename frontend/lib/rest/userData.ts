import { useEffect, useState } from "react"
import { fetchTemplate, fetchUserDataUrl } from "./restApi"

export const fetchUserData = async () => {
    const url = fetchUserDataUrl()
    return fetchTemplate(url)
}

export const useFetchUserData = () => {
    const [data, setData] = useState<any>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    useEffect(() => {
        setLoading(true);
        fetchUserData().then((res) => res.json())
            .then((userDataDto) => {
            setError(null);
            setLoading(false);
            setData(userDataDto);
        })
            .catch((err) => {
            setError(err);
            setLoading(false);
            setData(null);
        }).finally(() => {
            setLoading(false);
        })
    }, [])
    return { data, loading, error };
}