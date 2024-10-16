import { useEffect, useState } from "react";
import { TeamkatalogenResult } from "./generatedDto";
import { fetchTemplate } from "./request";
import { buildPath } from "./apiUrl";

const teamKatalogenPath = buildPath('teamkatalogen')
const buildSearchTeamKatalogenUrl = (gcpGroups?: string[]) => 
  teamKatalogenPath()({gcpGroups: gcpGroups?.length ? gcpGroups.map(group => `gcpGroups=${encodeURIComponent(group)}`).join('&') : ''})

export const searchTeamKatalogen = async (gcpGroups?: string[]) => 
    fetchTemplate(buildSearchTeamKatalogenUrl(gcpGroups))

export const useSearchTeamKatalogen = (gcpGroups?: string[]) => {
    const [searchResult, setSearchResult] = useState<TeamkatalogenResult[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    useEffect(() => {
        setLoading(true);
        searchTeamKatalogen(gcpGroups)
            .then((searchResultDto) => {
            setError(null);
            setSearchResult(searchResultDto);
        })
            .catch((err) => {
            setError(err);
            setSearchResult([]);
        }).finally(() => {
            setLoading(false);
        })
    }, gcpGroups? [JSON.stringify(gcpGroups)]: [])
    return { searchResult, loading, error };
}
