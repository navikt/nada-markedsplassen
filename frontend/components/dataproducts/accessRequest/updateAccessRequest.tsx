import * as React from 'react'
import AccessRequestForm, { AccessRequestFormInput } from './accessRequestForm'
import { useState } from 'react'
import { updateAccessRequest } from '../../../lib/rest/access'
import { NewAccessRequestDTO, UpdateAccessRequestDTO } from '../../../lib/rest/generatedDto'

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

  const onSubmit = async (requestData: NewAccessRequestDTO) => {
    try{
      await updateAccessRequest(
        {
          id: updateAccessRequestData.id,/* uuid */
          owner: requestData.owner !== undefined ? requestData.owner : updateAccessRequestData.owner,
          expires: requestData.expires,/* RFC3339 */
          polly: requestData.polly,
        }
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
