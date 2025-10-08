-- name: GetDashboard :one
SELECT *
FROM "dashboards"
WHERE id = @id;

-- name: CreatePublicDashboard :one
INSERT INTO metabase_dashboard (
    name,
    description,
    "group",
    public_dashboard_id,
    metabase_id,
    created_by,
    keywords,
    teamkatalogen_url
) 
VALUES(
    @name,
    @description,
    @owner_group,
    @public_dashboard_id,
    @metabase_id,
    @created_by,
    @keywords,
    @teamkatalogen_url
) RETURNING *;

-- name: DeletePublicDashboard :exec
DELETE FROM metabase_dashboard
WHERE id = @id;

