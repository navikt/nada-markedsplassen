-- +goose NO TRANSACTION

-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS workstations_activity_history_action_created_at_idx
    ON workstations_activity_history (action, created_at);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS workstations_activity_history_action_created_at_idx;
