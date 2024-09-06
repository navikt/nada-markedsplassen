import User from './header/user'
import { useRouter } from 'next/router'
import { Header } from '@navikt/ds-react-internal'
import React, { useState } from 'react'
import { Search } from '@navikt/ds-react'
import Link from 'next/link'
import { HeaderLogo } from './index/frontPageLogo'

export const PageLayout = ({ children }: { children: React.ReactNode }) => {
  const router = useRouter()
  const [searchTerm, setSearchTerm] = useState('')

  return (
    <div className="min-h-screen flex flex-col">
      <Header className="flex flex-row justify-between">
        <div className="flex flex-row">
          <Header.Title href="/">
            <div className="cursor-pointer w-8 md:w-fit flex items-center">
              <HeaderLogo />
            </div>
          </Header.Title>
          <form
            className="self-center px-5"
            onSubmit={(e) => {
              e.preventDefault()
              router.push({
                pathname: '/search',
                query: { text: searchTerm, preferredType: 'story' },
              })
            }}
          >
            <Search
              label="header søk"
              size="small"
              variant="simple"
              placeholder="Søk"
              onChange={(text) => setSearchTerm(text)}
            />
          </form>
        </div>
        <div className="flex flex-row min-w-fit">
          <User />
        </div>
      </Header>
      <main className="md:w-screen flex flex-col items-center">{children}</main>
      <footer className="flex gap-4 justify-center items-center border-t border-border-on-inverted bg-surface-subtle min-h-[3rem] mt-auto">
        <Link href="/about">Om Datamarkedsplassen</Link>
        <a href="https://docs.knada.io/">Docs</a>
      </footer>
    </div>
  )
}

export default PageLayout
