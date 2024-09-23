package metabase_tables

import (
	"context"
	"fmt"

	"github.com/navikt/nada-backend/pkg/syncers"
	"github.com/rs/zerolog"

	"github.com/navikt/nada-backend/pkg/service"
)

var _ syncers.Runner = &Runner{}

type Runner struct {
	service service.MetabaseService
}

func New(service service.MetabaseService) *Runner {
	return &Runner{
		service: service,
	}
}

func (r *Runner) Name() string {
	return "MetabaseTableVisibilitySynchronizer"
}

func (r *Runner) RunOnce(ctx context.Context, _ zerolog.Logger) error {
	err := r.service.SyncAllTablesVisibility(ctx)
	if err != nil {
		return fmt.Errorf("syncing tables visibility: %w", err)
	}

	return nil
}
