import {NadaTokens, UserInfo} from "./generatedDto"
import { fetchTemplate, putTemplate } from "./request"
import { buildUrl } from "./apiUrl"
import { useQuery } from '@tanstack/react-query'

const userDataPath = buildUrl('userData')
const buildFetchUserDataUrl = () => userDataPath()()

const teamTokenPath = buildUrl('user/token')
const buildUpdateTeamTokenUrl = (team: string) => teamTokenPath()({ team: team })

const tokensPath = buildUrl('tokens')
const buildFetchTokensUrl = () => tokensPath()()

const fetchUserData = async () =>
    fetchTemplate(buildFetchUserDataUrl())

export const updateTeamToken = async (team: string) =>
    putTemplate(buildUpdateTeamTokenUrl(team))

export const useFetchUserData = () => useQuery<UserInfo, any>({
    queryKey: ['userData'], 
    queryFn: fetchUserData
})

const fetchTokens = async () =>
    fetchTemplate(buildFetchTokensUrl())

export const useFetchTokens = () => useQuery<NadaTokens, any>({
    queryKey: ['tokens'],
    queryFn: fetchTokens
})
