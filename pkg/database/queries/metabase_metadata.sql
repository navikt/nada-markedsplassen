-- name: CreateRestrictedMetabaseMetadata :exec
INSERT INTO restricted_metabase_metadata (
    "dataset_id"
) VALUES (
    @dataset_id
);

-- name: SoftDeleteRestrictedMetabaseMetadata :exec
UPDATE restricted_metabase_metadata
SET "deleted_at" = NOW()
WHERE dataset_id = @dataset_id;

-- name: RestoreRestrictedMetabaseMetadata :exec
UPDATE restricted_metabase_metadata
SET "deleted_at" = null
WHERE dataset_id = @dataset_id;

-- name: SetPermissionGroupRestrictedMetabaseMetadata :one
UPDATE restricted_metabase_metadata
SET "permission_group_id" = @permission_group_id
WHERE dataset_id = @dataset_id
RETURNING *;

-- name: SetCollectionRestrictedMetabaseMetadata :one
UPDATE restricted_metabase_metadata
SET "collection_id" = @collection_id
WHERE dataset_id = @dataset_id
RETURNING *;

-- name: SetDatabaseRestrictedMetabaseMetadata :one
UPDATE restricted_metabase_metadata
SET "database_id" = @database_id
WHERE dataset_id = @dataset_id
RETURNING *;

-- name: SetServiceAccountRestrictedMetabaseMetadata :one
UPDATE restricted_metabase_metadata
SET "sa_email" = @sa_email
WHERE dataset_id = @dataset_id
RETURNING *;

-- name: SetServiceAccountPrivateKeyRestrictedMetabaseMetadata :one
UPDATE restricted_metabase_metadata
SET "sa_private_key" = @sa_private_key
WHERE dataset_id = @dataset_id
RETURNING *;

-- name: SetSyncCompletedRestrictedMetabaseMetadata :exec
UPDATE restricted_metabase_metadata
SET "sync_completed" = NOW()
WHERE dataset_id = @dataset_id;

-- name: GetRestrictedMetabaseMetadata :one
SELECT *
FROM restricted_metabase_metadata
WHERE "dataset_id" = @dataset_id AND "deleted_at" IS NULL;

-- name: GetAllRestrictedMetabaseMetadata :many
SELECT *
FROM restricted_metabase_metadata;

-- name: GetRestrictedMetabaseMetadataWithDeleted :one
SELECT *
FROM restricted_metabase_metadata
WHERE "dataset_id" = @dataset_id;

-- name: DeleteRestrictedMetabaseMetadata :exec
DELETE 
FROM restricted_metabase_metadata
WHERE "dataset_id" = @dataset_id;

-- name: GetOpenMetabaseTablesInSameBigQueryDataset :many
WITH sources_in_same_dataset AS (
  SELECT * FROM datasource_bigquery 
  WHERE project_id = @project_id AND dataset = @dataset
)

SELECT table_name FROM sources_in_same_dataset sds
JOIN restricted_metabase_metadata mbm
ON mbm.dataset_id = sds.dataset_id
WHERE mbm.permission_group_id = 0;
