import { useRouter } from 'next/router'
import { FrontPageLogo } from '../components/index/frontPageLogo'
import { useEffect, useState } from 'react'
import amplitudeLog from '../lib/amplitude'
import Head from 'next/head'
import StoryLogo from '../components/lib/icons/storyLogo'
import Link from 'next/link'
import ProductAreaLinks from '../components/productArea/productAreaLinks'
import DataproductLogo from '../components/lib/icons/dataproductLogo'
import { Heading, Search } from '@navikt/ds-react'
import LegalGuidanceIcon from "../components/lib/icons/legalGuidanceIcon"
import GetStartedIcon from "../components/lib/icons/getStartedIcon"
import DatadrivenIcon from "../components/lib/icons/datadrivenIcon"
import { AccessRequestAlert } from '../components/user/accessRequestAlert'
import { ChevronRightIcon } from '@navikt/aksel-icons'

const SEARCH_LIMIT = 6

const LandingPage = () => {
    const router = useRouter()

    const [searchTerm, setSearchTerm] = useState('')

    useEffect(() => {
        const eventProperties = {
            sidetittel: 'hovedside',
        }
        amplitudeLog('sidevisning', eventProperties)
    }, [])

    return (
        <div className="w-screen min-h-[calc(100vh-6rem)] flex flex-col gap-8 bg-gray-100">
            <Head>
                <title>datamarkedsplassen</title>
            </Head>
            <AccessRequestAlert></AccessRequestAlert>
            <div className="bg-surface-subtle p-8 min-h-[34rem] items-center justify-center flex flex-col md:flex-row gap-8">
                <ProductAreaLinks/>

                <div className="border border-border-default bg-white rounded-lg w-11/12 md:w-[17rem] md:h-[22rem] p-4 pt-8 flex items-center flex-col gap-8">
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

                <div className="border border-border-default bg-white rounded-lg w-11/12 md:w-[17rem] md:h-[22rem] p-4 pt-8 flex items-center flex-col gap-8">
                    <GetStartedIcon />
                    <div>
                        <Heading level="2" size="small">
                            <Link href="https://docs.knada.io/">
                                Hvordan komme i gang?
                            </Link>
                        </Heading>
                        <p>Er du usikker på hva du trenger for å lage dataprodukter, eller hvordan du får dashbaord? Dokumentasjonen til plattformen ligger her.</p>
                    </div>
                </div>

                <div className="border border-border-default bg-white rounded-lg w-11/12 md:w-[17rem] md:h-[22rem] p-4 pt-8 flex items-center flex-col gap-8">
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
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 w-full md:w-[32rem]">
                    <form
                        data-umami-event="Frontpage search"
                        className="col-span-1 md:col-span-2 select-none"
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
                        className="grid grid-cols-2 items-center no-underline text-text-default border-2 border-border-on-inverted bg-white rounded shadow-sm shadow-border-on-inverted transition-all hover:shadow-lg pr-4">

                        <div className="flex flex-row items-center gap-2 md:gap-4">
                            <div className="text-deepblue-500 bg-deepblue-50">
                                <StoryLogo className="h-4 w-4 m-4 md:h-6 md:w-6 md:m-4" />
                            </div>
                            <p>Fortellinger</p>
                        </div>
                        <ChevronRightIcon className="justify-self-end md:hidden" />

                    </Link>
                    <Link
                        href="/search?preferredType=dataproduct"
                        className="grid grid-cols-2 items-center no-underline text-text-default border-2 border-border-on-inverted bg-white rounded shadow-sm shadow-border-on-inverted transition-all hover:shadow-lg pr-4">

                        <div className="flex flex-row items-center gap-2 md:gap-4">
                            <div className="text-deepblue-500 bg-deepblue-50">
                                <DataproductLogo className="h-4 w-4 m-4 md:h-6 md:w-6 md:m-4" />
                            </div>
                            <p>Produkter</p>
                        </div>
                        <ChevronRightIcon className="justify-self-end md:hidden" />

                    </Link>
                </div>
            </div>
        </div>
    );
}

export default LandingPage
