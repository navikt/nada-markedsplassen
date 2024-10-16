import { useEffect, useState } from "react";
import { buildApproveAccessRequestUrl, buildCreateAccessRequestUrl, buildDeleteAccessRequestUrl, buildDenyAccessRequestUrl, buildFetchAccessRequestUrl, buildGrantAccessUrl, buildRevokeAccessUrl, buildUpdateAccessRequestUrl } from "./apiUrl";
import { AccessRequestsWrapper, GrantAccessData, NewAccessRequestDTO, UpdateAccessRequestDTO } from "./generatedDto";
import { deleteTemplate, fetchTemplate, postTemplate, putTemplate } from "./request";

export enum SubjectType {
    Group = 'group',
    ServiceAccount = 'serviceAccount',
    User = 'user'
}

export const fetchAccessRequests = async (datasetId: string) => 
    fetchTemplate(buildFetchAccessRequestUrl(datasetId))

export const createAccessRequest = async (newAccessRequest: NewAccessRequestDTO) => 
    postTemplate(buildCreateAccessRequestUrl(), newAccessRequest)

export const deleteAccessRequest = async (id: string) =>
    deleteTemplate(buildDeleteAccessRequestUrl(id))

export const updateAccessRequest = async (updateAccessRequest: UpdateAccessRequestDTO) => 
    putTemplate(buildUpdateAccessRequestUrl(updateAccessRequest.id), updateAccessRequest)

export const apporveAccessRequest = async (accessRequestId: string) => 
    postTemplate(buildApproveAccessRequestUrl(accessRequestId))

export const denyAccessRequest = async (accessRequestId: string, reason: string) => 
    postTemplate(buildDenyAccessRequestUrl(accessRequestId, reason))

export const grantDatasetAccess = async (grantAccess: GrantAccessData) => 
    postTemplate(buildGrantAccessUrl(), grantAccess)

export const revokeDatasetAccess = async (accessId: string) => 
    postTemplate(buildRevokeAccessUrl(accessId))

export const useFetchAccessRequestsForDataset = (datasetId: string)=>{
    const [data, setData] = useState<AccessRequestsWrapper| null>(null)
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
            setData(null)            
        }).finally(()=>{
            setLoading(false)
        })
    }, [datasetId])

    return {data: data?.accessRequests, loading, error}
}