-- +goose Up

-- River pro migration 003 [up]
CREATE INDEX IF NOT EXISTS river_job_sequence_latest_active
    ON river_job ((metadata->>'seq_key'), id DESC)
    INCLUDE (state)
    WHERE metadata ? 'seq_key'
        AND state NOT IN ('pending','completed');

CREATE INDEX IF NOT EXISTS river_job_sequence_first_pending
    ON river_job ((metadata->>'seq_key'), id)
    INCLUDE (scheduled_at)
    WHERE metadata ? 'seq_key'
        AND state = 'pending';

ALTER TABLE river_job
    ADD COLUMN IF NOT EXISTS partition_key text
        GENERATED ALWAYS AS (
            CASE WHEN metadata ? 'river:partition' THEN metadata->>'river:partition' ELSE NULL END
            ) STORED;

CREATE INDEX IF NOT EXISTS river_job_partition_key_index
    ON river_job(queue, partition_key, priority, scheduled_at, id)
    WHERE state = 'available'
        AND partition_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS river_job_partition_key_null_index
    ON river_job (queue, priority, scheduled_at, id)
    WHERE state='available'
        AND partition_key IS NULL;

CREATE TABLE IF NOT EXISTS river_periodic_job (
                                                  id text PRIMARY KEY,
                                                  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                                  next_run_at timestamptz NOT NULL,
                                                  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                                  CONSTRAINT id_length CHECK (char_length(id) > 0 AND char_length(id) < 128)
);

CREATE INDEX IF NOT EXISTS river_periodic_job_updated_at_id_index
    ON river_periodic_job (updated_at, id);

-- +goose Down

-- River pro migration 003 [down]
DROP INDEX river_job_partition_key_null_index;
DROP INDEX river_job_partition_key_index;
ALTER TABLE river_job DROP COLUMN partition_key;
DROP INDEX river_job_sequence_first_pending;
DROP INDEX river_job_sequence_latest_active;

DROP TABLE river_periodic_job;
