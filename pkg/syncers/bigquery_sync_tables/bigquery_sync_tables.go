package bigquery_sync_tables

import (
	"context"
	"fmt"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/syncers"
	"github.com/rs/zerolog"
)

var _ syncers.Runner = &Runner{}

type Runner struct {
	service service.BigQueryService
}

func (r *Runner) Name() string {
	return "BigQueryDatasourceSynchronizer"
}

func (r *Runner) RunOnce(ctx context.Context, _ zerolog.Logger) error {
	err := r.service.SyncBigQueryTables(ctx)
	if err != nil {
		return fmt.Errorf("syncing bigquery tables: %w", err)
	}

	return nil
}

func New(service service.BigQueryService) *Runner {
	return &Runner{
		service: service,
	}
}
