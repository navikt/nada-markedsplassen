package postgres

import (
	"context"
	"fmt"

	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/database/gensql"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.NaisConsoleStorage = &naisConsoleStorage{}

type naisConsoleStorage struct {
	db *database.Repo
}

func (s *naisConsoleStorage) GetAllTeamProjects(ctx context.Context) (map[string]service.NaisTeamMapping, error) {
	const op errs.Op = "naisConsoleStorage.GetAllTeamProjects"

	raw, err := s.db.Querier.GetTeamProjects(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, op, err)
	}

	teamProjects := map[string]service.NaisTeamMapping{}
	for _, teamProject := range raw {
		teamProjects[teamProject.GroupEmail] = service.NaisTeamMapping{
			Slug:       teamProject.Team,
			GroupEmail: teamProject.GroupEmail,
			ProjectID:  teamProject.Project,
		}
	}

	return teamProjects, nil
}

func (s *naisConsoleStorage) UpdateAllTeamProjects(ctx context.Context, teams map[string]service.NaisTeamMapping) error {
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

	for team, teamProps := range teams {
		_, err := q.AddTeamProject(ctx, gensql.AddTeamProjectParams{
			Team:       team,
			Project:    teamProps.ProjectID,
			GroupEmail: teamProps.GroupEmail,
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

	teamProjects, err := s.GetAllTeamProjects(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	project, ok := teamProjects[googleEmail]
	if !ok {
		return nil, errs.E(errs.NotExist, op, errs.Parameter("naisTeam"), fmt.Errorf("team %s not found", googleEmail))
	}

	return &project, nil
}

func NewNaisConsoleStorage(db *database.Repo) *naisConsoleStorage {
	return &naisConsoleStorage{
		db: db,
	}
}
