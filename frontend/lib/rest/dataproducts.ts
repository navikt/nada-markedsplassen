import {
  DataproductWithDataset,
  DatasetWithAccess,
  NewDataproduct,
  NewDataset,
  PseudoDataset,
  UpdateDataproductDto,
  UpdateDatasetDto,
  MetabaseBigQueryDatasetStatus,
} from './generatedDto'
import { deleteTemplate, fetchTemplate, HttpError, postTemplate, putTemplate } from "./request"
import { buildUrl } from "./apiUrl"
import { useQuery } from '@tanstack/react-query'

const dataproductPath = buildUrl('dataproducts')
const buildFetchDataproductUrl = (id: string) => dataproductPath(id)()
const buildCreateDataproductUrl = () => dataproductPath('new')()
const buildUpdateDataproductUrl = (id: string) => dataproductPath(id)()
const buildDeleteDataproductUrl = (id: string) => dataproductPath(id)()

const datasetPath = buildUrl('datasets')
const buildFetchDatasetUrl = (id: string) => datasetPath(id)()
const buildMetabaseBigQueryRestrictedDatasetUrl = (datasetId: string) => `${datasetPath(datasetId)()}/bigquery_restricted`
const buildMetabaseBigQueryOpenDatasetUrl = (datasetId: string) => `${datasetPath(datasetId)()}/bigquery_open`
const buildCreateDatasetUrl = () => datasetPath('new')()
const buildDeleteDatasetUrl = (id: string) => datasetPath(id)()
const buildUpdateDatasetUrl = (id: string) => datasetPath(id)()

const pseudoPath = buildUrl('datasets/pseudo')
const buildGetAccessiblePseudoDatasetsUrl = () => pseudoPath('accessible')()

const getDataproduct = async (id: string): Promise<DataproductWithDataset> =>
    fetchTemplate(buildFetchDataproductUrl(id))

const getDataset = async (id: string) =>
    fetchTemplate(buildFetchDatasetUrl(id))

export const createDataproduct = async (dp: NewDataproduct) =>
    postTemplate(buildCreateDataproductUrl(), dp)

export const updateDataproduct = async (id: string, dp: UpdateDataproductDto) =>
    putTemplate(buildUpdateDataproductUrl(id), dp)

export const deleteDataproduct = async (id: string) =>
    deleteTemplate(buildDeleteDataproductUrl(id))

export const createDataset = async (dataset: NewDataset) =>
    postTemplate(buildCreateDatasetUrl(), dataset)

export const deleteDataset = async (id: string) =>
    deleteTemplate(buildDeleteDatasetUrl(id))

export const updateDataset = async (id: string, dataset: UpdateDatasetDto) =>
    putTemplate(buildUpdateDatasetUrl(id), dataset)

const getAccessiblePseudoDatasets = async () =>
    fetchTemplate(buildGetAccessiblePseudoDatasetsUrl())

export const createMetabaseBigQueryRestrictedDataset = async (datasetId: string) =>
  postTemplate(buildMetabaseBigQueryRestrictedDatasetUrl(datasetId), {})

export const getMetabaseBigQueryRestrictedDataset = async (datasetId: string) =>
    fetchTemplate(buildMetabaseBigQueryRestrictedDatasetUrl(datasetId))

export const deleteMetabaseBigQueryRestrictedDataset = async (datasetId: string) =>
  deleteTemplate(buildMetabaseBigQueryRestrictedDatasetUrl(datasetId))

export const createMetabaseBigQueryOpenDataset = async (datasetId: string) =>
  postTemplate(buildMetabaseBigQueryOpenDatasetUrl(datasetId), {})

export const getMetabaseBigQueryOpenDataset = async (datasetId: string) =>
    fetchTemplate(buildMetabaseBigQueryOpenDatasetUrl(datasetId))

export const deleteMetabaseBigQueryOpenDataset = async (datasetId: string) =>
  deleteTemplate(buildMetabaseBigQueryOpenDatasetUrl(datasetId))

export const useGetDataproduct = (id: string, activeDataSetID?: string) =>
    useQuery<DataproductWithDataset, HttpError>({
        queryKey: ['dataproduct', id, activeDataSetID],
        queryFn: ()=>getDataproduct(id)
    })

export const useGetDataset = (id: string) => useQuery<DatasetWithAccess>({
    queryKey: ['dataset', id],
    queryFn: ()=>getDataset(id)})

export const useGetAccessiblePseudoDatasets = () =>
    useQuery<PseudoDataset[], HttpError>({
        queryKey: ['accessiblePseudoDatasets'],
        queryFn: getAccessiblePseudoDatasets
    })
