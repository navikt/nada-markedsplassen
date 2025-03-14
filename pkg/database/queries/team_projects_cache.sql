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

-- name: GetTeamProjectFromGroupEmail :one
SELECT *
FROM team_projects
WHERE group_email = @group_email;

-- name: GetGroupEmailFromTeamSlug :one
SELECT group_email
FROM team_projects
WHERE team = @team;

-- name: ClearTeamProjectsCache :exec
TRUNCATE team_projects;
