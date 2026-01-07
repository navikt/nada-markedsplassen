import * as React from 'react'
import { useState } from 'react'
import EditDataset from './editDataset'
import ViewDataset from './viewDataset'
import { useGetDataset } from '../../../lib/rest/dataproducts'
import LoaderSpinner from '../../lib/spinner'
import ErrorStripe from '../../lib/errorStripe'
import Head from 'next/head'
import { DataproductWithDataset, DatasetWithAccess, UserInfo } from '../../../lib/rest/generatedDto'

export type AccessType = 'utlogget' | 'owner' | 'none' | 'user'
const findAccessType = (
  loggedIn: boolean,
  dataset: DatasetWithAccess | undefined,
  isOwner: boolean,
): AccessType => {
  if (!loggedIn) return 'utlogget'
  if (isOwner) return 'owner'

  const hasAccess = dataset?.access.some(a => a.active.length > 0)
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
  const accessType = findAccessType(!!userInfo, dataset, isOwner)

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
