import * as React from 'react'
import AccessRequestForm, { AccessRequestFormInput } from './accessRequestForm'
import { useState } from 'react'
import { updateAccessRequest } from '../../../lib/rest/access'

interface UpdateAccessRequestFormProps {
  updateAccessRequestData: any
  dataset: any
  setModal: (value: boolean) => void
}

const UpdateAccessRequest = ({
  dataset,
  updateAccessRequestData,
  setModal,
}: UpdateAccessRequestFormProps) => {
  const [error, setError] = useState<any>(null)
  const accessRequest: AccessRequestFormInput = {
    ...updateAccessRequestData,
  }

  const onSubmit = async (requestData: AccessRequestFormInput) => {
    try{
      await updateAccessRequest(
        updateAccessRequestData.id,
        requestData.expires,
        requestData.owner !== undefined ? requestData.owner : updateAccessRequestData.owner,
        requestData.polly,
        undefined,
        undefined
      )

      setModal(false)
    } catch (e) {
      setError(e)
    }
  }
  
  return (
    <AccessRequestForm
      dataset={dataset}
      accessRequest={accessRequest}
      isEdit={true}
      onSubmit={onSubmit}
      error={error}
      setModal={setModal}
    />
  )
}

export default UpdateAccessRequest
