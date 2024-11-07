package http

import (
	"context"

	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/nc"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.NaisConsoleAPI = &naisConsoleAPI{}

type naisConsoleAPI struct {
	fetcher nc.Fetcher
}

func (a *naisConsoleAPI) GetGoogleProjectsForAllTeams(ctx context.Context) ([]*service.NaisTeamMapping, error) {
	const op errs.Op = "naisConsoleAPI.GetGoogleProjectsForAllTeams"

	projects, err := a.fetcher.GetTeamGoogleProjects(ctx)
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeNaisConsole, op, err)
	}

	out := []*service.NaisTeamMapping{}
	for team, teamProps := range projects {
		out = append(out, &service.NaisTeamMapping{
			Slug:       team,
			GroupEmail: teamProps.GroupEmail,
			ProjectID:  teamProps.GCPProjectID,
		})
	}

	return out, nil
}

func NewNaisConsoleAPI(fetcher nc.Fetcher) *naisConsoleAPI {
	return &naisConsoleAPI{
		fetcher: fetcher,
	}
}
