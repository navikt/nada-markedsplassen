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
