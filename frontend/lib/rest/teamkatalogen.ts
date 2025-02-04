import { useEffect, useState } from "react";
import { TeamkatalogenResult } from "./generatedDto";
import { fetchTemplate } from "./request";
import { buildUrl } from "./apiUrl";
import { useQuery } from '@tanstack/react-query';

const teamKatalogenPath = buildUrl('teamkatalogen')
const buildSearchTeamKatalogenUrl = (gcpGroups?: string[]) => 
  teamKatalogenPath()({gcpGroups: gcpGroups?.length ? gcpGroups.map(group => `gcpGroups=${encodeURIComponent(group)}`).join('&') : ''})

const searchTeamKatalogen = async (gcpGroups?: string[]) => 
    fetchTemplate(buildSearchTeamKatalogenUrl(gcpGroups))

export const useSearchTeamKatalogen = (gcpGroups?: string[]) => useQuery<TeamkatalogenResult[], any>({
  queryKey: ['teamkatalogen', gcpGroups], 
  queryFn: ()=>searchTeamKatalogen(gcpGroups)
})
