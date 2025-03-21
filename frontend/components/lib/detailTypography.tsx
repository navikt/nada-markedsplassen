import { Accordion, Alert, Heading, Link } from '@navikt/ds-react'
import { useRouter } from 'next/router'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import TagPill from './tagPill'

interface DescriptionProps {
    dataproduct: any,
    isOwner: boolean
}

export const Description = ({dataproduct, isOwner}: DescriptionProps) => {
    const router = useRouter()
    const handleChange = (newSlug: string) => {
        router.push(`/dataproduct/${dataproduct.id}/${dataproduct.slug}/${newSlug}`)
      }

    return (<div className="mt-8 flex flex-col gap-4 max-w-4xl">
        {isOwner && dataproduct.datasets.length== 0 && <Alert variant="warning">Det er ikke noe datasett i dataproduktet. Vil du <Link href={ `/dataproduct/${dataproduct.id}/${dataproduct.slug}/new`}>legge til et datasett</Link>?</Alert>}

        <Accordion className="block md:hidden w-full">
            <Accordion.Item defaultOpen={true}>
                <Accordion.Header>Beskrivelse</Accordion.Header>
                <Accordion.Content>
                    <div className="spaced-out text-justify">
                        <ReactMarkdown remarkPlugins={[remarkGfm]}>
                            {dataproduct.description || '*ingen beskrivelse*'}
                        </ReactMarkdown>
                    </div>
                </Accordion.Content>
            </Accordion.Item>
        </Accordion>
        <div className="spaced-out hidden md:block text-justify">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
                {dataproduct.description || '*ingen beskrivelse*'}
            </ReactMarkdown>
        </div>
        {!!dataproduct.keywords.length && (
            <div className="flex flex-row gap-1 flex-wrap my-2">
                {dataproduct.keywords.map((k: string, i: number) => (
                    <TagPill key={k} keyword={k} onClick={() => { router.push(`/search?keywords=${k}`) }}>
                        {k}
                    </TagPill>
                ))}
            </div>
        )}
        <div className="block md:hidden">
            <Heading level="2" size="medium">Datasett</Heading>
            <div className="flex flex-col gap-2">
                {dataproduct.datasets.map((dataset: any, idx: number) => <a
                  className=""
                  href="#"
                  key={idx}
                  onClick={() => handleChange(dataset.id)}
                >
                  {dataset.name}
                </a>)}
            </div>
        </div>
    </div>)
}


