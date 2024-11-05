-- name: GetNadaTokenFromGroupEmail :one
SELECT
    token
FROM
    nada_tokens nt
JOIN 
    team_projects tp
ON
    nt.team = tp.team
WHERE
    tp.group_email = @group_email;

-- name: GetNadaTokens :many
SELECT 
    *
FROM 
    nada_tokens;

-- name: GetNadaTokensForTeams :many
SELECT
    *
FROM
    nada_tokens
WHERE
    team = ANY (@teams :: text [])
ORDER BY
    team;

-- name: GetTeamEmailFromNadaToken :one
SELECT group_email
FROM team_projects tp
JOIN nada_tokens nt
ON tp.team = nt.team
WHERE nt.token = @token;

-- name: RotateNadaToken :exec
UPDATE nada_tokens
SET token = gen_random_uuid()
WHERE team = @team;

-- name: DeleteNadaToken :exec
DELETE FROM
    nada_tokens
WHERE
    team = @team;