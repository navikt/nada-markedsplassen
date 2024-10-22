import { UserInfo } from "./generatedDto"
import { fetchTemplate, putTemplate } from "./request"
import { buildUrl } from "./apiUrl"
import { useQuery } from "react-query"

const userDataPath = buildUrl('userData')
const buildFetchUserDataUrl = () => userDataPath()()

const teamTokenPath = buildUrl('user/token')
const buildUpdateTeamTokenUrl = (team: string) => teamTokenPath()({ team: team })

const fetchUserData = async () =>
    fetchTemplate(buildFetchUserDataUrl())

export const updateTeamToken = async (team: string) =>
    putTemplate(buildUpdateTeamTokenUrl(team))

export const useFetchUserData = () => useQuery<UserInfo, any>(['userData'], fetchUserData)
