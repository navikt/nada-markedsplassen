import { useEffect, useState } from "react";
import { TeamkatalogenResult } from "./generatedDto";
import { fetchTemplate } from "./request";
import { buildSearchTeamKatalogenUrl } from "./apiUrl";

export const searchTeamKatalogen = async (gcpGroups?: string[]) => 
    fetchTemplate(buildSearchTeamKatalogenUrl(gcpGroups))

export const useSearchTeamKatalogen = (gcpGroups?: string[]) => {
    const [searchResult, setSearchResult] = useState<TeamkatalogenResult[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    useEffect(() => {
        setLoading(true);
        searchTeamKatalogen(gcpGroups).then((res) => res.json())
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
