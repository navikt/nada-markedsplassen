import { DatasetWithAccess } from '../../lib/rest/generatedDto'
import BigqueryLink from './datasource/bigqueryLink'
import MetabaseBigQueryIntegration from './metabaseBigquery'

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
