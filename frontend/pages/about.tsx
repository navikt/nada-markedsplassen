import { Heading } from '@navikt/ds-react'
import Head from 'next/head'

const About = () => {

    return (
        <div className="flex flex-col md:max-w-2xl px-4 py-8">
            <Head>
                <title>About</title>
            </Head>
            <Heading size="xlarge" level="1" spacing>Om Nav Data</Heading>
            <div className="flex flex-col gap-2">
                <p>
                    Nav Data er Navs markedsplass for deling av data og innsikt.<br /> Dette er stedet hvor vi ønsker
                    at
                    alle
                    team
                    i Nav skal dele data fra sine domener for analytisk bruk.<br /> Det er også et sted hvor innsikt
                    man
                    finner i
                    data kan deles.
                </p>
                <p>

                    Data deles fra teamene som <a
                        href={'https://docs.knada.io/dataprodukter/dataprodukt/'}>dataprodukter</a>.<br />
                    Dataprodukter som ikke inneholder personopplysninger eller andre sensitive data vil som regel
                    kunne
                    være
                    åpent tilgjengelig for alle ansatte i Nav å ta i bruk.<br />
                    Dataprodukter av mer sensitiv art vil typisk ha mer begrensede tilganger, men felles for alle
                    dataprodukter er at alle ansatte i Nav kan lese metadata
                    og beskrivelser av hvilke data som finnes. Dette skal gjøre det enkelt å finne de dataene man
                    har
                    behov
                    for.
                </p>
                <p>
                    Innsikt fra data kan deles på markedsplassen som <a
                        href={'https://docs.knada.io/analyse/datafortellinger/'}>datafortellinger</a> eller <a
                            href={'https://docs.knada.io/analyse/metabase/'}>spørsmål og dashboards i
                        visualiseringsverktøyet Metabase</a>.
                    Utdypende dokumentasjon for datamarkedsplassen og tilknyttede verktøy finnes på <a
                        href={'https://docs.knada.io/'}>docs.knada.io</a>.</p>
                <p>
                    Spørsmål og diskusjoner om markedsplassen foregår på slack i <a
                        href={'https://nav-it.slack.com/archives/CGRMQHT50'}>#nada</a>
                </p>
            </div>
        </div>
    )
}
export default About
