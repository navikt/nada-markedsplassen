import { useEffect, useState } from "react"
import { IsValidSlackChannelResult } from "./generatedDto"
import { fetchTemplate, HttpError } from "./request"
import { buildUrl } from "./apiUrl"
import { useQuery } from '@tanstack/react-query'

const slackPath = buildUrl('slack')
const buildIsValidSlackChannelUrl = (channel: string) => slackPath('isValid')({channel: channel})

export const IsValidSlackChannel = (channel: string)=>
    fetchTemplate(buildIsValidSlackChannelUrl(channel))

export const useIsValidSlackChannel = (channel: string)=> useQuery<boolean, HttpError>({
    queryKey: ['slack', channel], 
    queryFn: ()=>
    IsValidSlackChannel(channel).then((r: IsValidSlackChannelResult)=> r.isValidSlackChannel)
})
