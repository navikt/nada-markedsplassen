import { useEffect, useState } from "react";
import { fetchTemplate, HttpError, postTemplate } from "./request";
import {
    WorkstationLogs,
    WorkstationOptions,
    WorkstationOutput,
    WorkstationJobs,
    WorkstationZonalTagBindingJobs, EffectiveTags
} from "./generatedDto";
import { buildUrl } from "./apiUrl";
import { useQuery } from "react-query";

const workstationsPath = buildUrl('workstations')
const buildGetWorkstationUrl = () => workstationsPath()()
const buildStartWorkstationUrl = () => workstationsPath('start')()
const buildGetWorkstationStartJobsUrl = () => workstationsPath('start')()
const buildGetWorkstationJobsUrl = () => workstationsPath('job')()
const buildStopWorkstationUrl = () => workstationsPath('stop')()
const buildGetWorkstationLogsURL = () => workstationsPath('logs')()
const buildGetWorkstationOptionsURL = () => workstationsPath('options')()
const buildGetWorkstationJobsURL = () => workstationsPath('job')()
const buildCreateWorkstationJobURL = () => workstationsPath('job')()
const buildCreateWorkstationZonalTagBindingJobsURL = () => workstationsPath('bindings')()
const buildGetWorkstationZonalTagBindingJobsURL = () => workstationsPath('bindings')()
const buildGetWorkstationZonalTagBindingsURL = () => workstationsPath('bindingstags')()


export const getWorkstation = async () =>
    fetchTemplate(buildGetWorkstationUrl())

export const createWorkstationJob = async (body: {}) =>
    postTemplate(buildCreateWorkstationJobURL(), body)

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

const getWorkstationOptions = async () => {
    const url = buildGetWorkstationOptionsURL();
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

export const getWorkstationZonalTagBindings = async () => {
    const url = buildGetWorkstationZonalTagBindingsURL();
    return fetchTemplate(url);
}

export const getWorkstationZonalTagBindingJobs = async () => {
    const url = buildGetWorkstationZonalTagBindingJobsURL();
    return fetchTemplate(url);
}

export const useGetWorkstation = ()=>
    useQuery<WorkstationOutput, HttpError>(['workstation'], getWorkstation)

export const useGetWorkstationOptions = ()=> 
    useQuery<WorkstationOptions, HttpError>(['workstationOptions'], getWorkstationOptions)

export const useGetWorkstationJobs = ()=>
    useQuery<WorkstationJobs, HttpError>(['workstationJobs'], ()=> getWorkstationJobs(), {refetchInterval: 5000})

export const useGetWorkstationStartJobs = ()=>
    useQuery<WorkstationJobs, HttpError>(['workstationStartJobs'], ()=> getWorkstationStartJobs(), {refetchInterval: 5000})

export const useConditionalEffectiveTags = (isRunning: boolean) => {
    return useQuery<EffectiveTags, HttpError>(
        ['effectiveTags'],
        getWorkstationZonalTagBindings,
        {
            enabled: isRunning,
        }
    );
};

export const useConditionalWorkstationLogs = (isRunning: boolean) => {
    return useQuery<WorkstationLogs, HttpError>(
        ['workstationLogs'],
        getWorkstationLogs,
        {
            enabled: isRunning,
        }
    );
};

export const useConditionalWorkstationZonalTagBindingJobs = (isRunning: boolean) => {
    return useQuery<WorkstationZonalTagBindingJobs, HttpError>(
        ['workstationZonalTagBindingJobs'],
        getWorkstationZonalTagBindingJobs,
        {
            enabled: isRunning,
        }
    );
}
