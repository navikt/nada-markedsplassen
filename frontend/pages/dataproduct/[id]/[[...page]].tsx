import { useRouter } from 'next/router'
import Head from 'next/head'
import { useGetDataproduct } from '../../../lib/rest/dataproducts'
import { Alert } from '@navikt/ds-react'

interface DataproductProps {
  id: string
}

const Dataproduct = () => {
  const router = useRouter()
  const id = router.query.id as string
  const {dataproduct, loading, error} = useGetDataproduct(id)

  const currentPage = router.query.page?.[0] ?? 'info'

  if(dataproduct!= null){
    router.push(`/dataproduct/${id}/${dataproduct?.slug}/${currentPage}`)
  }

  return error?(
    <>
      <Head>
        <Alert variant='error'>Internal error: {error}</Alert>
      </Head>
      <></>
    </>
  ): (
    <>
      <Head>
        <title>Redirigerer deg til ny side</title>
      </Head>
      <></>
    </>
  )
}

export default Dataproduct
