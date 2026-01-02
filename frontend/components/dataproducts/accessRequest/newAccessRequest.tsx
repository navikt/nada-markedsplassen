import { useState } from 'react'
import AccessRequestForm from './accessRequestForm'
import { useRouter } from 'next/router'
import LoaderSpinner from '../../lib/spinner'
import { useGetDataproduct } from '../../../lib/rest/dataproducts'
import { createAccessRequest, SubjectType } from '../../../lib/rest/access'
import { NewAccessRequestDTO } from '../../../lib/rest/generatedDto'
import ErrorStripe from '../../lib/errorStripe'

interface NewAccessRequestFormProps {
  dataset: any
  setModal: (value: boolean) => void
}

const NewAccessRequestForm = ({ dataset, setModal }: NewAccessRequestFormProps) => {
  const { data: dataproduct, error: dpError, isLoading: dpLoading } = useGetDataproduct(dataset.dataproductID)
  const [error, setError] = useState<any>(null)
  const router = useRouter()

  if (dpError) return <ErrorStripe error={dpError} />
  if (dpLoading || !dataproduct) return <LoaderSpinner />

  const onSubmit = async (requestData: NewAccessRequestDTO) => {
    try {
      await createAccessRequest(
        {
          ...requestData,
          platform: 'bigquery'
        }
      )
      if (dataset.metabaseDataset && requestData.subjectType === SubjectType.User) {
        await createAccessRequest(
          {
            ...requestData,
            platform: 'metabase'
          }
        )
      }
      router.push(`/dataproduct/${dataproduct.id}/${dataset.id}`)
    } catch (e) {
      setError(e)
    }
  }

  return (
    <AccessRequestForm setModal={setModal} dataset={dataset} isEdit={false} onSubmit={onSubmit} error={error} />
  )
}

export default NewAccessRequestForm
