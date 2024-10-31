import { useEffect, useState } from "react";
import { fetchTemplate, HttpError, postTemplate } from "./request";
import { WorkstationLogs, WorkstationOptions, WorkstationOutput, WorkstationJobs } from "./generatedDto";
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


const getWorkstation = async () => 
    fetchTemplate(buildGetWorkstationUrl())

export const createWorkstationJob = async (body: {}) =>
    postTemplate(buildCreateWorkstationJobURL(), body)

export const startWorkstation = async () =>
    postTemplate(buildStartWorkstationUrl())

export const stopWorkstation = async () => 
    postTemplate(buildStopWorkstationUrl())

const getWorkstationLogs = async () => {
    const url = buildGetWorkstationLogsURL();
    return fetchTemplate(url);
}

const getWorkstationOptions = async () => {
    const url = buildGetWorkstationOptionsURL();
    return fetchTemplate(url);
}

const getWorkstationJobs = async () => {
    const url = buildGetWorkstationJobsURL();
    return fetchTemplate(url);
}

const getWorkstationStartJobs = async () => {
    const url = buildGetWorkstationStartJobsUrl()
    return fetchTemplate(url);
}

export const useGetWorkstation = ()=>
    useQuery<WorkstationOutput, HttpError>(['workstation'], getWorkstation, {refetchInterval: 5000})

export const useGetWorkstationOptions = ()=> 
    useQuery<WorkstationOptions, HttpError>(['workstationOptions'], getWorkstationOptions)

export const useGetWorkstationJobs = ()=>
    useQuery<WorkstationJobs, HttpError>(['workstationJobs'], ()=> getWorkstationJobs(), {refetchInterval: 5000})

export const useGetWorkstationStartJobs = ()=>
    useQuery<WorkstationJobs, HttpError>(['workstationStartJobs'], ()=> getWorkstationStartJobs(), {refetchInterval: 5000})

export const useConditionalWorkstationLogs = (isRunning: boolean) => {
    return useQuery<WorkstationLogs, HttpError>(
        ['workstationLogs'],
        getWorkstationLogs,
        {
            enabled: isRunning,
            refetchInterval: isRunning ? 5000 : false,
        }
    );
};
