-- +goose Up
CREATE TABLE workstations_jobs
(
    "user_ident" TEXT NOT NULL UNIQUE,
    "job_id"     BIGINT NOT NULL
);

-- +goose Down

DROP TABLE workstations_jobs;
