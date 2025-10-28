-- +goose Up
CREATE TABLE open_metabase_metadata (
    "dataset_id" UUID PRIMARY KEY,
    "database_id" INT NOT NULL,
    "deleted_at" TIMESTAMPTZ,
    "sync_completed" TIMESTAMPTZ,
    CONSTRAINT fk_open_metabase_metadata
    FOREIGN KEY (dataset_id)
        REFERENCES datasets (id) ON DELETE CASCADE);

INSERT INTO open_metabase_metadata
(SELECT dataset_id, database_id, sa_email, deleted_at, sync_completed
FROM metabase_metadata WHERE permission_group_id = 0);

DELETE FROM metabase_metadata WHERE permission_group_id = 0;

ALTER TABLE metabase_metadata RENAME TO restricted_metabase_metadata;

-- +goose Down

INSERT INTO restricted_metabase_metadata
(SELECT dataset_id, database_id, sa_email, deleted_at, sync_completed
FROM open_metabase_metadata);

UPDATE restricted_metabase_metadata
SET permission_group_id = 0
WHERE permission_group_id is null;

DROP TABLE open_metabase_metadata;

ALTER TABLE restricted_metabase_metadata RENAME TO metabase_metadata;
