import { TreeItem } from '@mui/x-tree-view/TreeItem';
import { Loader } from '@navikt/ds-react'
import { ExpandFilled, NextFilled } from '@navikt/ds-icons'
import { Dataset } from './dataset'
import { useFetchBQDatasets } from '../../../lib/rest/bigquery';

export interface DataproductSourceDatasetProps {
  activePaths: string[]
  projectID: string
}

export const Project = ({
  projectID,
  activePaths,
}: DataproductSourceDatasetProps) => {
  const fetchBQDatasets= useFetchBQDatasets(projectID)

  const emptyPlaceholder = (
    <TreeItem
      itemId={`${projectID}/emptyPlaceholder`}
      label={'ingen datasett i prosjekt'}
    />
  )

  const loadingPlaceholder = (
    <TreeItem
      slots={{ endIcon: Loader}}
      itemId={`${projectID}/loadingPlaceholder`}
      label={'laster...'}
    />
  )

  return (
    <TreeItem
      slots={{ collapseIcon: ExpandFilled, expandIcon: NextFilled}}
      itemId={projectID}
      label={projectID}
    >
      {fetchBQDatasets.loading
        ? loadingPlaceholder
        : !fetchBQDatasets.bqDatasets?.length
        ? emptyPlaceholder
        : fetchBQDatasets.bqDatasets?.map((datasetID) => (
            <Dataset
              key={datasetID}
              projectID={projectID}
              datasetID={datasetID}
              active={activePaths.includes(`${projectID}/${datasetID}`)}
            />
          ))}
    </TreeItem>
  )
}
