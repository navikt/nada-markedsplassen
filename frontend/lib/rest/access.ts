import { useEffect, useState } from "react";
import { approveAccessRequestUrl, createAccessRequestUrl, deleteAccessRequestUrl, deleteTemplate, denyAccessRequestUrl, fetchAccessRequestUrl, fetchTemplate, grantAccessUrl, postTemplate, putTemplate, revokeAccessUrl } from "./restApi";

export const fetchAccessRequests = async (datasetId: string) => {
    const url = fetchAccessRequestUrl(datasetId);
    return fetchTemplate(url);
}

export const useFetchAccessRequestsForDataset = (datasetId: string)=>{
    const [data, setData] = useState<any>(undefined)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)


    useEffect(()=>{
        if(!datasetId) return
        fetchAccessRequests(datasetId).then((res)=> res.json())
        .then((data)=>
        {
            setError(null)
            setData(data)
        })
        .catch((err)=>{
            setError(err)
            setData(undefined)            
        }).finally(()=>{
            setLoading(false)
        })
    }, [datasetId])

    return {data, loading, error}
}

export enum SubjectType {
    Group = 'group',
    ServiceAccount = 'serviceAccount',
    User = 'user'
  }

export type PollyInput = {
    externalID: string;
    id?: string;
    name: string;
    url: string;
};

export const createAccessRequest = async ( 
datasetID: string,
expires: Date| undefined,
owner: string|undefined,
polly: PollyInput|undefined,
subject: string|undefined,
subjectType: string|undefined)=>{
    const url = createAccessRequestUrl();
    return postTemplate(url, {datasetID, expires, owner, polly, subject, subjectType});
}

export const deleteAccessRequest = async (id: string) => {
    const url = deleteAccessRequestUrl(id);
    return deleteTemplate(url);
}

export const updateAccessRequest = async (id: string, expires: Date|undefined, owner: string, polly: PollyInput|undefined, subject: string|undefined, subjectType: string|undefined) => {
    const url = deleteAccessRequestUrl(id);
    return putTemplate(url, {id, expires, owner, polly, subject, subjectType});
}

export const apporveAccessRequest = async (accessRequestId: string) => {
    const url = approveAccessRequestUrl(accessRequestId);
    return postTemplate(url);
}

export const denyAccessRequest = async (accessRequestId: string, reason: string) => {
    const url = denyAccessRequestUrl(accessRequestId, reason);
    return postTemplate(url);
}

export const grantDatasetAccess = async (datasetId: string, expires: Date|undefined, subject: string, owner: string, subjectType: string) => {
    const url = grantAccessUrl();
    return postTemplate(url, { datasetId, expires, subject, owner, subjectType });
}

export const revokeDatasetAccess = async (accessId: string) => {
    const url = revokeAccessUrl(accessId);
    return postTemplate(url);
}
