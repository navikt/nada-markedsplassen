-- name: GrantAccessToDataset :exec
INSERT INTO dataset_access (dataset_id,
                            "subject",
                            "owner",
                            granter,
                            expires,
                            access_request_id,
                            platform)
VALUES (@dataset_id,
        @subject,
        @owner,
        LOWER(@granter),
        @expires,
        @access_request_id,
        @platform);

-- name: RevokeAccessToDataset :exec
UPDATE dataset_access
SET revoked = NOW()
WHERE id = @id;

-- name: ListUnrevokedExpiredAccessEntries :many
SELECT *
FROM dataset_access_view
WHERE access_revoked IS NULL
  AND access_expires < NOW();

-- name: ListAccessToDataset :many
SELECT *
FROM dataset_access_view
WHERE access_dataset_id = @dataset_id;

-- name: GetAccessToDataset :one
SELECT *
FROM dataset_access_view
WHERE access_id = @id;

-- name: ListActiveAccessToDataset :many
SELECT *
FROM dataset_access_view
WHERE access_dataset_id = @dataset_id AND access_revoked IS NULL AND (access_expires IS NULL OR access_expires >= NOW());

-- name: GetActiveAccessToDatasetForSubject :one
SELECT *
FROM dataset_access_view
WHERE access_dataset_id = @dataset_id 
AND access_subject = @subject 
AND access_revoked IS NULL 
AND (
  access_expires IS NULL 
  OR access_expires >= NOW()
)
AND access_platform = @platform;


-- name: GetUserAccesses :many
SELECT dsa.*,
    dp.id AS dataproduct_id,
    dp.name as dataproduct_name,
    dp.description as dataproduct_description,
    dp.slug as dataproduct_slug,
    dp.group as dataproduct_group,
    ds.id as dataset_id,
    ds.name as dataset_name,
    ds.description as dataset_description,
    ds.slug as dataset_slug
FROM dataset_access_view dsa 
    JOIN datasets ds on dsa.access_dataset_id = ds.id
    JOIN dataproducts dp on ds.dataproduct_id = dp.id
WHERE dsa.access_subject = ANY(@subjects::TEXT[]) OR dsa.access_owner = ANY(@owners::TEXT[])
  AND dsa.access_revoked IS NULL
  AND (dsa.access_expires > NOW() OR dsa.access_expires IS NULL)
ORDER BY
    dsa.access_created DESC;
