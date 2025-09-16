import { Meta } from '@storybook/nextjs'
import MetabaseSync from '../components/dataproducts/metabaseSync'
import {
  JobHeader,
  JobStateCompleted,
  JobStateFailed,
  JobStatePending,
  JobStateRunning,
  MetabaseBigQueryDatasetStatus,
} from '../lib/rest/generatedDto'

export default {
  title: 'Components/MetabaseSync',
  component: MetabaseSync,
  parameters: {
    layout: 'padded',
  },
} as Meta<typeof MetabaseSync>

// Helper function to create mock job data
const createJob = (
  id: number,
  kind: string,
  state: string,
  error?: string,
): JobHeader => {
  const job: JobHeader = {
    id,
    kind,
    state,
    startTime: new Date(Date.now() - 60000 * id).toISOString(),
    endTime: state === JobStateRunning || state === JobStatePending
      ? undefined
      : new Date().toISOString(),
    duplicate: false,
    errors: []
  }

  if (error) {
    job.errors = [error]
  }

  return job
}

export const Running = {
  args: {
    status: {
      isRunning: true,
      isCompleted: false,
      status: 'Syncing dataset to Metabase',
      jobs: [
        createJob(1, 'Create BigQuery Connection', JobStateCompleted),
        createJob(2, 'Create Metabase Database', JobStateCompleted),
        createJob(3, 'Sync Metabase Schema', JobStateRunning),
        createJob(4, 'Configure Permissions', JobStatePending),
      ],
    } as unknown as MetabaseBigQueryDatasetStatus,
  },
}

export const Completed = {
  args: {
    status: {
      isRunning: false,
      isCompleted: true,
      status: 'Sync completed successfully',
      jobs: [
        createJob(1, 'Create BigQuery Connection', JobStateCompleted),
        createJob(2, 'Create Metabase Database', JobStateCompleted),
        createJob(3, 'Sync Metabase Schema', JobStateCompleted),
        createJob(4, 'Configure Permissions', JobStateCompleted),
      ],
    } as unknown as MetabaseBigQueryDatasetStatus,
  },
}

export const Failed = {
  args: {
    status: {
      isRunning: false,
      isCompleted: true,
      status: 'Sync failed',
      jobs: [
        createJob(1, 'Create BigQuery Connection', JobStateCompleted),
        createJob(2, 'Create Metabase Database', JobStateCompleted),
        createJob(
          3,
          'Sync Metabase Schema',
          JobStateFailed,
          'Failed to sync schema: Permission denied'
        ),
        createJob(4, 'Configure Permissions', JobStatePending),
      ],
    } as unknown as MetabaseBigQueryDatasetStatus,
  },
}

export const Starting = {
  args: {
    status: {
      isRunning: true,
      isCompleted: false,
      status: 'Starting sync process',
      jobs: [
        createJob(1, 'Create BigQuery Connection', JobStatePending),
        createJob(2, 'Create Metabase Database', JobStatePending),
        createJob(3, 'Sync Metabase Schema', JobStatePending),
        createJob(4, 'Configure Permissions', JobStatePending),
      ],
    } as unknown as MetabaseBigQueryDatasetStatus,
  },
}

export const MixedStates = {
  args: {
    status: {
      isRunning: true,
      isCompleted: false,
      status: 'Processing',
      jobs: [
        createJob(1, 'Create BigQuery Connection', JobStateCompleted),
        createJob(
          2,
          'Create Metabase Database',
          JobStateFailed,
          'Connection timeout'
        ),
        createJob(3, 'Sync Metabase Schema', JobStateRunning),
        createJob(4, 'Configure Permissions', JobStatePending),
      ],
    } as unknown as MetabaseBigQueryDatasetStatus,
  },
}

export const Retrying = {
  args: {
    status: {
      isRunning: true,
      isCompleted: false,
      status: 'Processing',
      jobs: [
        createJob(1, 'Create BigQuery Connection', JobStateCompleted),
        createJob(
          2,
          'Create Metabase Database',
          JobStateRunning,
          'Connection timeout'
        ),
        createJob(3, 'Sync Metabase Schema', JobStatePending),
        createJob(4, 'Configure Permissions', JobStatePending),
      ],
    } as unknown as MetabaseBigQueryDatasetStatus,
  },
}

export const EmptyStatus = {
  args: {
    status: null as unknown as MetabaseBigQueryDatasetStatus,
  },
}
