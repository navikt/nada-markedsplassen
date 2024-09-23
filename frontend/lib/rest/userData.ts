import { useEffect, useState } from "react"
import { fetchTemplate, fetchUserDataUrl, ensureWorkstationURL, postTemplate, getWorkstationURL, startWorkstationURL, stopWorkstationURL } from "./restApi"

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

export const getWorkstation = async () => {
    const url = getWorkstationURL();
    return fetchTemplate(url);
}

export const ensureWorkstation = async (body: {}) => {
    const url = ensureWorkstationURL();
    return postTemplate(url, body);
}

export const startWorkstation = async () => {
    const url = startWorkstationURL();
    return postTemplate(url);
}

export const stopWorkstation = async () => {
    const url = stopWorkstationURL();
    return postTemplate(url);
}

export const useGetWorkstation = ()=>{
    const [workstation, setWorkstation] = useState<any>(null)
    const [loading, setLoading] = useState(true)

    useEffect(()=>{
        const fetchWorkstation = () => {
            getWorkstation().then((res)=> res.json())
                .then((workstation)=>
                {
                    setWorkstation(workstation)
                })
                .catch((err)=>{
                    setWorkstation(null)
                    setLoading(false)
                }).finally(()=>{
                setLoading(false)
            })
        }

        fetchWorkstation();
        const interval = setInterval(fetchWorkstation, 5000); // 5000ms = 5 seconds

        return () => clearInterval(interval); // Cleanup interval on component unmount
    }, [])

    return {workstation, loading}
}
