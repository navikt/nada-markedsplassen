package teamprojectsupdater

import (
	"context"
	"fmt"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/syncers"
	"github.com/rs/zerolog"
)

var _ syncers.Runner = &Runner{}

type Runner struct {
	service service.NaisConsoleService
}

func New(service service.NaisConsoleService) *Runner {
	return &Runner{
		service: service,
	}
}

func (r *Runner) Name() string {
	return "TeamProjectsUpdater"
}

func (r *Runner) RunOnce(ctx context.Context, _ zerolog.Logger) error {
	err := r.service.UpdateAllTeamProjects(ctx)
	if err != nil {
		return fmt.Errorf("updating all team projects %w", err)
	}

	return nil
}
