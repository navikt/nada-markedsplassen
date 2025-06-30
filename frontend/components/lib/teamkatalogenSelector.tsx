import * as React from 'react'
import { UNSAFE_Combobox } from '@navikt/ds-react'
import LoaderSpinner from './spinner'
import { Dispatch, SetStateAction } from 'react'
import { useSearchTeamKatalogen } from '../../lib/rest/teamkatalogen'

type TeamkatalogenSelectorProps = {
  gcpGroups?: string []
  register: any
  watch: any
  errors: any
  setValue: any
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
  setValue,
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

  const buildOptions = () => {
    const options = []

    if (!error) {
      options.push({ label: "Velg team", value: "" })
    }

    if (error) {
      options.push({
        label: "Kan ikke hente teamene, men du kan registrere senere",
        value: "TeamkatalogenError"
      })
    }

    if (!error && (!relevantTeams || relevantTeams.length === 0)) {
      options.push({ label: "Ingen team", value: "NA" })
    }

    relevantTeams?.forEach((team: any) => {
      options.push({ label: team.name, value: team.url })
    })

    otherTeams?.forEach((team: any) => {
      options.push({ label: team.name, value: team.url })
    })

    return options
  }

  const handleSelectionChange = (value: string, isSelected: boolean) => {
    if (isSelected) {
      setValue('teamkatalogenURL', value)
      updateTeamkatalogInfo(value)
    }
  }

  // Find the selected team name to display in the combobox
  const getSelectedTeamName = () => {
    if (!teamkatalogenURL) return []

    const selectedTeam = allTeams?.find((team: any) => team.url === teamkatalogenURL)
    return selectedTeam ? [selectedTeam.name] : [teamkatalogenURL]
  }

  return (
    <UNSAFE_Combobox
      className="w-full"
      label="Team i Teamkatalogen"
      options={buildOptions()}
      selectedOptions={getSelectedTeamName()}
      onToggleSelected={handleSelectionChange}
      error={errors.teamkatalogenURL?.message}
    />
  )
}
export default TeamkatalogenSelector
