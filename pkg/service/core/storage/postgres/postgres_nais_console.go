package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/database/gensql"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.NaisConsoleStorage = &naisConsoleStorage{}

type naisConsoleStorage struct {
	db *database.Repo
}

func (s *naisConsoleStorage) GetAllTeamProjects(ctx context.Context) ([]*service.NaisTeamMapping, error) {
	const op errs.Op = "naisConsoleStorage.GetAllTeamProjects"

	raw, err := s.db.Querier.GetTeamProjects(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, op, err)
	}

	teamProjects := []*service.NaisTeamMapping{}
	for _, team := range raw {
		teamProjects = append(teamProjects, &service.NaisTeamMapping{
			Slug:       team.Team,
			GroupEmail: team.GroupEmail,
			ProjectID:  team.Project,
		})
	}

	return teamProjects, nil
}

func (s *naisConsoleStorage) UpdateAllTeamProjects(ctx context.Context, teams []*service.NaisTeamMapping) error {
	const op errs.Op = "naisConsoleStorage.UpdateAllTeamProjects"

	tx, err := s.db.GetDB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := s.db.Querier.WithTx(tx)

	err = q.ClearTeamProjectsCache(ctx)
	if err != nil {
		return errs.E(errs.Database, op, err)
	}

	for _, team := range teams {
		_, err := q.AddTeamProject(ctx, gensql.AddTeamProjectParams{
			Team:       team.Slug,
			Project:    team.ProjectID,
			GroupEmail: team.GroupEmail,
		})
		if err != nil {
			return errs.E(errs.Database, op, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return errs.E(errs.Database, op, err)
	}

	return nil
}

func (s *naisConsoleStorage) GetTeamProject(ctx context.Context, googleEmail string) (*service.NaisTeamMapping, error) {
	const op errs.Op = "naisConsoleStorage.GetTeamProject"

	raw, err := s.db.Querier.GetTeamProjectFromGroupEmail(ctx, googleEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.E(errs.NotExist, op, errs.Parameter("googleEmail"), errs.UserName(googleEmail), err)
		}

		return nil, errs.E(errs.Database, op, err)
	}

	return &service.NaisTeamMapping{
		Slug:       raw.Team,
		GroupEmail: raw.GroupEmail,
		ProjectID:  raw.Project,
	}, nil
}

func NewNaisConsoleStorage(db *database.Repo) *naisConsoleStorage {
	return &naisConsoleStorage{
		db: db,
	}
}
