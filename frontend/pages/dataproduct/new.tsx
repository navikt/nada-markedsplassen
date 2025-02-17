import * as React from 'react'
import { NewDataproductForm } from '../../components/dataproducts/newDataproductForm'
import Head from 'next/head'
import InnerContainer from '../../components/lib/innerContainer'
import LoaderSpinner from '../../components/lib/spinner'
import { useFetchUserData } from '../../lib/rest/userData'

const NewDataproduct = () => {
  const userData = useFetchUserData()

  if(!userData?.data || userData?.isLoading){
    return <LoaderSpinner />
  }

  if (!userData?.data)
    return (
      <div>
        <h1>Du må være logget inn!</h1>
        <p>Bruk login-knappen øverst.</p>
      </div>
    )

  return (
    <InnerContainer>
      <Head>
        <title>Nytt dataprodukt</title>
      </Head>
      <NewDataproductForm />
    </InnerContainer>
  )
}

export default NewDataproduct
