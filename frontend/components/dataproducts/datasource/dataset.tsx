import { TreeItem } from '@mui/x-tree-view/TreeItem';

import { ChevronDownIcon, ChevronRightIcon } from '@navikt/aksel-icons';
import { Loader } from '@navikt/ds-react';
import { useFetchBQTables } from '../../../lib/rest/bigquery';
import { BigQueryTable } from '../../../lib/rest/generatedDto';
import Tabell from '../../lib/icons/tabell';

const DataproductTableIconMap: Record<string, () => React.JSX.Element> = {
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
      slots={{ collapseIcon: ChevronRightIcon, expandIcon: ChevronDownIcon}}
      itemId={`${projectID}/${datasetID}`}
      label={datasetID}
    >
      {fetchBQTables.isLoading
        ? loadingPlaceholder
        : fetchBQTables.data?.length
        ? datasetContents(fetchBQTables?.data)
        : emptyPlaceholder}
    </TreeItem>
  )
}
