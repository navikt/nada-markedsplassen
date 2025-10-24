-- +goose Up
UPDATE metabase_metadata SET permission_group_id = 1 WHERE permission_group_id = 0 OR permission_group_id IS NULL;

-- +goose Down