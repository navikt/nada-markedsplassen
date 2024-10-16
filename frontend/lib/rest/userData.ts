import { useEffect, useState } from "react"
import { UserInfo } from "./generatedDto"
import { fetchTemplate, putTemplate } from "./request"
import { buildFetchUserDataUrl, buildUpdateTeamTokenUrl } from "./apiUrl"

export const fetchUserData = async () => 
    fetchTemplate(buildFetchUserDataUrl())

export const updateTeamToken = async (team: string) => 
    putTemplate(buildUpdateTeamTokenUrl(team))

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
