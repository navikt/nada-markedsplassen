package teamkatalogen

import (
	"context"
	"fmt"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/syncers"
	"github.com/rs/zerolog"
)

var _ syncers.Runner = &Runner{}

type Runner struct {
	api     service.TeamKatalogenAPI
	storage service.ProductAreaStorage
}

func New(api service.TeamKatalogenAPI, storage service.ProductAreaStorage) *Runner {
	tk := &Runner{
		api:     api,
		storage: storage,
	}

	return tk
}

func (r *Runner) Name() string {
	return "TeamKatalogenSyncer"
}

func (r *Runner) RunOnce(ctx context.Context, log zerolog.Logger) error {
	pas, err := r.api.GetProductAreas(ctx)
	if err != nil {
		return fmt.Errorf("getting product areas from Team Katalogen: %w", err)
	}

	if len(pas) == 0 {
		log.Info().Msg("No product areas found in Team Katalogen")

		return nil
	}

	var allTeams []*service.TeamkatalogenTeam

	for _, pa := range pas {
		teams, err := r.api.GetTeamsInProductArea(ctx, pa.ID)
		if err != nil {
			return fmt.Errorf("getting teams in product area %s: %w", pa.Name, err)
		}
		allTeams = append(allTeams, teams...)
	}

	inputTeams := make([]*service.UpsertTeamRequest, len(allTeams))
	for i, team := range allTeams {
		inputTeams[i] = &service.UpsertTeamRequest{
			ID:            team.ID,
			ProductAreaID: team.ProductAreaID,
			Name:          team.Name,
		}
	}

	inputProductAreas := make([]*service.UpsertProductAreaRequest, len(pas))
	for i, pa := range pas {
		inputProductAreas[i] = &service.UpsertProductAreaRequest{
			ID:   pa.ID,
			Name: pa.Name,
		}
	}

	err = r.storage.UpsertProductAreaAndTeam(ctx, inputProductAreas, inputTeams)
	if err != nil {
		return fmt.Errorf("upserting product areas and teams: %w", err)
	}

	return nil
}
