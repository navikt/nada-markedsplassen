import { useEffect, useState } from "react"
import { fetchTemplate, isValidSlackChannelUrl } from "./restApi"

export const IsValidSlackChannel = (channel: string)=>{
    var url = isValidSlackChannelUrl(channel)
    return fetchTemplate(url)
}

export const useIsValidSlackChannel = (channel: string)=>{
    const [isValid, setIsValid] = useState<boolean>(false)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)

    useEffect(()=>{
        if(!channel) return
        setLoading(true)
        IsValidSlackChannel(channel).then((res)=> res.json())
        .then((isValid)=>
        {
            setError(null)
            setIsValid(isValid)
        })
        .catch((err)=>{
            setError(err)
            setIsValid(false)            
        }).finally(()=>{
            setLoading(false)
        })
    }, [channel])

    return {isValid, loading, error}
}