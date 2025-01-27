import { TreeItem } from '@mui/x-tree-view/TreeItem';
import { Loader } from '@navikt/ds-react'
import { Dataset } from './dataset'
import { useFetchBQDatasets } from '../../../lib/rest/bigquery';
import { ChevronDownIcon, ChevronRightIcon } from '@navikt/aksel-icons';

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
      slots={{ collapseIcon: ChevronRightIcon, expandIcon: ChevronDownIcon}}
      itemId={projectID}
      label={projectID}
    >
      {fetchBQDatasets.isLoading
        ? loadingPlaceholder
        : !fetchBQDatasets.data?.length
        ? emptyPlaceholder
        : fetchBQDatasets.data?.map((datasetID) => (
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
