import { PlusCircleIcon } from '@navikt/aksel-icons'
import { Heading } from '@navikt/ds-react'
import Head from 'next/head'
import { useRouter } from 'next/router'
import { useContext, useEffect, useState } from 'react'
import DataproductOwnerMenu from '../../../../components/dataproducts/dataproductOwnerMenu'
import { DataproductSidebar } from '../../../../components/dataproducts/dataproductSidebar'
import Dataset from '../../../../components/dataproducts/dataset/dataset'
import NewDatasetForm from '../../../../components/dataproducts/dataset/newDatasetForm'
import DeleteModal from '../../../../components/lib/deleteModal'
import { Description } from '../../../../components/lib/detailTypography'
import ErrorStripe from '../../../../components/lib/errorStripe'
import InnerContainer from '../../../../components/lib/innerContainer'
import LoaderSpinner from '../../../../components/lib/spinner'
import TabPanel, { TabPanelType } from '../../../../components/lib/tabPanel'
import { UserState } from '../../../../lib/context'
import { deleteDataproduct, useGetDataproduct } from '../../../../lib/rest/dataproducts'
import { truncate } from '../../../../lib/stringUtils'


const Dataproduct = () => {
  const router = useRouter()
  const id = router.query.id as string
  const pageParam = router.query.page?.[0] ?? 'info'
  const [showDelete, setShowDelete] = useState(false)
  const [deleteError, setDeleteError] = useState('')

  const { data: dataproduct, isLoading: loading, error } = useGetDataproduct(id, pageParam)

  const userInfo = useContext(UserState)

  const isOwner =
    !userInfo?.googleGroups
      ? false
      : userInfo.googleGroups.some(
        (g: any) => g.email === dataproduct?.owner?.group
      )

  useEffect(() => {
    const eventProperties = {
      sidetittel: 'produktside',
      title: dataproduct?.name,
    }
  })

  const onDelete = async () => {
    if (!dataproduct) return
    deleteDataproduct(dataproduct.id).then(() => {
      router.push('/')
    }).catch(error => {
      setDeleteError(error)
    })
  }

  if (error) return <ErrorStripe error={error} />
  if (loading || !dataproduct)
    return <LoaderSpinner />

  const menuItems: Array<{
    title: any
    slug: string
    component: any
  }> = [
      {
        title: 'Beskrivelse',
        slug: 'info',
        component: (
          <Description
            dataproduct={dataproduct}
            isOwner={isOwner}
          />
        ),
      },
    ]

  if (dataproduct.datasets) {
    dataproduct.datasets.forEach((dataset: any) => {
      menuItems.push({
        title: `${truncate(dataset.name, 120)}`,
        slug: dataset.id,
        component: (
          <Dataset
            datasetID={dataset.id}
            userInfo={userInfo}
            isOwner={isOwner}
            dataproduct={dataproduct}
          />
        ),
      })
    })
  }

  if (isOwner) {
    menuItems.push({
      title: (
        <div className="flex flex-row items-center text-base">
          <PlusCircleIcon className="mr-2" />
          Legg til datasett
        </div>
      ),
      slug: 'new',
      component: <NewDatasetForm dataproduct={dataproduct} />,
    })
  }

  const currentPage = menuItems
    .map((e) => e.slug)
    .indexOf(pageParam)
  return (
    <InnerContainer>
      <Head>
        <title>{dataproduct.name}</title>
      </Head>
      <div className='flex flex-row items-center border-b-[1px] border-border-on-inverted'>
        <Heading size='xlarge'>
          {dataproduct.name}
        </Heading>
          <DataproductOwnerMenu dataproduct={dataproduct} className='ml-2'/>
      </div>
      <div className="flex flex-row h-full grow">
        <DataproductSidebar
          product={dataproduct}
          isOwner={isOwner}
          menuItems={menuItems}
          currentPage={currentPage}
        />
        <div className="md:pl-4 grow md:border-l border-border-on-inverted">
          {menuItems.map((i, idx) => (
            <TabPanel
              key={idx}
              value={currentPage}
              index={idx}
              type={TabPanelType.simple}
            >
              {i.component}
            </TabPanel>
          ))}
          <DeleteModal
            open={showDelete}
            onCancel={() => setShowDelete(false)}
            onConfirm={() => onDelete()}
            name={dataproduct.name}
            error={deleteError}
            resource="dataprodukt"
          />
        </div>
      </div>
    </InnerContainer>
  )
}

export default Dataproduct
