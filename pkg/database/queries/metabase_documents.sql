-- name: CreatePublicDocument :one
INSERT INTO metabase_document (
    name,
    description,
    "group",
    public_document_id,
    metabase_id,
    created_by,
    keywords,
    teamkatalogen_url,
    team_id
)
VALUES (
    sqlc.arg(name),
    sqlc.narg(description),
    @owner_group,
    sqlc.arg(public_document_id),
    sqlc.arg(metabase_id),
    sqlc.arg(created_by),
    sqlc.arg(keywords),
    sqlc.narg(teamkatalogen_url),
    sqlc.narg(team_id)
)
RETURNING *;

-- name: GetPublicDocument :one
SELECT *
FROM metabase_document
WHERE id = sqlc.arg(id);

-- name: GetPublicDocumentsForGroups :many
SELECT *
FROM metabase_document
WHERE "group" = ANY(sqlc.arg(groups)::text[]);

-- name: UpdatePublicDocument :one
UPDATE metabase_document
SET name = sqlc.arg(name),
    description = sqlc.narg(description),
    keywords = sqlc.arg(keywords),
    teamkatalogen_url = sqlc.narg(teamkatalogen_url),
    team_id = sqlc.narg(team_id)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeletePublicDocument :exec
DELETE FROM metabase_document
WHERE id = sqlc.arg(id);
