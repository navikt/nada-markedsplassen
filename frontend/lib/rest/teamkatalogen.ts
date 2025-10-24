import { TeamkatalogenResult } from "./generatedDto";
import { fetchTemplate } from "./request";
import { buildUrl } from "./apiUrl";
import { useQuery } from '@tanstack/react-query';

const teamKatalogenPath = buildUrl('teamkatalogen')
const buildSearchTeamKatalogenUrl = (gcpGroups?: string[]) => {
	const query = gcpGroups?.filter(it => it !== "")?.length ?  {gcpGroups: gcpGroups} : undefined

  return teamKatalogenPath()(query)
}

const searchTeamKatalogen = async (gcpGroups?: string[]) => 
    fetchTemplate(buildSearchTeamKatalogenUrl(gcpGroups))

export const useSearchTeamKatalogen = (gcpGroups?: string[]) => useQuery<TeamkatalogenResult[], any>({
  queryKey: ['teamkatalogen', gcpGroups], 
  queryFn: ()=>searchTeamKatalogen(gcpGroups)
})
