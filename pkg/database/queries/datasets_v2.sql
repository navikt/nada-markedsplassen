-- name: GetAllDatasetsMinimal :many
SELECT ds.id, ds.created, name, project_id, dataset, table_name 
FROM datasets ds 
JOIN datasource_bigquery dsb 
ON ds.id = dsb.dataset_id;

-- name: GetDatasetComplete :many
SELECT
  *
FROM
  dataset_view
WHERE
  ds_id = @id;

-- name: GetDatasetCompleteWithAccess :many
SELECT
  *
FROM
  dataset_view dv
LEFT JOIN dataset_access_view da ON da.access_dataset_id = dv.ds_id 
    AND (
        dp_group = ANY(@groups::TEXT[])
        OR (
            SPLIT_PART(da.access_subject, ':', 2) = ANY(@groups::TEXT[])
            AND da.access_revoked IS NULL
        )
        OR (
            da.access_revoked IS NULL
            AND SPLIT_PART(da.access_subject, ':', 1) = 'serviceAccount'
            AND da.access_owner = ANY(@groups::TEXT[])
        )
    )
WHERE ds_id = @id;

