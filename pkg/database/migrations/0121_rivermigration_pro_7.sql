-- +goose Up

-- River pro migration 007 [up]
DROP INDEX IF EXISTS river_job_workflow_scheduling;
DROP INDEX IF EXISTS river_job_workflow_list_active;
DROP INDEX IF EXISTS river_job_workflow_list_inactive;

-- +goose Down

-- River pro migration 007 [down]
CREATE INDEX IF NOT EXISTS river_job_workflow_scheduling
ON river_job USING btree(state, (metadata->>'workflow_id'), (metadata->>'task'))
WHERE metadata ? 'workflow_id';

CREATE INDEX IF NOT EXISTS river_job_workflow_list_active
ON river_job USING btree((metadata->>'workflow_id'), state)
WHERE metadata ? 'workflow_id'
  AND state IN ('available', 'pending', 'retryable', 'running', 'scheduled');

CREATE INDEX IF NOT EXISTS river_job_workflow_list_inactive
ON river_job USING btree((metadata->>'workflow_id'), state)
WHERE metadata ? 'workflow_id'
  AND state IN ('cancelled', 'completed', 'discarded');
