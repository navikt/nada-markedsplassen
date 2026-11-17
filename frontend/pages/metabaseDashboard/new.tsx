import Head from 'next/head'
import InnerContainer from '../../components/lib/innerContainer'
import LoaderSpinner from '../../components/lib/spinner'
import { NewMetabaseDashboardForm } from '../../components/metabaseDashboards/NewMetabaseDashboard'
import { useFetchUserData } from '../../lib/rest/userData'

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
        <title>Legg til ny offentlig Metabase-lenke</title>
      </Head>
      <NewMetabaseDashboardForm />
    </InnerContainer>
  )
}

export default NewMetabaseDashboard
