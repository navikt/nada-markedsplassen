-- +goose Up

-- River pro migration 006 [up]
--
-- partition key triggers
--

-- Detects whether partition_key is a generated column (from the original
-- version of migration 003) and converts it to a plain column with triggers.
-- If partition_key is already a plain column (fresh installs running the
-- updated 003), this migration is a no-op.

-- +goose StatementBegin
DO $$
BEGIN
    -- Check if partition_key is a generated column. Uses pg_attribute
    -- with a regclass cast so schema resolution works naturally through
    -- the existing  mechanism.
    IF EXISTS (
        SELECT 1
        FROM pg_attribute
        WHERE attrelid = 'river_job'::regclass
          AND attname = 'partition_key'
          AND attgenerated = 's'
    ) THEN
        -- Convert from generated to plain column in place, preserving
        -- existing data and indexes.
        ALTER TABLE river_job
            ALTER COLUMN partition_key DROP EXPRESSION;

        -- Keep partition_key in sync with metadata. Two triggers share one
        -- function: one for INSERT, one for UPDATE. Both use WHEN clauses
        -- evaluated in C before the plpgsql function is invoked, so the
        -- common case (non-partition inserts and metadata updates that don't
        -- touch the partition key) pays zero function-call overhead.
        CREATE OR REPLACE FUNCTION river_job_partition_key_column()
        RETURNS trigger
        LANGUAGE plpgsql
        AS $fn$
        BEGIN
            NEW.partition_key := NEW.metadata->>'river:partition';
            RETURN NEW;
        END;
        $fn$;

        -- INSERT: fire only for partition jobs (non-partition inserts skip
        -- entirely).
        CREATE OR REPLACE TRIGGER river_job_set_partition_key_on_insert
        BEFORE INSERT ON river_job
        FOR EACH ROW
        WHEN (
            NEW.metadata->>'river:partition' IS NOT NULL
        )
        EXECUTE FUNCTION river_job_partition_key_column();

        -- UPDATE: fire only when the partition key actually changes in
        -- metadata.
        CREATE OR REPLACE TRIGGER river_job_set_partition_key_on_update
        BEFORE UPDATE OF metadata ON river_job
        FOR EACH ROW
        WHEN (
            OLD.metadata->>'river:partition' IS DISTINCT FROM NEW.metadata->>'river:partition'
        )
        EXECUTE FUNCTION river_job_partition_key_column();
    END IF;
END;
$$;
-- +goose StatementEnd

--
-- workflow V2
--

CREATE TABLE IF NOT EXISTS river_workflow (
    id text PRIMARY KEY,
    name text NULL,
    state text NOT NULL DEFAULT 'active',
    current_attempt integer NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    finalized_at timestamptz NULL,
    wait_eval_cursor_job_id bigint NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS river_workflow_state_created_idx
ON river_workflow (state, created_at, id);

CREATE INDEX IF NOT EXISTS river_workflow_state_finalized_idx
ON river_workflow (state, finalized_at, id)
WHERE finalized_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS river_workflow_unfinalized_id_idx
ON river_workflow (id)
WHERE finalized_at IS NULL;

CREATE INDEX IF NOT EXISTS river_workflow_created_id_idx
ON river_workflow (created_at, id);

CREATE TABLE IF NOT EXISTS river_workflow_signal (
    id bigserial PRIMARY KEY,
    workflow_id text NOT NULL,
    attempt integer NOT NULL,
    key text NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    idempotency_key text NULL,
    source jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS river_workflow_signal_workflow_key_id_idx
ON river_workflow_signal (workflow_id, key, id);

CREATE INDEX IF NOT EXISTS river_workflow_signal_workflow_key_attempt_id_idx
ON river_workflow_signal (workflow_id, key, attempt, id);

CREATE INDEX IF NOT EXISTS river_workflow_signal_workflow_attempt_id_idx
ON river_workflow_signal (workflow_id, attempt, id);

CREATE INDEX IF NOT EXISTS river_workflow_signal_workflow_id_idx
ON river_workflow_signal (workflow_id, id);

CREATE UNIQUE INDEX IF NOT EXISTS river_workflow_signal_workflow_key_idempotency_idx
ON river_workflow_signal (workflow_id, key, idempotency_key)
WHERE idempotency_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS river_workflow_worklist (
    id bigserial PRIMARY KEY,
    workflow_id text NOT NULL,
    reason smallint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS river_workflow_worklist_workflow_id_id_idx
ON river_workflow_worklist (workflow_id, id);

CREATE INDEX IF NOT EXISTS river_workflow_worklist_id_workflow_id_idx
ON river_workflow_worklist (id) INCLUDE (workflow_id);

CREATE TABLE IF NOT EXISTS river_workflow_timer (
    id bigserial PRIMARY KEY,
    workflow_id text NOT NULL UNIQUE,
    next_fire_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS river_workflow_timer_next_fire_at_idx
ON river_workflow_timer (next_fire_at, workflow_id);

CREATE TABLE IF NOT EXISTS river_workflow_attempt (
    workflow_id text NOT NULL,
    attempt integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    triggered_by jsonb NOT NULL DEFAULT '{}'::jsonb,
    retry_mode text NOT NULL DEFAULT 'initial',
    reset_history boolean NOT NULL DEFAULT false,
    PRIMARY KEY (workflow_id, attempt)
);

CREATE TABLE IF NOT EXISTS river_workflow_attempt_task (
    workflow_id text NOT NULL,
    attempt integer NOT NULL,
    task text NOT NULL,
    job_id bigint NOT NULL,
    state text NOT NULL,
    finalized_at timestamptz NULL,
    attempt_count smallint NOT NULL,
    errors jsonb[] NOT NULL DEFAULT '{}'::jsonb[],
    metadata jsonb NOT NULL,
    PRIMARY KEY (workflow_id, attempt, task)
);

CREATE INDEX IF NOT EXISTS river_workflow_attempt_task_workflow_attempt_idx
ON river_workflow_attempt_task (workflow_id, attempt);

ALTER TABLE river_job
    ADD COLUMN IF NOT EXISTS workflow_id text,
    ADD COLUMN IF NOT EXISTS workflow_task text;

-- Populate workflow columns for existing rows. On a fresh database this is a
-- no-op; on an upgrade it backfills from metadata in a single pass. For very
-- large tables a batched approach may be preferable (see TODOS.md).
UPDATE river_job
SET
    workflow_task = metadata->>'task',
    workflow_id = metadata->>'workflow_id'
WHERE (workflow_task IS NULL AND metadata->>'task' IS NOT NULL)
   OR (workflow_id IS NULL AND metadata->>'workflow_id' IS NOT NULL);

-- Keep workflow columns in sync with metadata. Two triggers share one
-- function: one for INSERT, one for UPDATE. Both use WHEN clauses evaluated
-- in C before the plpgsql function is invoked, so the common case
-- (non-workflow inserts and metadata updates that only touch wait state,
-- staging timestamps, etc.) pays zero function-call overhead.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION river_job_workflow_columns()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.workflow_id   := NEW.metadata->>'workflow_id';
    NEW.workflow_task := NEW.metadata->>'task';
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

-- INSERT: fire only for workflow jobs (non-workflow inserts skip entirely).
CREATE OR REPLACE TRIGGER river_job_set_workflow_columns_on_insert
BEFORE INSERT ON river_job
FOR EACH ROW
WHEN (
    NEW.metadata->>'workflow_id' IS NOT NULL
)
EXECUTE FUNCTION river_job_workflow_columns();

-- UPDATE: fire only when the workflow_id or task keys actually change in
-- metadata, which should be never under normal operation but guards against
-- manual edits or future metadata mutations.
CREATE OR REPLACE TRIGGER river_job_set_workflow_columns_on_update
BEFORE UPDATE OF metadata ON river_job
FOR EACH ROW
WHEN (
    OLD.metadata->>'workflow_id' IS DISTINCT FROM NEW.metadata->>'workflow_id'
    OR OLD.metadata->>'task' IS DISTINCT FROM NEW.metadata->>'task'
)
EXECUTE FUNCTION river_job_workflow_columns();

ALTER TABLE river_job_dead_letter
    ADD COLUMN IF NOT EXISTS workflow_id text,
    ADD COLUMN IF NOT EXISTS workflow_task text;

UPDATE river_job_dead_letter
SET
    workflow_task = metadata->>'task',
    workflow_id = metadata->>'workflow_id'
WHERE metadata->>'workflow_id' IS NOT NULL
  AND (
      workflow_task IS DISTINCT FROM metadata->>'task'
      OR workflow_id IS DISTINCT FROM metadata->>'workflow_id'
  );

CREATE INDEX IF NOT EXISTS river_job_non_workflow_queue_state_finalized_idx
ON river_job (queue, state, finalized_at, id)
WHERE finalized_at IS NOT NULL AND workflow_id IS NULL;

CREATE INDEX IF NOT EXISTS river_job_non_workflow_state_finalized_id_idx
ON river_job (state, finalized_at, id)
INCLUDE (queue)
WHERE workflow_id IS NULL
  AND state IN ('cancelled', 'completed', 'discarded');

CREATE INDEX IF NOT EXISTS river_job_workflow_id_id_idx
ON river_job (workflow_id, id)
WHERE workflow_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS river_job_workflow_active_idx
ON river_job (workflow_id)
WHERE workflow_id IS NOT NULL
  AND state IN ('available', 'pending', 'retryable', 'running', 'scheduled');

CREATE INDEX IF NOT EXISTS river_job_workflow_inactive_idx
ON river_job (workflow_id)
WHERE workflow_id IS NOT NULL
  AND state IN ('cancelled', 'completed', 'discarded');

CREATE INDEX IF NOT EXISTS river_job_dead_letter_workflow_id_idx
ON river_job_dead_letter (workflow_id)
WHERE workflow_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS river_job_workflow_state_task_idx
ON river_job (state, workflow_id, workflow_task)
WHERE workflow_id IS NOT NULL AND workflow_task IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS river_job_workflow_id_task_unique_idx
ON river_job (workflow_id, workflow_task)
INCLUDE (state)
WHERE workflow_id IS NOT NULL AND workflow_task IS NOT NULL;

CREATE INDEX IF NOT EXISTS river_job_workflow_pending_id_idx
ON river_job (workflow_id, id)
WHERE state = 'pending' AND workflow_id IS NOT NULL AND workflow_task IS NOT NULL;

CREATE INDEX IF NOT EXISTS river_job_workflow_pending_wait_idx
ON river_job (workflow_id)
WHERE
    state = 'pending' AND
    workflow_id IS NOT NULL AND
    metadata ? 'river:workflow_wait';

CREATE INDEX IF NOT EXISTS river_job_workflow_pending_wait_active_idx
ON river_job (workflow_id, id)
WHERE
    state = 'pending' AND
    workflow_id IS NOT NULL AND
    metadata ? 'river:workflow_wait' AND
    (metadata->'river:workflow_wait_state'->>'started_at') IS NOT NULL AND
    (metadata->'river:workflow_wait_state'->>'resolved_at') IS NULL;

-- Trigger-based delete backstop: enqueues affected workflows for evaluation
-- when workflow jobs are deleted. Uses a trigger instead of a periodic sweep
-- to avoid costly full-table scans of all active workflows. The session flag
-- river.skip_workflow_delete_enqueue allows internal bulk deletes (e.g.
-- ProJobCleaner) to suppress enqueue when they manage worklist entries
-- themselves.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION river_workflow_enqueue_on_job_delete()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF COALESCE(current_setting('river.skip_workflow_delete_enqueue', true), '') = 'on' THEN
        RETURN NULL;
    END IF;

    INSERT INTO river_workflow_worklist (workflow_id, reason)
    SELECT DISTINCT old_rows.workflow_id, 9
    FROM old_rows
    LEFT JOIN river_workflow
        ON river_workflow.id = old_rows.workflow_id
    WHERE old_rows.workflow_id IS NOT NULL
        AND (river_workflow.id IS NULL OR river_workflow.finalized_at IS NULL);

    RETURN NULL;
END;
$$;
-- +goose StatementEnd

CREATE OR REPLACE TRIGGER river_job_enqueue_workflow_on_delete
AFTER DELETE ON river_job
REFERENCING OLD TABLE AS old_rows
FOR EACH STATEMENT
EXECUTE FUNCTION river_workflow_enqueue_on_job_delete();

-- +goose Down

-- River pro migration 006 [down]
--
-- partition key triggers
--

-- No-op: trigger and function cleanup is handled by the 003 down migration.

--
-- workflow V2
--

DROP TRIGGER IF EXISTS river_job_enqueue_workflow_on_delete ON river_job;
DROP FUNCTION IF EXISTS river_workflow_enqueue_on_job_delete();

DROP TRIGGER IF EXISTS river_job_set_workflow_columns_on_insert ON river_job;
DROP TRIGGER IF EXISTS river_job_set_workflow_columns_on_update ON river_job;
DROP FUNCTION IF EXISTS river_job_workflow_columns();

DROP INDEX IF EXISTS river_job_workflow_state_task_idx;
DROP INDEX IF EXISTS river_job_workflow_id_id_idx;
DROP INDEX IF EXISTS river_job_workflow_active_idx;
DROP INDEX IF EXISTS river_job_workflow_inactive_idx;
DROP INDEX IF EXISTS river_job_dead_letter_workflow_id_idx;
DROP INDEX IF EXISTS river_job_workflow_pending_id_idx;
DROP INDEX IF EXISTS river_job_workflow_pending_wait_idx;
DROP INDEX IF EXISTS river_job_workflow_pending_wait_active_idx;
DROP INDEX IF EXISTS river_job_workflow_id_task_unique_idx;

ALTER TABLE river_job
    DROP COLUMN IF EXISTS workflow_task,
    DROP COLUMN IF EXISTS workflow_id;

ALTER TABLE river_job_dead_letter
    DROP COLUMN IF EXISTS workflow_task,
    DROP COLUMN IF EXISTS workflow_id;

DROP INDEX IF EXISTS river_workflow_attempt_task_workflow_attempt_idx;
DROP TABLE IF EXISTS river_workflow_attempt_task;
DROP TABLE IF EXISTS river_workflow_attempt;

DROP INDEX IF EXISTS river_job_non_workflow_queue_state_finalized_idx;
DROP INDEX IF EXISTS river_job_non_workflow_state_finalized_id_idx;

DROP INDEX IF EXISTS river_workflow_timer_next_fire_at_idx;
DROP TABLE IF EXISTS river_workflow_timer;

DROP INDEX IF EXISTS river_workflow_worklist_workflow_id_id_idx;
DROP INDEX IF EXISTS river_workflow_worklist_id_workflow_id_idx;
DROP TABLE IF EXISTS river_workflow_worklist;

DROP INDEX IF EXISTS river_workflow_signal_workflow_key_idempotency_idx;
DROP INDEX IF EXISTS river_workflow_signal_workflow_id_idx;
DROP INDEX IF EXISTS river_workflow_signal_workflow_attempt_id_idx;
DROP INDEX IF EXISTS river_workflow_signal_workflow_key_attempt_id_idx;
DROP INDEX IF EXISTS river_workflow_signal_workflow_key_id_idx;
DROP TABLE IF EXISTS river_workflow_signal;

DROP INDEX IF EXISTS river_workflow_created_id_idx;
DROP INDEX IF EXISTS river_workflow_state_finalized_idx;
DROP INDEX IF EXISTS river_workflow_state_created_idx;
DROP INDEX IF EXISTS river_workflow_unfinalized_id_idx;
DROP TABLE IF EXISTS river_workflow;
