-- name: GetWorkstationsJob :one
SELECT
    *
FROM
    workstations_jobs
WHERE
    user_ident = @user_ident;

-- name: CreateWorkstationsJob :exec
INSERT INTO workstations_jobs ("user_ident", "job_id")
VALUES (@user_ident, @job_id);

-- name: DeleteWorkstationsJob :exec
DELETE FROM workstations_jobs
WHERE user_ident = @user_ident;
