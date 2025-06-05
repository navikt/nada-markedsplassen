import * as React from 'react'
import { Alert, Box, Button, Heading, Loader, Panel } from '@navikt/ds-react'
import { ArrowsCirclepathIcon, CheckmarkIcon, HourglassIcon, XMarkIcon } from '@navikt/aksel-icons'
import {
  JobHeader,
  JobStateCompleted,
  JobStateFailed, JobStatePending,
  MetabaseBigQueryDatasetStatus,
} from '../../lib/rest/generatedDto'

interface MetabaseSyncProps {
  status: MetabaseBigQueryDatasetStatus
  handleReset: () => void
}

const JobStatusItem: React.FC<{ job: JobHeader }> = ({ job }) => {
  let icon = <Loader size="xsmall" title="Job is running" />
  let statusText = 'Kjører'
  let statusColor = 'bg-gray-100'

  switch (job.state) {
    case JobStateCompleted:
      icon = <CheckmarkIcon className="text-success" aria-label="Job completed successfully" />
      statusText = `Fullført ${job.endTime ? new Date(job.endTime).toLocaleString('nb-NO') : ''}`
      statusColor = 'bg-green-50'
      break
    case JobStateFailed:
      icon = <XMarkIcon className="text-error" aria-label="Job failed" />
      statusText = 'Feilet'
      statusColor = 'bg-red-50'
      break
    case JobStatePending:
      icon = <HourglassIcon title="Job is pending" />
      statusText = 'Venter'
      statusColor = 'bg-gray-200'
      break
    default:
      // Default is running
      break
  }

  return (
    <div className={`p-4 rounded ${statusColor} flex items-center gap-4 mb-2`}>
      <div className="flex-shrink-0">{icon}</div>
      <div className="flex-grow">
        <div className="font-medium">{job.kind}</div>
        <div className="text-sm text-text-subtle">{statusText}</div>
        {job.state == JobStateFailed && job.errors && job.errors.length > 0 && (
          <div className="mt-4 text-small bg-red-100 p-2 rounded">
            {job.errors[0]}
          </div>
        )}
      </div>
    </div>
  )
}

export const MetabaseSync: React.FC<MetabaseSyncProps> = ({ status, handleReset }) => {
  if (!status) return null

  const { isRunning, isCompleted, isRestricted, jobs } = status

  const sortedJobs = [...jobs].sort((a, b) => a.id - b.id)

  const completedJobs = jobs.filter(job => job.state === JobStateCompleted).length
  const failedJobs = jobs.filter(job => job.state === JobStateFailed).length

  return (
    <Box padding="4">
      <Heading level="2" size="small" spacing>
        Legger til {isRestricted ? 'tilgangsstyrt' : 'åpen'} kilde i Metabase
      </Heading>

      {isRunning && (
        <Alert variant="info" className="mb-4" size="small">
          Synkronisering pågår ({completedJobs} av {jobs.length} jobber fullført)
        </Alert>
      )}

      <div className="space-y-2">
        {sortedJobs.map((job, index) => (
          <JobStatusItem key={job.id} job={job} />
        ))}
      </div>

      {failedJobs > 0 && (
        <div>
          <Alert variant="error" className="mt-4">
            En jobb har feilet. Vennligst kontakt administrator hvis problemet vedvarer.
          </Alert>
          <div className="mt-2">
            <Button icon={<ArrowsCirclepathIcon title="Prøv igjen" onClick={handleReset}/>} >
              Prøv igjen
            </Button>
          </div>
        </div>
      )}

      {isCompleted && failedJobs === 0 && (
        <Alert variant="success" className="mt-4">
          Metabase-integrasjon fullført. Du kan nå bruke datasettet i Metabase.
        </Alert>
      )}
    </Box>
  )
}

export default MetabaseSync
