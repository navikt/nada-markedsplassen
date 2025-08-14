-- name: CreateWorkstationsConfigChange :exec
INSERT INTO workstations_config_history (
    "nav_ident", 
    "workstation_config"
)
VALUES (
    @nav_ident,
    @workstation_config
);

-- name: CreateWorkstationsOnpremAllowlistChange :exec
INSERT INTO workstations_onprem_allowlist_history (
    "nav_ident", 
    "hosts"
)
VALUES (
    @nav_ident,
    @hosts
);

-- name: GetLastWorkstationsOnpremAllowlistChange :one
SELECT 
    * 
FROM workstations_onprem_allowlist_history
WHERE nav_ident = @nav_ident
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateWorkstationsURLListChange :exec
INSERT INTO workstations_url_list_history (
    "nav_ident", 
    "url_list",
    "disable_global_url_list"
)
VALUES (
    @nav_ident,
    @url_list,
    @disable_global_url_list
);

-- name: GetLastWorkstationsURLListChange :one
SELECT
    *
FROM workstations_url_list_history
WHERE nav_ident = @nav_ident
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateWorkstationsActivityHistory :exec
INSERT INTO workstations_activity_history (
    "nav_ident",
    "action",
    "instance_id"
)
VALUES (
    @nav_ident,
    @action,
    @instance_id
);

-- name: GetWorkstationURLListForIdent :many
SELECT
    *
FROM workstations_url_lists
WHERE nav_ident = @nav_ident
ORDER BY created_at DESC;

-- name: CreateWorkstationURLListItemForIdent :one
INSERT INTO workstations_url_lists (nav_ident, url, description, duration)
    VALUES (@nav_ident, @url, @description, @duration)
RETURNING *;

-- name: UpdateWorkstationURLListItemForIdent :one
UPDATE workstations_url_lists
SET url = @url, description = @description, duration = @duration
WHERE id = @id
RETURNING *;

-- name: DeleteWorkstationURLListItemForIdent :exec
DELETE FROM workstations_url_lists
WHERE id = @id;
