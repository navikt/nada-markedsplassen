import * as React from 'react'
import humanizeDate from '../../../lib/humanizeDate'
import Copy from '../../lib/copy'
import { Heading, Link } from '@navikt/ds-react'
import { ExternalLink } from '@navikt/ds-icons'

interface DataproductTableSchemaProps {
  dataset: any
}

const DatasetMetadata = ({ dataset }: DataproductTableSchemaProps) => {
  const datasource = dataset.datasource
  const schema = datasource.schema
  if (!schema) return <div>Ingen skjemainformasjon</div>

  const entries: Array<{
    k: string
    v: string | JSX.Element
    copy?: boolean | undefined
  }> = [
    { k: 'GCP-prosjekt', v: datasource.projectID },
    { k: 'Datasett', v: datasource.dataset },
    { k: 'Tabell', v: datasource.table },
    { k: 'Tabelltype', v: datasource.tableType.toUpperCase() },
    { k: 'Oppdatert', v: humanizeDate(datasource.lastModified)},
    { k: 'Opprettet', v: humanizeDate(datasource.created) },
    {
      k: 'Link til kildekode',
      v: dataset.repo ? (
        <Link target="_blank" rel="norefferer" href={dataset.repo}>
          {dataset.repo} <ExternalLink />
        </Link>
      ) : (
        ''
      ),
    },
  ]

  datasource.expires &&
    entries.push({ k: 'Utløper', v: humanizeDate(datasource.expires) })
  datasource.description &&
    entries.push({ k: 'Beskrivelse', v: datasource.description })
  entries.push({
    k: 'URI',
    v: `${datasource.projectID}.${datasource.dataset}.${datasource.table}`,
    copy: true,
  })

  return (
    <div className="mb-3">
      <Heading level="3" size="small">
        Metadata
      </Heading>
      <>
        {entries.map(
          ({ k, v, copy }, idx) =>
            v && (
              <div className="mb-1 items-center flex gap-1" key={idx}>
                <span>{k}:</span> {v}{' '}
                {copy && <Copy text={v.toString()} />}
              </div>
            )
        )}
      </>
    </div>
  )
}
export default DatasetMetadata
