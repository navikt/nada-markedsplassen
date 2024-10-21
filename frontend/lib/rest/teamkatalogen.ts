import { useEffect, useState } from "react";
import { TeamkatalogenResult } from "./generatedDto";
import { fetchTemplate } from "./request";
import { buildPath } from "./apiUrl";
import { useQuery } from "react-query";

const teamKatalogenPath = buildPath('teamkatalogen')
const buildSearchTeamKatalogenUrl = (gcpGroups?: string[]) => 
  teamKatalogenPath()({gcpGroups: gcpGroups?.length ? gcpGroups.map(group => `gcpGroups=${encodeURIComponent(group)}`).join('&') : ''})

export const searchTeamKatalogen = async (gcpGroups?: string[]) => 
    fetchTemplate(buildSearchTeamKatalogenUrl(gcpGroups))

export const useSearchTeamKatalogen = (gcpGroups?: string[]) => useQuery<TeamkatalogenResult[], any>(['teamkatalogen', gcpGroups], ()=>searchTeamKatalogen(gcpGroups))
