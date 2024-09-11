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
WHERE revoked IS NULL
  AND expires < NOW();

-- name: ListAccessToDataset :many
SELECT *
FROM dataset_access_view
WHERE dataset_id = @dataset_id;

-- name: GetAccessToDataset :one
SELECT *
FROM dataset_access_view
WHERE id = @id;

-- name: ListActiveAccessToDataset :many
SELECT *
FROM dataset_access_view
WHERE dataset_id = @dataset_id AND revoked IS NULL AND (expires IS NULL OR expires >= NOW());

-- name: GetActiveAccessToDatasetForSubject :one
SELECT *
FROM dataset_access_view
WHERE dataset_id = @dataset_id 
AND "subject" = @subject 
AND revoked IS NULL 
AND (
  expires IS NULL 
  OR expires >= NOW()
);
