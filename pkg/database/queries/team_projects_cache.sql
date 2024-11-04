-- name: AddTeamProject :one
INSERT INTO team_projects ("team",
                           "group_email",
                           "project")
VALUES (
    @team,
    @group_email,
    @project
)
RETURNING *;

-- name: GetTeamProjects :many
SELECT *
FROM team_projects;

-- name: GetTeamProject :one
SELECT *
FROM team_projects
WHERE group_email = @group_email;

-- name: ClearTeamProjectsCache :exec
TRUNCATE team_projects;
