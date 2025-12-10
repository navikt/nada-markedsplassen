import { deleteTemplate, fetchTemplate, HttpError, postTemplate, putTemplate } from './request'
import {
  WorkstationOutput,
  WorkstationInput, WorkstationURLList, WorkstationOnpremAllowList, WorkstationURLListItem,
  ResyncAll, WorkstationURLListSettingsOpts
} from './generatedDto'
import { buildUrl } from './apiUrl'
import { useQuery } from '@tanstack/react-query'

const workstationsPath = buildUrl('workstations')
const buildGetWorkstationUrl = () => workstationsPath()()
const buildStartWorkstationUrl = () => workstationsPath('start')()
const buildStopWorkstationUrl = () => workstationsPath('stop')()
const buildRestartWorkstationUrl = () => workstationsPath('restart')()
const buildResyncWorkstationUrl = (id: string) => workstationsPath('resync', id)()
const buildGetWorkstationLogsURL = () => workstationsPath('logs')()
const buildGetWorkstationOptionsURL = () => workstationsPath('options')()
const buildGetWorkstationJobsURL = () => workstationsPath('job')()
const buildGetWorkstationResyncJobsURL = () => workstationsPath('resyncjob')()
const buildCreateWorkstationJobURL = () => workstationsPath('job')()
const buildGetWorkstationZonalTagBindingsURL = () => workstationsPath('bindings', 'tags')()
const buildUpdateWorkstationUrlAllowListURL = () => workstationsPath('urllist')()
const buildGetWorkstationUrlAllowList = () => workstationsPath('urllist', 'globalAllow')()
const buildListWorkstationsUrl = () => workstationsPath('list')()
const buildDleteWorkstationUrl = (id: string) => workstationsPath(id)()
const buildUpdateWorkstationOnpremMapping = () => workstationsPath('onpremhosts')()
const buildGetWorkstationOnpremMapping = () => workstationsPath('onpremhosts')()
const buildGetWorkstationURLList = () => workstationsPath('urllist')()
const buildGetWorkstationURLListForIdent = () => workstationsPath('urllist')()
const buildCreateWorkstationURLListItemForIdent = () => workstationsPath('urllist')()
const buildUpdateWorkstationURLListItemForIdent = () => workstationsPath('urllist')()
const buildDeleteWorkstationURLListItemForIdent = (id: string) => workstationsPath('urllist', id)()
const buildActivateWorkstationURLListForIdent = () => workstationsPath('urllist', 'activate')()
const buildDeactivateWorkstationURLListForIdent = () => workstationsPath('urllist', 'deactivate')()
const buildUpdateWorkstationsURLListUserSettings = () => workstationsPath('urllist', 'settings')()
const buildCreateWorkstationConnectivityURL = () => workstationsPath('workflow', 'connectivity')()
const buildCreateWorkstationsResyncAllURL = () => workstationsPath('workflow', 'resyncall')()
const buildConfigWorkstationSSHURL = (allow: string) => workstationsPath('ssh')({allow: allow})
const buildGetWorkstationSSHJobURL = () => workstationsPath('ssh', 'job')()

export const configWorkstationSSH = async (allow: boolean) =>
  postTemplate(buildConfigWorkstationSSHURL(allow.toString()))

export const getConfigWorkstationSSHJob = async () => {
  const url = buildGetWorkstationSSHJobURL()
  return fetchTemplate(url)
}

export const createWorkstationConnectivityWorkflow = async (input: WorkstationOnpremAllowList) =>
  postTemplate(buildCreateWorkstationConnectivityURL(), input)

export const createWorkstationResyncAllWorkflow = async (input: ResyncAll) =>
  postTemplate(buildCreateWorkstationsResyncAllURL(), input)

export const getWorkstationConnectivityWorkflow = async () => {
  const url = buildCreateWorkstationConnectivityURL()
  return fetchTemplate(url)
}

export const getWorkstationURLList = async () => {
  const url = buildGetWorkstationURLList()
  return fetchTemplate(url)
}

export const getWorkstationOnpremMapping = async () => {
  const url = buildGetWorkstationOnpremMapping()
  return fetchTemplate(url)
}

export const updateWorkstationOnpremMapping = async (mapping: WorkstationOnpremAllowList) => {
  const url = buildUpdateWorkstationOnpremMapping()
  return putTemplate(url, mapping)
}

export const getWorkstation = async () =>
  fetchTemplate(buildGetWorkstationUrl())

export const createWorkstationJob = async (input: WorkstationInput) =>
  postTemplate(buildCreateWorkstationJobURL(), input)

export const restartWorkstation = async () =>
  postTemplate(buildRestartWorkstationUrl())

export const resyncWorkstation = async (id: string) =>
  postTemplate(buildResyncWorkstationUrl(id))

export const startWorkstation = async () =>
  postTemplate(buildStartWorkstationUrl())

export const stopWorkstation = async () =>
  postTemplate(buildStopWorkstationUrl())

export const getWorkstationLogs = async () => {
  const url = buildGetWorkstationLogsURL()
  return fetchTemplate(url)
}

export const getWorkstationOptions = async () => {
  const url = buildGetWorkstationOptionsURL()
  return fetchTemplate(url)
}

export const getWorkstationJobs = async () => {
  const url = buildGetWorkstationJobsURL()
  return fetchTemplate(url)
}

export const getWorkstationResyncJobs = async () => {
  const url = buildGetWorkstationResyncJobsURL()
  return fetchTemplate(url)
}

export const getWorkstationZonalTagBindings = async () => {
  const url = buildGetWorkstationZonalTagBindingsURL()
  return fetchTemplate(url)
}

export const getWorkstationUrlAllowList = async () => {
    const url = buildGetWorkstationUrlAllowList()
    return fetchTemplate(url)
}
export const updateWorkstationUrlAllowList = async (urls: WorkstationURLList) => {
  const url = buildUpdateWorkstationUrlAllowListURL()
  return putTemplate(url, urls)
}

export const updateWorkstationsURLListUserSettings = async (workstationURLListSettings: WorkstationURLListSettingsOpts) => {
    const url = buildUpdateWorkstationsURLListUserSettings()
    return putTemplate(url, workstationURLListSettings)
}

const listWorkstations = async () =>
  fetchTemplate(buildListWorkstationsUrl())

export const useListWorkstationsPeriodically = () => useQuery<WorkstationOutput[], HttpError>({
  queryKey: ['workstations'],
  queryFn: listWorkstations,
  refetchInterval: 5000,
})

export const deleteWorkstation = async (slug: string) => deleteTemplate(buildDleteWorkstationUrl(slug))

export const createWorkstationURLListItemForIdent = async (input: WorkstationURLListItem) =>
  postTemplate(buildCreateWorkstationURLListItemForIdent(), input)

export const getWorkstationURLListForIdent = async () => {
  const url = buildGetWorkstationURLListForIdent()
  return fetchTemplate(url)
}

export const updateWorkstationURLListItemForIdent = async (input: WorkstationURLListItem) =>
  putTemplate(buildUpdateWorkstationURLListItemForIdent(), input)

export const deleteWorkstationURLListItemForIdent = async (id: string) =>
  deleteTemplate(buildDeleteWorkstationURLListItemForIdent(id))

export const activateWorkstationURLListForIdent = async (item_ids: string[]) =>
  putTemplate(buildActivateWorkstationURLListForIdent(), { item_ids })

export const deactivateWorkstationURLListForIdent = async (item_ids: string[]) =>
  putTemplate(buildDeactivateWorkstationURLListForIdent(), { item_ids })