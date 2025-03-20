-- +goose Up

-- River pro migration 001 [up]
-- This migration consolidates the previously separate 'sequence' and 'workflow'
-- migration lines into a single 'pro' line. It is idempotent and can be
-- re-run safely from any unmigrated or partially migrated state.

-- workflow indexes
CREATE INDEX IF NOT EXISTS river_job_workflow_scheduling
    ON river_job
        USING btree (state, (metadata ->> 'workflow_id'), (metadata ->> 'task'))
    WHERE metadata ? 'workflow_id';

CREATE INDEX IF NOT EXISTS river_job_workflow_list_active
    ON river_job
        USING btree ((metadata ->> 'workflow_id'), state)
    WHERE metadata ? 'workflow_id' AND state IN ('available', 'pending', 'retryable', 'running', 'scheduled');

CREATE INDEX IF NOT EXISTS river_job_workflow_list_inactive
    ON river_job
        USING btree ((metadata ->> 'workflow_id'), state)
    WHERE metadata ? 'workflow_id' AND state IN ('cancelled', 'completed', 'discarded');

-- sequence table and indexes
CREATE UNLOGGED TABLE IF NOT EXISTS river_job_sequence
(
    id         BIGSERIAL PRIMARY KEY,
    key        TEXT                                   NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX IF NOT EXISTS river_job_sequence_key
    ON river_job_sequence
        USING btree (key, id);

CREATE INDEX IF NOT EXISTS river_job_sequence_scheduling
    ON river_job
        USING btree (state, (metadata ->> 'seq_key'), id)
    WHERE metadata ? 'seq_key';

CREATE INDEX IF NOT EXISTS idx_river_job_non_pending_seq_key_id_desc
    ON river_job
        USING btree ((metadata ->> 'seq_key'), id DESC)
    WHERE state <> 'pending' AND metadata ? 'seq_key';

-- +goose Down

-- River pro migration 001 [down]
-- Drop workflow indexes
DROP INDEX IF EXISTS river_job_workflow_list_inactive;
DROP INDEX IF EXISTS river_job_workflow_list_active;
DROP INDEX IF EXISTS river_job_workflow_scheduling;

-- Drop sequence-related indexes and table
DROP INDEX IF EXISTS idx_river_job_non_pending_seq_key_id_desc;
DROP INDEX IF EXISTS river_job_sequence_scheduling;
DROP INDEX IF EXISTS river_job_sequence_key;

DROP TABLE IF EXISTS river_job_sequence;
