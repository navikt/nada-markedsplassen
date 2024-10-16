import { useEffect, useState } from "react";
import { fetchTemplate, postTemplate } from "./request";
import { WorkstationOutput } from "./generatedDto";
import { buildPath } from "./apiUrl";

const workstationsPath = buildPath('workstations')
const buildGetWorkstationUrl = () => workstationsPath()()
const buildEnsureWorkstationUrl = () => workstationsPath()()
const buildStartWorkstationUrl = () => workstationsPath('start')()
const buildStopWorkstationUrl = () => workstationsPath('stop')()

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
            getWorkstation()
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
