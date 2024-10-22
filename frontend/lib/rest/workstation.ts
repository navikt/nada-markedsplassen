import { useEffect, useState } from "react";
import { fetchTemplate, HttpError, postTemplate } from "./request";
import { WorkstationLogs, WorkstationOptions, WorkstationOutput } from "./generatedDto";
import { buildUrl } from "./apiUrl";
import { useQuery } from "react-query";

const workstationsPath = buildUrl('workstations')
const buildGetWorkstationUrl = () => workstationsPath()()
const buildEnsureWorkstationUrl = () => workstationsPath()()
const buildStartWorkstationUrl = () => workstationsPath('start')()
const buildStopWorkstationUrl = () => workstationsPath('stop')()
const buildGetWorkstationLogsURL = () => workstationsPath('logs')()
const buildGetWorkstationOptionsURL = () => workstationsPath('options')()


const getWorkstation = async () => 
    fetchTemplate(buildGetWorkstationUrl())

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

export const useGetWorkstation = ()=> useQuery<WorkstationOutput, HttpError>(['workstation'], getWorkstation)

export const useGetWorkstationOptions = ()=> 
    useQuery<WorkstationOptions, HttpError>(['workstationOptions'], getWorkstationOptions)

export const useGetWorkstationLogs = ()=>
    useQuery<WorkstationLogs, HttpError>(['workstationLogs'], ()=> getWorkstationLogs(), {refetchInterval: 5000})
