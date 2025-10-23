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
    teamkatalogen_url,
    team_id
) 
VALUES(
    @name,
    @description,
    @owner_group,
    @public_dashboard_id,
    @metabase_id,
    @created_by,
    @keywords,
    @teamkatalogen_url,
    @team_id
) RETURNING *;

-- name: DeletePublicDashboard :exec
DELETE FROM metabase_dashboard
WHERE id = @id;

-- name: GetPublicDashboard :one
SELECT *
FROM metabase_dashboard
WHERE id = @id;

-- name: GetPublicDashboardsForGroups :many
SELECT *
FROM metabase_dashboard
WHERE "group" = ANY(@groups::text[]);

-- name: UpdatePublicDashboard :one
UPDATE metabase_dashboard
SET
    name = @name,
    description = @description,
    keywords = @keywords,
    teamkatalogen_url = @teamkatalogen_url,
    team_id = @team_id
WHERE id = @id
RETURNING *;
