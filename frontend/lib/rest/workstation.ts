import { fetchTemplate, HttpError, postTemplate } from "./request";
import {
    WorkstationLogs,
    WorkstationOptions,
    WorkstationOutput,
    WorkstationJobs,
    WorkstationZonalTagBindingJobs, EffectiveTags, WorkstationInput
} from "./generatedDto";
import { buildUrl } from "./apiUrl";
import { useQuery } from "react-query";

const workstationsPath = buildUrl('workstations')
const buildGetWorkstationUrl = () => workstationsPath()()
const buildStartWorkstationUrl = () => workstationsPath('start')()
const buildGetWorkstationStartJobsUrl = () => workstationsPath('start')()
const buildGetWorkstationStartJobUrl = (id: string) => workstationsPath('start', id)()
const buildStopWorkstationUrl = () => workstationsPath('stop')()
const buildGetWorkstationLogsURL = () => workstationsPath('logs')()
const buildGetWorkstationOptionsURL = () => workstationsPath('options')()
const buildGetWorkstationJobsURL = () => workstationsPath('job')()
const buildGetWorkstationJobURL = (id: string) => workstationsPath('job', id)()
const buildCreateWorkstationJobURL = () => workstationsPath('job')()
const buildCreateWorkstationZonalTagBindingJobsURL = () => workstationsPath('bindings')()
const buildGetWorkstationZonalTagBindingJobsURL = () => workstationsPath('bindings')()
const buildGetWorkstationZonalTagBindingsURL = () => workstationsPath('bindings', 'tags')()

export const getWorkstation = async () =>
    fetchTemplate(buildGetWorkstationUrl())

export const createWorkstationJob = async (input: WorkstationInput) =>
    postTemplate(buildCreateWorkstationJobURL(), input)

export const startWorkstation = async () =>
    postTemplate(buildStartWorkstationUrl())

export const stopWorkstation = async () => 
    postTemplate(buildStopWorkstationUrl())

export const createWorkstationZonalTagBindingJob = async () =>
    postTemplate(buildCreateWorkstationZonalTagBindingJobsURL())

export const getWorkstationLogs = async () => {
    const url = buildGetWorkstationLogsURL();
    return fetchTemplate(url);
}

export const getWorkstationOptions = async () => {
    const url = buildGetWorkstationOptionsURL();
    return fetchTemplate(url);
}

export const getWorkstationJob = async (id: string) => {
    const url = buildGetWorkstationJobURL(id);
    return fetchTemplate(url);
}

export const getWorkstationJobs = async () => {
    const url = buildGetWorkstationJobsURL();
    return fetchTemplate(url);
}

export const getWorkstationStartJobs = async () => {
    const url = buildGetWorkstationStartJobsUrl()
    return fetchTemplate(url);
}

export const getWorkstationStartJob = async (id: string) => {
    const url = buildGetWorkstationStartJobUrl(id)
    return fetchTemplate(url);
}

export const getWorkstationZonalTagBindings = async () => {
    const url = buildGetWorkstationZonalTagBindingsURL();
    return fetchTemplate(url);
}

export const getWorkstationZonalTagBindingJobs = async () => {
    const url = buildGetWorkstationZonalTagBindingJobsURL();
    return fetchTemplate(url);
}
