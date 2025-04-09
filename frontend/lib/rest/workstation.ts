import { deleteTemplate, fetchTemplate, HttpError, postTemplate, putTemplate } from './request'
import {
  WorkstationOutput,
  WorkstationInput, WorkstationURLList, WorkstationOnpremAllowList,
} from './generatedDto'
import { buildUrl } from './apiUrl'
import { useQuery } from '@tanstack/react-query'

const workstationsPath = buildUrl('workstations')
const buildGetWorkstationUrl = () => workstationsPath()()
const buildStartWorkstationUrl = () => workstationsPath('start')()
const buildStopWorkstationUrl = () => workstationsPath('stop')()
const buildGetWorkstationLogsURL = () => workstationsPath('logs')()
const buildGetWorkstationOptionsURL = () => workstationsPath('options')()
const buildGetWorkstationJobsURL = () => workstationsPath('job')()
const buildCreateWorkstationJobURL = () => workstationsPath('job')()
const buildGetWorkstationZonalTagBindingsURL = () => workstationsPath('bindings', 'tags')()
const buildUpdateWorkstationUrlAllowListURL = () => workstationsPath('urllist')()
const buildListWorkstationsUrl = () => workstationsPath('list')()
const buildDleteWorkstationUrl = (id: string) => workstationsPath(id)()
const buildUpdateWorkstationOnpremMapping = () => workstationsPath('onpremhosts')()
const buildGetWorkstationOnpremMapping = () => workstationsPath('onpremhosts')()
const buildGetWorkstationURLList = () => workstationsPath('urllist')()
const buildCreateWorkstationConnectivityURL = () => workstationsPath('workflow', 'connectivity')()

export const createWorkstationConnectivityWorkflow = async (input: WorkstationOnpremAllowList) =>
  postTemplate(buildCreateWorkstationConnectivityURL(), input)

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

export const getWorkstationZonalTagBindings = async () => {
  const url = buildGetWorkstationZonalTagBindingsURL()
  return fetchTemplate(url)
}

export const updateWorkstationUrlAllowList = async (urls: WorkstationURLList) => {
  const url = buildUpdateWorkstationUrlAllowListURL()
  return putTemplate(url, urls)
}

const listWorkstations = async () =>
  fetchTemplate(buildListWorkstationsUrl())

export const useListWorkstationsPeriodically = () => useQuery<WorkstationOutput[], HttpError>({
  queryKey: ['workstations'],
  queryFn: listWorkstations,
  refetchInterval: 5000,
})

export const deleteWorkstation = async (slug: string) => deleteTemplate(buildDleteWorkstationUrl(slug))
