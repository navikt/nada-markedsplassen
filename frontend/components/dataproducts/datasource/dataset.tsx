import { TreeItem } from '@mui/x-tree-view/TreeItem';

import { Loader } from '@navikt/ds-react'
import { ExpandFilled, NextFilled } from '@navikt/ds-icons'
import Tabell from '../../lib/icons/tabell'
import { useFetchBQTables } from '../../../lib/rest/bigquery';
import { BigQueryTable } from '../../../lib/rest/generatedDto';

const DataproductTableIconMap: Record<string, () => JSX.Element> = {
  "materialized_view": Tabell,
  "table": Tabell,
  "view": Tabell,
}

export interface DataproductSourceDatasetProps {
  active: boolean
  projectID: string
  datasetID: string
}

export const Dataset = ({
  projectID,
  datasetID,
  active,
}: DataproductSourceDatasetProps) => {
  const fetchBQTables= useFetchBQTables(projectID, datasetID)

  const loadingPlaceholder = (
    <TreeItem
      slots={{ endIcon: Loader}}
      itemId={`${projectID}/${datasetID}/loadingPlaceholder`}
      label={'laster...'}
    />
  )

  const emptyPlaceholder = (
    <TreeItem
      itemId={`${projectID}/${datasetID}/emptyPlaceholder`}
      label={'ingenting her'}
    />
  )

  const datasetContents = (contents: BigQueryTable[]) =>
    contents?.map(it => (
      <TreeItem
        className="MuiTreeView-leaf"
        slots={{ endIcon: DataproductTableIconMap[it.type as string]}}
        itemId={`${projectID}/${datasetID}/${it.name}`}
        key={`${projectID}/${datasetID}/${it.name}`}
        label={it.name}
      />
    ))

  return (
    <TreeItem
      slots={{ collapseIcon: ExpandFilled, expandIcon: NextFilled}}
      itemId={`${projectID}/${datasetID}`}
      label={datasetID}
    >
      {fetchBQTables.loading
        ? loadingPlaceholder
        : fetchBQTables.bqTables?.length
        ? datasetContents(fetchBQTables?.bqTables)
        : emptyPlaceholder}
    </TreeItem>
  )
}
