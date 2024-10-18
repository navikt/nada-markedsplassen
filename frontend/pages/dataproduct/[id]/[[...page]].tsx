import { useRouter } from 'next/router'
import Head from 'next/head'
import { useGetDataproduct } from '../../../lib/rest/dataproducts'
import { Alert } from '@navikt/ds-react'
import ErrorStripe from '../../../components/lib/errorStripe'
import { HttpError } from '../../../lib/rest/request'

interface DataproductProps {
  id: string
}

const Dataproduct = () => {
  const router = useRouter()
  const id = router.query.id as string
  console.log(id)
  const { data: dataproduct, error } = useGetDataproduct(id)

  const currentPage = router.query.page?.[0] ?? 'info'

  if (dataproduct != null) {
    router.push(`/dataproduct/${id}/${dataproduct?.slug}/${currentPage}`)
  }

  return <>
    {error?(
    <>
      <ErrorStripe error={error}></ErrorStripe>
    </>
    ): (
    <>
      <Head>
        <title>Redirigerer deg til ny side</title>
      </Head>
      <></>
    </>
    )}
  </>

}

export default Dataproduct
