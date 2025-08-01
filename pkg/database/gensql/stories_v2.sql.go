// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: stories_v2.sql

package gensql

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const getStoriesWithTeamkatalogenByGroups = `-- name: GetStoriesWithTeamkatalogenByGroups :many
SELECT id, name, creator, created, last_modified, description, keywords, teamkatalogen_url, team_id, "group", team_name, pa_name
FROM story_with_teamkatalogen_view swtv
WHERE "group" = ANY ($1::text[])
ORDER BY swtv."team_name", swtv.name ASC
`

func (q *Queries) GetStoriesWithTeamkatalogenByGroups(ctx context.Context, groups []string) ([]StoryWithTeamkatalogenView, error) {
	rows, err := q.db.QueryContext(ctx, getStoriesWithTeamkatalogenByGroups, pq.Array(groups))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []StoryWithTeamkatalogenView{}
	for rows.Next() {
		var i StoryWithTeamkatalogenView
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Creator,
			&i.Created,
			&i.LastModified,
			&i.Description,
			pq.Array(&i.Keywords),
			&i.TeamkatalogenUrl,
			&i.TeamID,
			&i.Group,
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

const getStoriesWithTeamkatalogenByIDs = `-- name: GetStoriesWithTeamkatalogenByIDs :many
SELECT id, name, creator, created, last_modified, description, keywords, teamkatalogen_url, team_id, "group", team_name, pa_name
FROM story_with_teamkatalogen_view
WHERE id = ANY ($1::uuid[])
ORDER BY last_modified DESC
`

func (q *Queries) GetStoriesWithTeamkatalogenByIDs(ctx context.Context, ids []uuid.UUID) ([]StoryWithTeamkatalogenView, error) {
	rows, err := q.db.QueryContext(ctx, getStoriesWithTeamkatalogenByIDs, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []StoryWithTeamkatalogenView{}
	for rows.Next() {
		var i StoryWithTeamkatalogenView
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Creator,
			&i.Created,
			&i.LastModified,
			&i.Description,
			pq.Array(&i.Keywords),
			&i.TeamkatalogenUrl,
			&i.TeamID,
			&i.Group,
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
