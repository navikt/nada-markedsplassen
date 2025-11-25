import * as React from 'react'
import { useState } from 'react'
import EditDataset from './editDataset'
import ViewDataset from './viewDataset'
import { useGetDataset } from '../../../lib/rest/dataproducts'
import LoaderSpinner from '../../lib/spinner'
import ErrorStripe from '../../lib/errorStripe'
import Head from 'next/head'

const findAccessType = (
  groups: any,
  dataset: any,
  isOwner: boolean,
  userEmail: string,
) => {
  if (!groups) return { type: 'utlogget' }
  if (isOwner) return { type: 'owner' }
  if (!dataset) return {type: 'none'}

  let ret = { type: 'none', expires: null }
  dataset.access.forEach((a: any) => {
    if (groups.includes(a.subject)) {
      ret = { type: 'user', expires: a.expires }
    } else if (a.subject === "user:"+userEmail) {
      ret = { type: 'user', expires: a.expires }
    }
  })

  return ret
}

interface EntryProps {
  dataproduct: any
  datasetID: string
  userInfo: any
  isOwner: boolean
}

const Dataset = ({ datasetID, userInfo, isOwner, dataproduct }: EntryProps) => {
  const [edit, setEdit] = useState(false)
  const {data: dataset, isLoading: loading, error} = useGetDataset(datasetID)
  const accessType = findAccessType(userInfo?.googleGroups, dataset, isOwner, userInfo?.email)

  if(error){
    return <ErrorStripe error={error}></ErrorStripe>
  }

  if(loading || !dataset){
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
