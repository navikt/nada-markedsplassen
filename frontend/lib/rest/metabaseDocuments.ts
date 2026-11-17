import { useQuery } from '@tanstack/react-query'
import {
  PublicMetabaseDocumentEditInput,
  PublicMetabaseDocumentInput,
  PublicMetabaseDocumentOutput,
} from './generatedDto'
import { buildUrl } from './apiUrl'
import { deleteTemplate, fetchTemplate, HttpError, postTemplate, putTemplate } from './request'

const metabaseDocumentPath = buildUrl('metabase-documents')
const buildCollectionUrl = () => metabaseDocumentPath()()
const buildDocumentUrl = (id: string) => metabaseDocumentPath(id)()

export const createMetabaseDocument = (input: PublicMetabaseDocumentInput) =>
  postTemplate(buildCollectionUrl(), input)

export const getMetabaseDocument = (id: string) =>
  fetchTemplate(buildDocumentUrl(id))

export const updateMetabaseDocument = (id: string, input: PublicMetabaseDocumentEditInput) =>
  putTemplate(buildDocumentUrl(id), input)

export const deleteMetabaseDocument = (id: string) =>
  deleteTemplate(buildDocumentUrl(id))

export const useGetMetabaseDocument = (id: string) =>
  useQuery<PublicMetabaseDocumentOutput, HttpError>({
    queryKey: ['metabaseDocument', id],
    queryFn: () => getMetabaseDocument(id),
    enabled: Boolean(id),
  })
