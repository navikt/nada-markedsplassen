import { useEffect, useState } from "react";
import { fetchTemplate, postTemplate } from "./request";
import { WorkstationLogs, WorkstationOptions, WorkstationOutput } from "./generatedDto";
import { buildPath } from "./apiUrl";

const workstationsPath = buildPath('workstations')
const buildGetWorkstationUrl = () => workstationsPath()()
const buildEnsureWorkstationUrl = () => workstationsPath()()
const buildStartWorkstationUrl = () => workstationsPath('start')()
const buildStopWorkstationUrl = () => workstationsPath('stop')()
const buildGetWorkstationLogsURL = () => workstationsPath('logs')()
const buildGetWorkstationOptionsURL = () => workstationsPath('options')()


export const getWorkstation = async () => 
    fetchTemplate(buildGetWorkstationUrl())

export const ensureWorkstation = async (body: {}) => 
    postTemplate(buildEnsureWorkstationUrl(), body)

export const startWorkstation = async () => 
    postTemplate(buildStartWorkstationUrl())

export const stopWorkstation = async () => 
    postTemplate(buildStopWorkstationUrl())

export const getWorkstationLogs = async () => {
    const url = buildGetWorkstationLogsURL();
    return fetchTemplate(url);
}

export const getWorkstationOptions = async () => {
    const url = buildGetWorkstationOptionsURL();
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
