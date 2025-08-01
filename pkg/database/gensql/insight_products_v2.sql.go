// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: insight_products_v2.sql

package gensql

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const getInsightProductWithTeamkatalogen = `-- name: GetInsightProductWithTeamkatalogen :one
SELECT
    id, name, description, creator, created, last_modified, type, tsv_document, link, keywords, "group", teamkatalogen_url, team_id, team_name, pa_name
FROM
    insight_product_with_teamkatalogen_view
WHERE
    "id" = $1
ORDER BY
    last_modified DESC
`

func (q *Queries) GetInsightProductWithTeamkatalogen(ctx context.Context, id uuid.UUID) (InsightProductWithTeamkatalogenView, error) {
	row := q.db.QueryRowContext(ctx, getInsightProductWithTeamkatalogen, id)
	var i InsightProductWithTeamkatalogenView
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Creator,
		&i.Created,
		&i.LastModified,
		&i.Type,
		&i.TsvDocument,
		&i.Link,
		pq.Array(&i.Keywords),
		&i.Group,
		&i.TeamkatalogenUrl,
		&i.TeamID,
		&i.TeamName,
		&i.PaName,
	)
	return i, err
}

const getInsightProductsByGroups = `-- name: GetInsightProductsByGroups :many
SELECT
    id, name, description, creator, created, last_modified, type, tsv_document, link, keywords, "group", teamkatalogen_url, team_id, team_name, pa_name
FROM
    insight_product_with_teamkatalogen_view ipwtv
WHERE
    "group" = ANY($1::text[])
ORDER BY
    ipwtv.team_name, ipwtv.name ASC
`

func (q *Queries) GetInsightProductsByGroups(ctx context.Context, groups []string) ([]InsightProductWithTeamkatalogenView, error) {
	rows, err := q.db.QueryContext(ctx, getInsightProductsByGroups, pq.Array(groups))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []InsightProductWithTeamkatalogenView{}
	for rows.Next() {
		var i InsightProductWithTeamkatalogenView
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Creator,
			&i.Created,
			&i.LastModified,
			&i.Type,
			&i.TsvDocument,
			&i.Link,
			pq.Array(&i.Keywords),
			&i.Group,
			&i.TeamkatalogenUrl,
			&i.TeamID,
			&i.TeamName,
			&i.PaName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getInsightProductsByProductArea = `-- name: GetInsightProductsByProductArea :many
SELECT
    id, name, description, creator, created, last_modified, type, tsv_document, link, keywords, "group", teamkatalogen_url, team_id, team_name, pa_name
FROM
    insight_product_with_teamkatalogen_view
WHERE
    team_id = ANY($1::uuid[])
ORDER BY
    last_modified DESC
`

func (q *Queries) GetInsightProductsByProductArea(ctx context.Context, teamID []uuid.UUID) ([]InsightProductWithTeamkatalogenView, error) {
	rows, err := q.db.QueryContext(ctx, getInsightProductsByProductArea, pq.Array(teamID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []InsightProductWithTeamkatalogenView{}
	for rows.Next() {
		var i InsightProductWithTeamkatalogenView
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Creator,
			&i.Created,
			&i.LastModified,
			&i.Type,
			&i.TsvDocument,
			&i.Link,
			pq.Array(&i.Keywords),
			&i.Group,
			&i.TeamkatalogenUrl,
			&i.TeamID,
			&i.TeamName,
			&i.PaName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
