import * as React from 'react'
import { Alert, Box, Heading, Loader, Panel } from '@navikt/ds-react'
import { CheckmarkIcon, HourglassIcon, XMarkIcon } from '@navikt/aksel-icons'
import {
  JobHeader,
  JobStateCompleted,
  JobStateFailed, JobStatePending,
  MetabaseBigQueryDatasetStatus,
} from '../../lib/rest/generatedDto'

interface MetabaseSyncProps {
  status: MetabaseBigQueryDatasetStatus
}

const JobStatusItem: React.FC<{ job: JobHeader }> = ({ job }) => {
  let icon = <Loader size="xsmall" title="Job is running" />
  let statusText = 'Kjører'
  let statusColor = 'bg-gray-100'

  switch (job.state) {
    case JobStateCompleted:
      icon = <CheckmarkIcon className="text-success" aria-label="Job completed successfully" />
      statusText = `Fullført ${job.endTime ? new Date(job.endTime).toLocaleString('nb-NO') : ''}`
      statusColor = 'bg-success-50'
      break
    case JobStateFailed:
      icon = <XMarkIcon className="text-error" aria-label="Job failed" />
      statusText = 'Feilet'
      statusColor = 'bg-error-50'
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
      </div>
      {job.errors && job.errors.length > 0 && (
        <Alert variant="error" size="small" className="mt-1">
          {job.errors[0]}
        </Alert>
      )}
    </div>
  )
}

export const MetabaseSync: React.FC<MetabaseSyncProps> = ({ status }) => {
  if (!status) return null

  const { isRunning, isCompleted, jobs } = status

  const sortedJobs = [...jobs].sort((a, b) => a.id - b.id)

  const completedJobs = jobs.filter(job => job.state === JobStateCompleted).length
  const failedJobs = jobs.filter(job => job.state === JobStateFailed).length

  return (
    <Box padding="4" borderRadius="small">
      <Heading level="2" size="small" spacing>
        Legger til tilgangsstyrt datasett i Metabase
      </Heading>

      {isRunning && (
        <Alert variant="info" className="mb-4">
          Synkronisering pågår ({completedJobs} av {jobs.length} jobber fullført)
        </Alert>
      )}

      <div className="space-y-2">
        {sortedJobs.map((job, index) => (
          <JobStatusItem key={job.id} job={job} />
        ))}
      </div>

      {failedJobs > 0 && (
        <Alert variant="error" className="mt-4">
          En eller flere jobber feilet. Vennligst kontakt administrator hvis problemet vedvarer.
        </Alert>
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
