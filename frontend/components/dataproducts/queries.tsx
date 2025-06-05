import { createQueryKeyStore } from '@lukemorales/query-key-factory'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { MetabaseBigQueryDatasetStatus } from '../../lib/rest/generatedDto'
import { HttpError } from '../../lib/rest/request'
import {
  createMetabaseBigQueryOpenDataset,
  createMetabaseBigQueryRestrictedDataset, deleteMetabaseBigqueryJobs,
  deleteMetabaseBigQueryOpenDataset,
  deleteMetabaseBigQueryRestrictedDataset,
  getMetabaseBigQueryOpenDataset,
  getMetabaseBigQueryRestrictedDataset,
} from '../../lib/rest/dataproducts'

export const queries = createQueryKeyStore({
  dataproducts: {
    metabaseBigQueryRestrictedDataset: (datasetId: string) => [datasetId],
    metabaseBigQueryOpenDataset: (datasetId: string) => [datasetId],
  },
})

export const useGetMetabaseBigQueryRestrictedDatasetPeriodically = (datasetId: string) =>
  useQuery<MetabaseBigQueryDatasetStatus, HttpError>({
    ...queries.dataproducts.metabaseBigQueryRestrictedDataset(datasetId),
    queryFn: () => getMetabaseBigQueryRestrictedDataset(datasetId),
    refetchInterval: 5000,
  })

export const useCreateMetabaseBigQueryRestrictedDataset = (datasetId: string) => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => createMetabaseBigQueryRestrictedDataset(datasetId),
    onSuccess: () => {
      queryClient.invalidateQueries(queries.dataproducts.metabaseBigQueryRestrictedDataset(datasetId)).then((r) => console.log(r))
    },
  })
}

export const useDeleteMetabaseBigQueryRestrictedDataset = (datasetId: string) => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => deleteMetabaseBigQueryRestrictedDataset(datasetId),
    onSuccess: () => {
      queryClient.invalidateQueries(queries.dataproducts.metabaseBigQueryRestrictedDataset(datasetId)).then((r) => console.log(r))
    },
  })
}

export const useGetMetabaseBigQueryOpenDatasetPeriodically = (datasetId: string) =>
  useQuery<MetabaseBigQueryDatasetStatus, HttpError>({
    ...queries.dataproducts.metabaseBigQueryOpenDataset(datasetId),
    queryFn: () => getMetabaseBigQueryOpenDataset(datasetId),
    refetchInterval: 5000,
  })

export const useCreateMetabaseBigQueryOpenDataset = (datasetId: string) => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => createMetabaseBigQueryOpenDataset(datasetId),
    onSuccess: () => {
      queryClient.invalidateQueries(queries.dataproducts.metabaseBigQueryOpenDataset(datasetId)).then((r) => console.log(r))
    },
  })
}

export const useDeleteMetabaseBigQueryOpenDataset = (datasetId: string) => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => deleteMetabaseBigQueryOpenDataset(datasetId),
    onSuccess: () => {
      queryClient.invalidateQueries(queries.dataproducts.metabaseBigQueryOpenDataset(datasetId)).then((r) => console.log(r))
    },
  })
}

export const useClearMetabaseBigqueryJobs = (datasetId: string) => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => deleteMetabaseBigqueryJobs(datasetId),
    onSuccess: () => {
      queryClient.invalidateQueries(queries.dataproducts.metabaseBigQueryRestrictedDataset(datasetId)).then((r) => console.log(r))
      queryClient.invalidateQueries(queries.dataproducts.metabaseBigQueryOpenDataset(datasetId)).then((r) => console.log(r))
    },
  })
}
