import React from 'react';
import { ComponentStory, ComponentMeta } from '@storybook/react';
import MetabaseSync from '../components/dataproducts/metabaseSync';
import {
  JobHeader,
  JobStateCompleted,
  JobStateFailed,
  JobStatePending,
  JobStateRunning,
  MetabaseBigQueryDatasetStatus
} from '../lib/rest/generatedDto';

export default {
  title: 'Components/MetabaseSync',
  component: MetabaseSync,
  parameters: {
    layout: 'padded',
  },
} as ComponentMeta<typeof MetabaseSync>;

const Template: ComponentStory<typeof MetabaseSync> = (args) => <MetabaseSync {...args} />;

// Helper function to create mock job data
const createJob = (id: number, kind: string, state: string, error?: string): JobHeader => {
  const job: JobHeader = {
    id,
    kind,
    state,
    startTime: new Date(Date.now() - 60000 * id).toISOString(),
    endTime: state === JobStateRunning || state === JobStatePending ? undefined : new Date().toISOString(),
  };

  if (error) {
    job.errors = [error];
  }

  return job;
};

// Running status with some completed and some running jobs
export const Running = Template.bind({});
Running.args = {
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
  } as MetabaseBigQueryDatasetStatus,
};

// Completed successfully
export const Completed = Template.bind({});
Completed.args = {
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
  } as MetabaseBigQueryDatasetStatus,
};

// Failed with errors
export const Failed = Template.bind({});
Failed.args = {
  status: {
    isRunning: false,
    isCompleted: true,
    status: 'Sync failed',
    jobs: [
      createJob(1, 'Create BigQuery Connection', JobStateCompleted),
      createJob(2, 'Create Metabase Database', JobStateCompleted),
      createJob(3, 'Sync Metabase Schema', JobStateFailed, 'Failed to sync schema: Permission denied'),
      createJob(4, 'Configure Permissions', JobStatePending),
    ],
  } as MetabaseBigQueryDatasetStatus,
};

// Just starting with pending jobs
export const Starting = Template.bind({});
Starting.args = {
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
  } as MetabaseBigQueryDatasetStatus,
};

// Mixed states
export const MixedStates = Template.bind({});
MixedStates.args = {
  status: {
    isRunning: true,
    isCompleted: false,
    status: 'Processing',
    jobs: [
      createJob(1, 'Create BigQuery Connection', JobStateCompleted),
      createJob(2, 'Create Metabase Database', JobStateFailed, 'Connection timeout'),
      createJob(3, 'Sync Metabase Schema', JobStateRunning),
      createJob(4, 'Configure Permissions', JobStatePending),
    ],
  } as MetabaseBigQueryDatasetStatus,
};

export const Retrying = Template.bind({});
Retrying.args = {
  status: {
    isRunning: true,
    isCompleted: false,
    status: 'Processing',
    jobs: [
      createJob(1, 'Create BigQuery Connection', JobStateCompleted),
      createJob(2, 'Create Metabase Database', JobStateRunning, 'Connection timeout'),
      createJob(3, 'Sync Metabase Schema', JobStatePending),
      createJob(4, 'Configure Permissions', JobStatePending),
    ],
  } as MetabaseBigQueryDatasetStatus,
};

// Empty status
export const EmptyStatus = Template.bind({});
EmptyStatus.args = {
  status: null as unknown as MetabaseBigQueryDatasetStatus,
};
