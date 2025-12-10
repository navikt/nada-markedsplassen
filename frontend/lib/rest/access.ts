import { AccessRequestsWrapper, GrantAccessData, NewAccessRequestDTO, UpdateAccessRequestDTO, UserAccesses } from "./generatedDto";
import { deleteTemplate, fetchTemplate, HttpError, postTemplate, putTemplate } from "./request";
import { buildUrl } from "./apiUrl";
import { useQuery } from '@tanstack/react-query';

const accessRequestsPath = buildUrl('accessRequests')
const buildFetchAccessRequestUrl = (datasetId: string) => accessRequestsPath()({datasetId: datasetId})
const buildCreateAccessRequestUrl = () => accessRequestsPath('new')()
const buildDeleteAccessRequestUrl = (id: string) => accessRequestsPath(id)()
const buildUpdateAccessRequestUrl = (id: string) => accessRequestsPath(id)()

const processAccessRequestsPath = buildUrl('accessRequests/process')
const buildApproveAccessRequestUrl = (accessRequestId: string) => processAccessRequestsPath(accessRequestId)({action: 'approve'})
const buildDenyAccessRequestUrl = (accessRequestId: string, reason: string) => processAccessRequestsPath(accessRequestId)({action: 'deny', reason: reason})

const accessPath = buildUrl('accesses/bigquery')
const buildGrantAccessUrl = () => accessPath('grant')()
const buildRevokeAccessUrl = (accessId: string) => accessPath('revoke')({accessId: accessId})

const metabaseAccessPath = buildUrl('accesses/metabase')
const buildGrantMetabaseAccessRestrictedUrl = () => metabaseAccessPath('grant')()
const buildGrantMetabaseAccessAllUsersUrl = () => metabaseAccessPath('grantAllUsers')()
const buildRevokeRestrictedMetabaseAccessUrl = (accessId: string) => metabaseAccessPath('revoke')({accessId: accessId})
const buildRevokeAllUsersMetabaseAccessUrl = (accessId: string) => metabaseAccessPath('revokeAllUsers')({accessId: accessId})

const userAccessesPath = buildUrl('accesses/user')()

export enum SubjectType {
    Group = 'group',
    ServiceAccount = 'serviceAccount',
    User = 'user'
}

const fetchAccessRequests = async (datasetId: string) => 
    fetchTemplate(buildFetchAccessRequestUrl(datasetId))

const fetchUserAccesses = async () => 
    fetchTemplate(userAccessesPath())

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

export const grantMetabaseAccessRestricted = async (grantAccess: GrantAccessData) => 
    postTemplate(buildGrantMetabaseAccessRestrictedUrl(), grantAccess)

export const grantMetabaseAccessAllUsers = async (grantAccess: GrantAccessData) => 
    postTemplate(buildGrantMetabaseAccessAllUsersUrl(), grantAccess)

export const revokeRestrictedMetabaseAccess = async (accessId: string) => 
    postTemplate(buildRevokeRestrictedMetabaseAccessUrl(accessId))

export const revokeAllUsersMetabaseAccess = async (accessId: string) => 
    postTemplate(buildRevokeAllUsersMetabaseAccessUrl(accessId))

export const useFetchAccessRequestsForDataset = (datasetId: string)=> useQuery<AccessRequestsWrapper, HttpError>({
    queryKey: ['accessRequests', datasetId], 
    queryFn: ()=>fetchAccessRequests(datasetId)
})

export const useFetchUserAccesses = ()=> useQuery<UserAccesses, HttpError>({
		queryKey: ['user-acceses'],
    queryFn: ()=>fetchUserAccesses()
})

