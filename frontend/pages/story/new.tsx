import * as React from 'react'
import { NewStoryForm } from '../../components/stories/newStory'
import Head from 'next/head'
import InnerContainer from '../../components/lib/innerContainer'
import LoaderSpinner from '../../components/lib/spinner'
import { useFetchUserData } from '../../lib/rest/userData'

const NewStory = () => {
  const userData = useFetchUserData()

  if(!userData?.data || userData?.loading){
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
        <title>Ny datafortelling</title>
      </Head>
      <NewStoryForm />
    </InnerContainer>
  )
}

export default NewStory

