// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: sessions.sql

package gensql

import (
	"context"
	"time"
)

const createSession = `-- name: CreateSession :exec
INSERT INTO sessions (
	"token",
	"access_token",
	"email",
	"name",
	"expires"
) VALUES (
	$1,
	$2,
	LOWER($3),
	$4,
	$5
)
`

type CreateSessionParams struct {
	Token       string
	AccessToken string
	Email       string
	Name        string
	Expires     time.Time
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) error {
	_, err := q.db.ExecContext(ctx, createSession,
		arg.Token,
		arg.AccessToken,
		arg.Email,
		arg.Name,
		arg.Expires,
	)
	return err
}

const deleteSession = `-- name: DeleteSession :exec
DELETE
FROM sessions
WHERE token = $1
`

func (q *Queries) DeleteSession(ctx context.Context, token string) error {
	_, err := q.db.ExecContext(ctx, deleteSession, token)
	return err
}

const getSession = `-- name: GetSession :one
SELECT token, access_token, email, name, created, expires
FROM sessions
WHERE token = $1
AND expires > now()
`

func (q *Queries) GetSession(ctx context.Context, token string) (Session, error) {
	row := q.db.QueryRowContext(ctx, getSession, token)
	var i Session
	err := row.Scan(
		&i.Token,
		&i.AccessToken,
		&i.Email,
		&i.Name,
		&i.Created,
		&i.Expires,
	)
	return i, err
}
