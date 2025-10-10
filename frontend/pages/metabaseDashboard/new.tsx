import * as React from 'react'
import Head from 'next/head'
import InnerContainer from '../../components/lib/innerContainer'
import LoaderSpinner from '../../components/lib/spinner'
import { useFetchUserData } from '../../lib/rest/userData'
import { NewMetabaseDashboardForm } from '../../components/metabaseDashboards/NewMetabaseDashboard'

const NewMetabaseDashboard = () => {
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
        <title>Nytt public Metabase dashboard</title>
      </Head>
      <NewMetabaseDashboardForm />
    </InnerContainer>
  )
}

export default NewMetabaseDashboard
