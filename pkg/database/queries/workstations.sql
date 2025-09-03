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
    (SELECT COALESCE(disable_global_allow_list, FALSE) FROM workstations_url_list_user_settings WHERE nav_ident = @nav_ident)
);

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

-- name: UpdateWorkstationURLListItemsExpiresAtForIdent :exec
UPDATE workstations_url_lists
SET expires_at = (NOW() + duration)
WHERE id = ANY(@id::uuid[]);

-- name: GetWorkstationActiveURLListForIdent :one
SELECT
    w.nav_ident,
    array_agg(w.url ORDER BY w.created_at DESC)::text[] AS url_list_items,
    h.disable_global_url_list
FROM workstations_url_lists w
JOIN (
    SELECT 
        uh.nav_ident, 
        disable_global_url_list
    FROM workstations_url_list_history uh
    WHERE uh.nav_ident = @nav_ident
    ORDER BY created_at DESC
    LIMIT 1
) h ON w.nav_ident = h.nav_ident
WHERE w.expires_at > NOW() AND w.nav_ident = @nav_ident
GROUP BY w.nav_ident, h.disable_global_url_list;

-- name: UpdateWorkstationURLListUserSettings :one
INSERT INTO workstations_url_list_user_settings (nav_ident, disable_global_allow_list)
VALUES (@nav_ident, @disable_global_allow_list)
ON CONFLICT (nav_ident) DO UPDATE
SET disable_global_allow_list = EXCLUDED.disable_global_allow_list
RETURNING *;

-- name: GetWorkstationURLListUserSettings :one
SELECT
    *
FROM workstations_url_list_user_settings
WHERE nav_ident = @nav_ident;

-- name: GetLatestWorkstationURLListHistoryEntry :one
SELECT
    *
FROM workstations_url_list_history
WHERE nav_ident = @nav_ident
ORDER BY created_at DESC
LIMIT 1;

-- name: GetWorkstationURLListUsers :many
SELECT
    DISTINCT nav_ident
FROM workstations_url_list_user_settings;

