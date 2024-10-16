import { useEffect, useState } from "react"
import { IsValidSlackChannelResult } from "./generatedDto"
import { fetchTemplate } from "./request"
import { buildPath } from "./apiUrl"

const slackPath = buildPath('slack')
const buildIsValidSlackChannelUrl = (channel: string) => slackPath('isValid')({channel: channel})

export const IsValidSlackChannel = (channel: string)=>
    fetchTemplate(buildIsValidSlackChannelUrl(channel))

export const useIsValidSlackChannel = (channel: string)=>{
    const [isValid, setIsValid] = useState<boolean>(false)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)

    useEffect(()=>{
        if(!channel) return
        setLoading(true)
        IsValidSlackChannel(channel)
        .then((r: IsValidSlackChannelResult)=>
        {
            setError(null)
            setIsValid(r.isValidSlackChannel)
        })
        .catch((err: any)=>{
            setError(err)
            setIsValid(false)            
        }).finally(()=>{
            setLoading(false)
        })
    }, [channel])

    return {isValid, loading, error}
}