import LoaderSpinner from '../../../../components/lib/spinner'
import * as React from 'react'
import { useContext, useEffect, useState } from 'react'
import amplitudeLog from '../../../../lib/amplitude'
import EditDataproduct from '../../../../components/dataproducts/editDataproduct'
import Head from 'next/head'
import TopBar from '../../../../components/lib/topBar'
import { UserState } from '../../../../lib/context'
import DeleteModal from '../../../../components/lib/deleteModal'
import { useRouter } from 'next/router'
import InnerContainer from '../../../../components/lib/innerContainer'
import { deleteDataproduct, useGetDataproduct } from '../../../../lib/rest/dataproducts'
import ErrorStripe from '../../../../components/lib/errorStripe'

const DataproductEdit = () => {
  const [showDelete, setShowDelete] = useState(false)
  const [deleteError, setDeleteError] = useState('')
  const userInfo = useContext(UserState)
  const router = useRouter()
  const id = router.query.id as string

  const { data: dataproduct, isLoading: loading, error } = useGetDataproduct(id)
  useEffect(() => {
    const eventProperties = {
      sidetittel: 'productEdit',
      title: dataproduct?.name,
    }
    amplitudeLog('sidevisning', eventProperties)
  }, [dataproduct?.name])

  if (error) return <ErrorStripe error={error} />

  if (loading || !dataproduct) return <LoaderSpinner />

  const isOwner =
    !userInfo?.googleGroups
      ? false
      : userInfo.googleGroups.some(
        (g: any) => g.email === dataproduct?.owner?.group
      )

  const onDelete = async () => {
    deleteDataproduct(dataproduct.id).then(() => {
      amplitudeLog('slett dataprodukt', { name: dataproduct.name })
      router.push('/')
    }).catch(error => {
      amplitudeLog('slett dataprodukt feilet', { name: dataproduct.name })
      setDeleteError(error)
    })
  }

  return (
    <InnerContainer>
      <Head>
        <title>{dataproduct.name}</title>
      </Head>
      <TopBar name={dataproduct.name}>
        {isOwner && (
          <div className="flex gap-2">
            <p className="font-bold px-2 border-r border-border-strong">
              Endre dataprodukt
            </p>
            <a href="#" onClick={() => setShowDelete(true)}>
              Slette dataprodukt
            </a>
          </div>
        )}
      </TopBar>
      <div className="flex flex-col h-full flex-grow">
        <EditDataproduct product={dataproduct} />
      </div>
      <DeleteModal
        open={showDelete}
        onCancel={() => setShowDelete(false)}
        onConfirm={() => onDelete()}
        name={dataproduct.name}
        error={deleteError}
        resource="dataprodukt"
      />
    </InnerContainer>
  )
}

export default DataproductEdit
