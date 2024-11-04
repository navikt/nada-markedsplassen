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

func (a *naisConsoleAPI) GetGoogleProjectsForAllTeams(ctx context.Context) (map[string]service.NaisTeamMapping, error) {
	const op errs.Op = "naisConsoleAPI.GetGoogleProjectsForAllTeams"

	projects, err := a.fetcher.GetTeamGoogleProjects(ctx)
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}

	out := make(map[string]service.NaisTeamMapping)
	for team, teamProps := range projects {
		out[team] = service.NaisTeamMapping{
			GroupEmail: teamProps.GroupEmail,
			ProjectID:  teamProps.GCPProjectID,
		}
	}

	return out, nil
}

func NewNaisConsoleAPI(fetcher nc.Fetcher) *naisConsoleAPI {
	return &naisConsoleAPI{
		fetcher: fetcher,
	}
}
