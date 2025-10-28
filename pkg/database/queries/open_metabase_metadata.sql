-- name: CreateOpenMetabaseMetadata :exec
INSERT INTO open_metabase_metadata (
    "dataset_id"
) VALUES (
    @dataset_id
);

-- name: SoftDeleteOpenMetabaseMetadata :exec
UPDATE open_metabase_metadata
SET "deleted_at" = NOW()
WHERE dataset_id = @dataset_id;

-- name: RestoreOpenMetabaseMetadata :exec
UPDATE open_metabase_metadata
SET "deleted_at" = null
WHERE dataset_id = @dataset_id;

-- name: SetDatabaseOpenMetabaseMetadata :one
UPDATE open_metabase_metadata
SET "database_id" = @database_id
WHERE dataset_id = @dataset_id
RETURNING *;

-- name: SetSyncCompletedOpenMetabaseMetadata :exec
UPDATE open_metabase_metadata
SET "sync_completed" = NOW()
WHERE dataset_id = @dataset_id;

-- name: GetOpenMetabaseMetadata :one
SELECT *
FROM open_metabase_metadata
WHERE "dataset_id" = @dataset_id AND "deleted_at" IS NULL;

-- name: GetAllOpenMetabaseMetadata :many
SELECT *
FROM open_metabase_metadata;

-- name: GetOpenMetabaseMetadataWithDeleted :one
SELECT *
FROM open_metabase_metadata
WHERE "dataset_id" = @dataset_id;

-- name: DeleteOpenMetabaseMetadata :exec
DELETE 
FROM open_metabase_metadata
WHERE "dataset_id" = @dataset_id;

-- name: GetOpenMetabaseTablesInSameBigQueryDataset2 :many
WITH sources_in_same_dataset AS (
  SELECT * FROM datasource_bigquery 
  WHERE project_id = @project_id AND dataset = @dataset
)

SELECT table_name FROM sources_in_same_dataset sds
JOIN open_metabase_metadata mbm
ON mbm.dataset_id = sds.dataset_id
WHERE mbm.permission_group_id = 0;
