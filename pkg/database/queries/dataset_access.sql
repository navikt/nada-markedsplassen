-- name: GrantAccessToDataset :exec
INSERT INTO dataset_access (dataset_id,
                            "subject",
                            "owner",
                            granter,
                            expires,
                            access_request_id)
VALUES (@dataset_id,
        @subject,
        @owner,
        LOWER(@granter),
        @expires,
        @access_request_id);

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
);
