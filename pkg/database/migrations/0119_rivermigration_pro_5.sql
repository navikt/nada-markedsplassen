-- +goose Up

-- River pro migration 005 [up]
CREATE INDEX IF NOT EXISTS river_job_uncoordinated_batching_index
ON river_job (queue, kind, (metadata->>'river:batch_key'), priority, scheduled_at, id)
INCLUDE (state)
WHERE metadata ? 'river:batch_key'
    AND metadata->>'river:batch_mode' = 'uncoordinated'
    AND state = 'available';

-- +goose Down

-- River pro migration 005 [down]
DROP INDEX river_job_uncoordinated_batching_index;
