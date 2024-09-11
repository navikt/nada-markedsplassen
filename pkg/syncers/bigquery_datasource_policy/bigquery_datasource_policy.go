package bigquery_datasource_policy

import (
	"context"
	"fmt"
	"strings"

	"github.com/navikt/nada-backend/pkg/bq"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/syncers"
	"github.com/rs/zerolog"
)

var _ syncers.Runner = &Runner{}

type Runner struct {
	storage       service.BigQueryStorage
	ops           bq.Operations
	evaluateRoles []string
}

func (r *Runner) Name() string {
	return "BigQueryDatasourcePolicyCleaner"
}

func (r *Runner) RunOnce(ctx context.Context, log zerolog.Logger) error {
	sources, err := r.storage.GetBigqueryDatasources(ctx)
	if err != nil {
		return fmt.Errorf("getting bigquery datasources: %w", err)
	}

	var errs []string

	for _, s := range sources {
		err := r.ops.UpdateTablePolicy(
			ctx,
			s.ProjectID,
			s.Dataset,
			s.Table,
			bq.RemoveDeletedMembersWithRole(s.ProjectID, s.Dataset, s.Table, r.evaluateRoles, log),
		)
		if err != nil {
			errs = append(errs, fmt.Errorf("updating policy for %s.%s.%s: %w", s.ProjectID, s.Dataset, s.Table, err).Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("updating policies: %s", strings.Join(errs, ","))
	}

	return nil
}

func New(evaluateRoles []string, storage service.BigQueryStorage, ops bq.Operations) *Runner {
	return &Runner{
		storage:       storage,
		ops:           ops,
		evaluateRoles: evaluateRoles,
	}
}
