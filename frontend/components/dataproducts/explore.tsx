import { DatasetWithAccess } from '../../lib/rest/generatedDto'
import BigqueryLink from './datasource/bigqueryLink'
import MetabaseBigQueryIntegration from './metabaseBigquery'

/** MappingService defines all possible service types that a dataset can be exposed to. */
export enum MappingService {
  Metabase = 'metabase'
}

interface ExploreProps {
  dataset: DatasetWithAccess
  isOwner: boolean
}

const Explore = ({ dataset, isOwner }: ExploreProps) => {
  return (
    <>
      <div className="flex flex-col w-fit">
        {dataset.datasource && <BigqueryLink source={dataset.datasource} />}
        <MetabaseBigQueryIntegration
          dataset={dataset}
          isOwner={isOwner}
          url={dataset.metabaseUrl}
          metabaseDeletedAt={dataset.metabaseDeletedAt}
        />
      </div>
    </>
  )
}
export default Explore
