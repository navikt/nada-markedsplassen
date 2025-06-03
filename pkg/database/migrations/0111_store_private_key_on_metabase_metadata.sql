-- +goose Up
ALTER TABLE metabase_metadata
    ADD COLUMN sa_private_key bytea DEFAULT NULL;

-- +goose Down
ALTER TABLE metabase_metadata
    DROP COLUMN sa_private_key;
