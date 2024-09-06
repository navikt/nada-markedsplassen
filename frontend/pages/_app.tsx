import '@uiw/react-md-editor/markdown-editor.css'
import '@uiw/react-markdown-preview/markdown.css'
import '../styles/globals.css'
import '@navikt/ds-css'
import '@navikt/ds-css-internal'
import type { AppContext, AppInitialProps } from 'next/app'
import Head from 'next/head'
import '@fontsource/source-sans-pro'
import '@fontsource/source-sans-pro/700.css'
import React from 'react'
import { UserState } from '../lib/context'
import PageLayout from '../components/pageLayout'
import { useFetchUserData } from '../lib/rest/userData'

const MyApp = ({ Component, pageProps }: AppInitialProps & AppContext) => {
  const userData = useFetchUserData()

  return (
    <UserState.Provider value={!userData.error ? userData.data : undefined}>
      <Head>
        <link
          rel="apple-touch-icon"
          sizes="180x180"
          href="/apple-touch-icon.png"
        />
        <link
          rel="icon"
          type="image/png"
          sizes="32x32"
          href="/favicon-32x32.png"
        />
        <link
          rel="icon"
          type="image/png"
          sizes="16x16"
          href="/favicon-16x16.png"
        />

        <link rel="manifest" href="/site.webmanifest" />
        <link rel="mask-icon" href="/safari-pinned-tab.svg" color="#5bbad5" />
        <meta name="msapplication-TileColor" content="#00aba9" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <meta name="theme-color" content="#ffffff" />
        <title>nav data</title>
      </Head>
      <PageLayout>
        <Component {...pageProps} />
      </PageLayout>
    </UserState.Provider>
  )
}

export default MyApp
