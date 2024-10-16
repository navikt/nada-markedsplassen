import { useEffect, useState } from "react";
import { buildEnsureWorkstationUrl, buildGetWorkstationUrl, buildStartWorkstationUrl, buildStopWorkstationUrl } from "./apiUrl";
import { fetchTemplate, postTemplate } from "./request";
import { WorkstationOutput } from "./generatedDto";

export const getWorkstation = async () => 
    fetchTemplate(buildGetWorkstationUrl())

export const ensureWorkstation = async (body: {}) => 
    postTemplate(buildEnsureWorkstationUrl(), body)

export const startWorkstation = async () => 
    postTemplate(buildStartWorkstationUrl())

export const stopWorkstation = async () => 
    postTemplate(buildStopWorkstationUrl())

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
        const interval = setInterval(fetchWorkstation, 5000); // 5000ms = 5 seconds

        return () => clearInterval(interval); // Cleanup interval on component unmount
    }, [])

    return {workstation, loading}
}
