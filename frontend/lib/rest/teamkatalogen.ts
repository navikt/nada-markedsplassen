import { useEffect, useState } from "react";
import { fetchTemplate, searchTeamKatalogenUrl } from "./restApi";
import { TeamkatalogenResult } from "./generatedDto";

export const searchTeamKatalogen = async (gcpGroups?: string[]) => {
    const url = searchTeamKatalogenUrl(gcpGroups)
    return fetchTemplate(url)
}

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
