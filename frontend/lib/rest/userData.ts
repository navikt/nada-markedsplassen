import { useEffect, useState } from "react"
import {
    fetchTemplate,
    fetchUserDataUrl,
    ensureWorkstationURL,
    postTemplate,
    getWorkstationURL,
    startWorkstationURL,
    stopWorkstationURL,
    getWorkstationLogsURL,
    getWorkstationOptionsURL,
} from "./restApi"
import { UserInfo, Workstation, WorkstationOutput, WorkstationOptions, WorkstationLogs } from "./generatedDto"

export const fetchUserData = async () => {
    const url = fetchUserDataUrl()
    return fetchTemplate(url)
}

export const useFetchUserData = () => {
    const [data, setData] = useState<UserInfo|null>(null);
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

export const getWorkstationLogs = async () => {
    const url = getWorkstationLogsURL();
    return fetchTemplate(url);
}

export const getWorkstationOptions = async () => {
    const url = getWorkstationOptionsURL();
    return fetchTemplate(url);
}

export const useGetWorkstation = ()=>{
    const [workstation, setWorkstation] = useState<WorkstationOutput|null>(null)
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
    }, [])

    return {workstation, loading}
}

export const useGetWorkstationOptions = ()=>{
    const [workstationOptions, setWorkstationOptions] = useState<WorkstationOptions|null>(null)
    const [loadingOptions, setLoadingOptions] = useState(true)

    useEffect(()=>{
        const fetchWorkstationOptions = () => {
            getWorkstationOptions().then((res)=> res.json())
                .then((workstationOptions)=>
                {
                    setWorkstationOptions(workstationOptions)
                })
                .catch((err)=>{
                    setWorkstationOptions(null)
                    setLoadingOptions(false)
                }).finally(()=>{
                setLoadingOptions(false)
            })
        }

        fetchWorkstationOptions();
    }, [])

    return {workstationOptions, loadingOptions}
}


export const useGetWorkstationLogs = ()=>{
    const [workstationLogs, setWorkstationLogs] = useState<WorkstationLogs|null>(null)
    const [loadingLogs, setLoadingLogs] = useState(true)

    useEffect(()=>{
        const fetchWorkstationLogs = () => {
            getWorkstationLogs().then((res)=> res.json())
                .then((workstationLogs)=>
                {
                    setWorkstationLogs(workstationLogs)
                })
                .catch((err)=>{
                    setWorkstationLogs(null)
                    setLoadingLogs(false)
                }).finally(()=>{
                setLoadingLogs(false)
            })
        }

        fetchWorkstationLogs();

        const interval = setInterval(fetchWorkstationLogs, 5000);
        return () => clearInterval(interval);
    }, [])

    return {workstationLogs, loadingLogs}
}