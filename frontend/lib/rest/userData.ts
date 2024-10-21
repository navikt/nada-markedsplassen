import { UserInfo } from "./generatedDto"
import { fetchTemplate, putTemplate } from "./request"
import { buildPath } from "./apiUrl"
import { useQuery } from "react-query"

const userDataPath = buildPath('userData')
const buildFetchUserDataUrl = () => userDataPath()()

const teamTokenPath = buildPath('user/token')
const buildUpdateTeamTokenUrl = (team: string) => teamTokenPath()({ team: team })

const fetchUserData = async () =>
    fetchTemplate(buildFetchUserDataUrl())

export const updateTeamToken = async (team: string) =>
    putTemplate(buildUpdateTeamTokenUrl(team))

export const useFetchUserData = () => useQuery<UserInfo, any>(['userData'], fetchUserData)
