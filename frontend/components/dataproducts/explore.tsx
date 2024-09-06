import ExploreLink, { ItemType } from './exploreLink'
import { useState } from 'react'
import ErrorMessage from '../lib/error'
import { mapDatasetToServices } from '../../lib/rest/dataproducts'

/** MappingService defines all possible service types that a dataset can be exposed to. */
export enum MappingService {
  Metabase = 'metabase'
}

interface ExploreProps {
  dataproductId: string
  dataset: any
  isOwner: boolean
}

const Explore = ({ dataproductId, dataset, isOwner }: ExploreProps) => {
  const [formError, setFormError] = useState(undefined)

  const addToMetabase = async () => {
    mapDatasetToServices(dataset.id, [MappingService.Metabase]).catch(e => setFormError(e))
  }

  const removeFromMetabase = async (datasetID: string) => {
    mapDatasetToServices(datasetID, []).catch(e => setFormError(e))
  }

  const datasource = dataset.datasource
  const metabaseUrl = dataset.metabaseUrl
  const mappings = dataset.mappings
  const bigQueryUrl = `https://console.cloud.google.com/bigquery?d=${datasource.dataset}&t=${datasource.table}&p=${datasource.projectID}&page=table`

  return (
    <>
      <div className="flex flex-col">
        <ExploreLink
          datasetID={dataset.id}
          isOwner={isOwner}
          url={bigQueryUrl}
          type={ItemType.bigQuery}
        />
        <ExploreLink
          datasetID={dataset.id}
          isOwner={isOwner}
          url={metabaseUrl}
          type={ItemType.metabase}
          add={addToMetabase}
          remove={removeFromMetabase}
          mappings={mappings}
          metabaseDeletedAt={dataset.metabaseDeletedAt}
        />
      </div>
      {formError && <ErrorMessage error={formError} />}
    </>
  )
}
export default Explore
