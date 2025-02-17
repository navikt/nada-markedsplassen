import '@uiw/react-md-editor/markdown-editor.css'
import '@uiw/react-markdown-preview/markdown.css'
import '../styles/globals.css'
import '@navikt/ds-css'
import '@navikt/ds-css-internal'
import type { AppProps } from 'next/app'
import Head from 'next/head'
import '@fontsource/source-sans-pro'
import '@fontsource/source-sans-pro/700.css'
import PageLayout from '../components/pageLayout'
import { useFetchUserData } from '../lib/rest/userData'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { UserState } from '../lib/context'
import Script from 'next/script'

const UserInfo = ({ Component, pageProps }: AppProps) => {
  const userData = useFetchUserData()
  return <UserState.Provider value={!userData.error ? userData.data : undefined}>
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
      <title>Datamarkedsplassen</title>
    </Head>
    <Script defer src="https://cdn.nav.no/team-researchops/sporing/sporing.js" data-host-url="https://umami.nav.no" data-website-id="0f5ab812-053c-4776-8ef1-0ea3778b8936" />
    <PageLayout>
      <Component {...pageProps} />
    </PageLayout>
  </UserState.Provider>

}

const MyApp = ({ Component, pageProps, router }: AppProps) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        refetchOnWindowFocus: false,
        refetchOnMount: true,
        refetchOnReconnect: false,
        retry: false,
      }
    }
  })

  return (
    <QueryClientProvider client={queryClient}>
      <UserInfo Component={Component} pageProps={pageProps} router={router} />
    </QueryClientProvider>
  )
}

export default MyApp
