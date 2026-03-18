import { ChevronRightIcon } from '@navikt/aksel-icons'
import { Heading, Search } from '@navikt/ds-react'
import Head from 'next/head'
import Link from 'next/link'
import { useRouter } from 'next/router'
import { useState } from 'react'
import { FrontPageLogo } from '../components/index/frontPageLogo'
import DatadrivenIcon from "../components/lib/icons/datadrivenIcon"
import DataproductLogo from '../components/lib/icons/dataproductLogo'
import GetStartedIcon from "../components/lib/icons/getStartedIcon"
import LegalGuidanceIcon from "../components/lib/icons/legalGuidanceIcon"
import StoryLogo from '../components/lib/icons/storyLogo'
import ProductAreaLinks from '../components/productArea/productAreaLinks'
import { AccessRequestAlert } from '../components/user/accessRequestAlert'

const SEARCH_LIMIT = 6

const LandingPage = () => {
    const router = useRouter()

    const [searchTerm, setSearchTerm] = useState('')

    return (
        <div className="w-screen min-h-[calc(100vh-6rem)] flex flex-col gap-8 bg-ax-neutral-200">
            <Head>
                <title>Datamarkedsplassen</title>
            </Head>
            <AccessRequestAlert></AccessRequestAlert>
            <div className="bg-ax-bg-neutral-soft p-8 min-h-[34rem] items-center justify-center flex flex-col ax-md:flex-row gap-8">
                <ProductAreaLinks/>

                <div className="border border-ax-border-neutral bg-white rounded-lg w-11/12 ax-md:w-[17rem] ax-md:h-[22rem] p-4 pt-8 flex items-center flex-col gap-8">
                    <DatadrivenIcon />
                    <div>
                        <Heading level="2" size="small">
                            <Link href="https://aksel.nav.no/god-praksis/artikler/malinger-i-produktutvikling?tema=produktledelse">
                                Hva er datadrevet?
                            </Link>
                        </Heading>
                        <p>På Aksel kan du lese om datadrevet produktutvikling. En inspirasjon for å komme i gang med målinger i ditt team?</p>
                        </div>
                </div>

                <div className="border border-ax-border-neutral bg-white rounded-lg w-11/12 ax-md:w-[17rem] ax-md:h-[22rem] p-4 pt-8 flex items-center flex-col gap-8">
                    <GetStartedIcon />
                    <div>
                        <Heading level="2" size="small">
                            <Link href="https://docs.knada.io/">
                                Hvordan komme i gang?
                            </Link>
                        </Heading>
                        <p>Er du usikker på hva du trenger for å lage dataprodukter, eller hvordan du får dashboard? Dokumentasjonen til plattformen ligger her.</p>
                    </div>
                </div>

                <div className="border border-ax-border-neutral bg-white rounded-lg w-11/12 ax-md:w-[17rem] ax-md:h-[22rem] p-4 pt-8 flex items-center flex-col gap-8">
                    <LegalGuidanceIcon />
                    <div>
                        <Heading level="2" size="small">
                            <Link
                                href="https://docs.knada.io/juridisk/spilleregler/"
                                className="underline"
                            >
                                Usikker på jus og data?
                            </Link>
                        </Heading>
                        <p>For at folk skal være trygge rundt deling og bruk av data med personopplysninger, har vi forsøkt å svare på noen spørsmål.</p>
                    </div>
                </div>
            </div>
            <div className="flex flex-col items-center gap-8 w-11/12 self-center pb-8">
                <FrontPageLogo />
                <div className="grid grid-cols-1 ax-md:grid-cols-2 gap-4 w-full ax-md:w-[32rem]">
                    <form
                        data-umami-event="Frontpage search"
                        className="col-span-1 ax-md:col-span-2 select-none"
                        role="search" 
                        onSubmit={e =>{
                            e.preventDefault()
                            router.push({ pathname: '/search', query: { text: searchTerm, preferredType: 'story' } })
                        }}>
                        <Search
                            label="Søk etter dataprodukter eller fortellinger"
                            onChange={(text) =>
                                setSearchTerm(text)
                            }
                        />
                    </form>
                    <Link
                        href="/search?preferredType=story"
                        className="grid grid-cols-2 items-center no-underline text-ax-text-neutral border-2 border-ax-border-neutral-subtle bg-white rounded-sm shadow-xs shadow-ax-border-neutral-subtle transition-all hover:shadow-lg pr-4">

                        <div className="flex flex-row items-center gap-2 ax-md:gap-4">
                            <div className="text-ax-brand-blue-600 bg-ax-brand-blue-100">
                                <StoryLogo className="h-4 w-4 m-4 ax-md:h-6 ax-md:w-6 ax-md:m-4" />
                            </div>
                            <p>Fortellinger</p>
                        </div>
                        <ChevronRightIcon className="justify-self-end ax-md:hidden" />

                    </Link>
                    <Link
                        href="/search?preferredType=dataproduct"
                        className="grid grid-cols-2 items-center no-underline text-ax-text-neutral border-2 border-ax-border-neutral-subtle bg-white rounded-sm shadow-xs shadow-ax-border-neutral-subtle transition-all hover:shadow-lg pr-4">

                        <div className="flex flex-row items-center gap-2 ax-md:gap-4">
                            <div className="text-ax-brand-blue-600 bg-ax-brand-blue-100">
                                <DataproductLogo className="h-4 w-4 m-4 ax-md:h-6 ax-md:w-6 ax-md:m-4" />
                            </div>
                            <p>Produkter</p>
                        </div>
                        <ChevronRightIcon className="justify-self-end ax-md:hidden" />

                    </Link>
                </div>
            </div>
        </div>
    );
}

export default LandingPage
