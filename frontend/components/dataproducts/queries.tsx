import { createQueryKeyStore } from '@lukemorales/query-key-factory'
import { useEffect, useRef } from 'react'
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
  openRestrictedMetabaseBigQueryDataset,
} from '../../lib/rest/dataproducts'

export const queries = createQueryKeyStore({
  dataproducts: {
    metabaseBigQueryRestrictedDataset: (datasetId: string) => [datasetId],
    metabaseBigQueryOpenDataset: (datasetId: string) => [datasetId],
  },
})

// Invalidates the dataset query when river jobs finish (isRunning: true → false)
const useInvalidateDatasetOnJobCompletion = (datasetId: string, isRunning: boolean | undefined) => {
  const queryClient = useQueryClient()
  const prevIsRunning = useRef<boolean | undefined>(undefined)
  useEffect(() => {
    if (prevIsRunning.current === true && isRunning === false) {
      queryClient.invalidateQueries({ queryKey: ["dataset", datasetId] }).then((r) => console.log(r))
    }
    prevIsRunning.current = isRunning
  }, [isRunning, queryClient, datasetId])
}

export const useGetMetabaseBigQueryRestrictedDatasetPeriodically = (datasetId: string) => {
  const result = useQuery<MetabaseBigQueryDatasetStatus, HttpError>({
    ...queries.dataproducts.metabaseBigQueryRestrictedDataset(datasetId),
    queryFn: () => getMetabaseBigQueryRestrictedDataset(datasetId),
    refetchInterval: (query) => query.state.data?.isRunning ? 5000 : false,
  })
  useInvalidateDatasetOnJobCompletion(datasetId, result.data?.isRunning)
  return result
}

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
      queryClient.invalidateQueries({ queryKey: ["dataset", datasetId] }).then((r) => console.log(r))
    },
  })
}

export const useGetMetabaseBigQueryOpenDatasetPeriodically = (datasetId: string) => {
  const result = useQuery<MetabaseBigQueryDatasetStatus, HttpError>({
    ...queries.dataproducts.metabaseBigQueryOpenDataset(datasetId),
    queryFn: () => getMetabaseBigQueryOpenDataset(datasetId),
    refetchInterval: (query) => query.state.data?.isRunning ? 5000 : false,
  })
  useInvalidateDatasetOnJobCompletion(datasetId, result.data?.isRunning)
  return result
}

export const useCreateMetabaseBigQueryOpenDataset = (datasetId: string) => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => createMetabaseBigQueryOpenDataset(datasetId),
    onSuccess: () => {
      queryClient.invalidateQueries(queries.dataproducts.metabaseBigQueryOpenDataset(datasetId)).then((r) => console.log(r))
    },
  })
}

export const useOpenRestrictedMetabaseBigQueryDataset = (datasetId: string) => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => openRestrictedMetabaseBigQueryDataset(datasetId),
    onSuccess: () => {
      queryClient.invalidateQueries(queries.dataproducts.metabaseBigQueryOpenDataset(datasetId)).then((r) => console.log(r))
      queryClient.invalidateQueries({ queryKey: ['dataset', datasetId] }).then((r) => console.log(r))
    },
  })
}

export const useDeleteMetabaseBigQueryOpenDataset = (datasetId: string) => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => deleteMetabaseBigQueryOpenDataset(datasetId),
    onSuccess: () => {
      queryClient.invalidateQueries(queries.dataproducts.metabaseBigQueryOpenDataset(datasetId)).then((r) => console.log(r))
      queryClient.invalidateQueries({ queryKey: ["dataset", datasetId] }).then((r) => console.log(r))
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
