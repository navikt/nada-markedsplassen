import { useEffect, useState } from "react";
import { fetchTemplate, HttpError, postTemplate } from "./request";
import { WorkstationLogs, WorkstationOptions, WorkstationOutput, WorkstationJobs } from "./generatedDto";
import { buildUrl } from "./apiUrl";
import { useQuery } from "react-query";

const workstationsPath = buildUrl('workstations')
const buildGetWorkstationUrl = () => workstationsPath()()
const buildEnsureWorkstationUrl = () => workstationsPath()()
const buildStartWorkstationUrl = () => workstationsPath('start')()
const buildStopWorkstationUrl = () => workstationsPath('stop')()
const buildGetWorkstationLogsURL = () => workstationsPath('logs')()
const buildGetWorkstationOptionsURL = () => workstationsPath('options')()
const buildGetWorkstationJobsURL = () => workstationsPath('job')()
const buildCreateWorkstationJobURL = () => workstationsPath('job')()


const getWorkstation = async () => 
    fetchTemplate(buildGetWorkstationUrl())

export const createWorkstationJob = async (body: {}) =>
    postTemplate(buildCreateWorkstationJobURL(), body)

export const ensureWorkstation = async (body: {}) => 
    postTemplate(buildEnsureWorkstationUrl(), body)

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

export const useGetWorkstation = ()=> useQuery<WorkstationOutput, HttpError>(['workstation'], getWorkstation)

export const useGetWorkstationOptions = ()=> 
    useQuery<WorkstationOptions, HttpError>(['workstationOptions'], getWorkstationOptions)

export const useGetWorkstationLogs = ()=>
    useQuery<WorkstationLogs, HttpError>(['workstationLogs'], ()=> getWorkstationLogs(), {refetchInterval: 5000})

export const useGetWorkstationJobs = ()=>
    useQuery<WorkstationJobs, HttpError>(['workstationJobs'], ()=> getWorkstationJobs(), {refetchInterval: 5000})

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
