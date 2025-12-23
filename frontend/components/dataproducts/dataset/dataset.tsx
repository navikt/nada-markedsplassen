import * as React from 'react'
import { useState } from 'react'
import EditDataset from './editDataset'
import ViewDataset from './viewDataset'
import { useGetDataset } from '../../../lib/rest/dataproducts'
import LoaderSpinner from '../../lib/spinner'
import ErrorStripe from '../../lib/errorStripe'
import Head from 'next/head'
import { DataproductWithDataset, DatasetWithAccess, GoogleGroup, UserInfo } from '../../../lib/rest/generatedDto'

export type AccessType = 'utlogget' | 'owner' | 'none' | 'user'
const findAccessType = (
  groups: GoogleGroup[],
  dataset: DatasetWithAccess | undefined,
  isOwner: boolean,
  userEmail: string,
): AccessType => {
  if (!groups) return 'utlogget'
  if (isOwner) return 'owner'
  if (!dataset) return 'none'

  const groupEmails = new Set(groups.map(g => g.email))
  
  const hasAccess = dataset.access
    .flatMap(a => a.active)
    .some(a => {
      const subject = a.subject.split(":")[1]
      return groupEmails.has(subject) || subject === userEmail
    })

  return hasAccess ? 'user' : 'none'
}

interface EntryProps {
  dataproduct: DataproductWithDataset
  datasetID: string
  userInfo: UserInfo
  isOwner: boolean
}

const Dataset = ({ datasetID, userInfo, isOwner, dataproduct }: EntryProps) => {
  const [edit, setEdit] = useState(false)
  const { data: dataset, isLoading: loading, error } = useGetDataset(datasetID)
  const accessType = findAccessType(userInfo?.googleGroups, dataset, isOwner, userInfo?.email)

  if (error) {
    return <ErrorStripe error={error}></ErrorStripe>
  }

  if (loading || !dataset) {
    return <LoaderSpinner></LoaderSpinner>
  }

  return (
    <>
      <Head>
        <title>{dataset.name}</title>
      </Head>
      {edit ? (
        <EditDataset datasetID={dataset.id} setEdit={setEdit} />
      ) : (
        <ViewDataset
          dataset={dataset}
          dataproduct={dataproduct}
          accessType={accessType}
          userInfo={userInfo}
          isOwner={isOwner}
          setEdit={setEdit}
        />
      )}
    </>
  )
}

export default Dataset
