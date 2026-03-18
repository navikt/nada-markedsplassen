import { ExternalLinkIcon } from '@navikt/aksel-icons'
import { Link } from '@navikt/ds-react'
import * as React from 'react'
import { BigQuery } from '../../../lib/rest/generatedDto'

interface BigqueryLinkProps {
  source: BigQuery
}

const BigqueryLink: React.FC<BigqueryLinkProps> = ({ source }) => {
  const bigQueryUrl = `https://console.cloud.google.com/bigquery?d=${source.dataset}&t=${source.table}&p=${source.projectID}&page=table`

  return (
    <Link
      className="border-l-8 border-border-on-inverted pl-4 py-1 pr-4 w-fit"
      target="_blank"
      rel="norefferer"
      href={bigQueryUrl}
    >
      Åpne i Google Cloud Console <ExternalLinkIcon />
    </Link>
  )
}

export default BigqueryLink
