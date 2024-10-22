import * as React from 'react'
import { Select } from '@navikt/ds-react'
import LoaderSpinner from './spinner'
import { Dispatch, SetStateAction } from 'react'
import { useSearchTeamKatalogen } from '../../lib/rest/teamkatalogen'

type TeamkatalogenSelectorProps = {
  gcpGroups?: string []
  register: any
  watch: any
  errors: any
  setProductAreaID?: Dispatch<SetStateAction<string>>
  setTeamID?: Dispatch<SetStateAction<string>>
}

export interface Team {
  name: string
  url: string
  productAreaID: string
  teamID: string
}

const useBuildTeamList = (gcpGroups: string [] | undefined) => {
  const {data: relevantTeamResult, isLoading: loadingRelevant, error: errorRelevant}= 
  useSearchTeamKatalogen(gcpGroups?.map(it=> it.split('@')[0]) || [])
  const {data: allTeamsResult, isLoading: loadingAllTeams, error: errorAllTeams} =
  useSearchTeamKatalogen()

  if (errorAllTeams || errorRelevant) {
    return {
      error: true,
    }
  }

  const relevantTeams = gcpGroups? relevantTeamResult: undefined
  const otherTeams = allTeamsResult?.filter(
    (it: any) => !relevantTeams || !relevantTeams.find((t: any) => t.teamID == it.teamID)
  )

  otherTeams?.sort((a:any, b:any) => a.name.localeCompare(b.name))

  return {
    relevantTeams: relevantTeams,
    otherTeams: otherTeams,
    allTeams: allTeamsResult,
  }
}

export const TeamkatalogenSelector = ({
  gcpGroups,
  register,
  watch,
  errors,
  setProductAreaID,
  setTeamID,
}: TeamkatalogenSelectorProps) => {

  const { relevantTeams, otherTeams, allTeams, error } =
    useBuildTeamList(gcpGroups)
  const teamkatalogenURL = watch('teamkatalogenURL')

  const updateTeamkatalogInfo = (url: string) => {
    const team = allTeams?.find((it:any) => it.url == url)
    setProductAreaID?.(team ? team.productAreaID : '')
    setTeamID?.(team ? team.teamID : '')
  }

  updateTeamkatalogInfo(teamkatalogenURL)

  if (!allTeams) return <LoaderSpinner />

  return (
    <Select
      className="w-full"
      label="Team i Teamkatalogen"
      {...register('teamkatalogenURL')}
      error={errors.teamkatalogenURL?.message}
      value={teamkatalogenURL}
    >
      {!error && <option value="">Velg team</option>}
      {error && (
        <option value="TeamkatalogenError">
          Kan ikke hente teamene, men du kan registrere senere
        </option>
      )}
      {!error && (!relevantTeams || relevantTeams.length == 0) && (
        <option value="NA" key="Ingen team">
          Ingen team
        </option>
      )}
      {relevantTeams?.map((team: any) => (
        <option value={team.url} key={team.name}>
          {team.name}
        </option>
      ))}
      {otherTeams?.map((team: any) => (
        <option value={team.url} key={team.name}>
          {team.name}
        </option>
      ))}
    </Select>
  )
}
export default TeamkatalogenSelector
