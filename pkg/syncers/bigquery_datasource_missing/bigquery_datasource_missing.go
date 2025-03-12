package bigquery_datasource_missing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/navikt/nada-backend/pkg/bq"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/syncers"
	"github.com/rs/zerolog"
)

const (
	removalTime = -168 * time.Hour // 1 week
)

var _ syncers.Runner = &Runner{}

type Runner struct {
	bqOps              bq.Operations
	bqStorage          service.BigQueryStorage
	metabaseService    service.MetabaseService
	metabaseStorage    service.MetabaseStorage
	dataProductStorage service.DataProductsStorage
}

func (r *Runner) Name() string {
	return "BigQueryDatasourceMissingSynchronizer"
}

func (r *Runner) RunOnce(ctx context.Context, log zerolog.Logger) error {
	sources, err := r.bqStorage.GetBigqueryDatasources(ctx)
	if err != nil {
		return err
	}

	var missing []*service.BigQuery

	for _, s := range sources {
		// We set MissingSince as part of the bigquery_sync_tables syncer
		if s.MissingSince == nil {
			continue
		}

		// Check if the table still doesn't exist
		_, err := r.bqOps.GetTable(ctx, s.ProjectID, s.Dataset, s.Table)
		if err != nil {
			if errors.Is(err, bq.ErrNotExist) {
				log.Info().Fields(map[string]interface{}{
					"project_id": s.ProjectID,
					"dataset":    s.Dataset,
					"table":      s.Table,
				}).Msg("found missing table")

				missing = append(missing, s)

				continue
			}

			return fmt.Errorf("getting table %s: %w", s.Dataset, err)
		}

		// If the table exists, we remove the missing flag
		err = r.bqStorage.UpdateBigqueryDatasourceNotMissing(ctx, s.DatasetID)
		if err != nil {
			return fmt.Errorf("updating missing table %s: %w", s.DatasetID, err)
		}
	}

	for _, m := range missing {
		// We only want to remove datasets that have been missing for a week
		if m.MissingSince.After(time.Now().Add(removalTime)) {
			continue
		}

		fields := map[string]interface{}{
			"project_id": m.ProjectID,
			"dataset":    m.Dataset,
			"table":      m.Table,
		}

		log.Info().Fields(fields).Msg("removing missing table")

		err := r.removeFromMetabase(ctx, m, log)
		if err != nil {
			return err
		}

		log.Info().Fields(fields).Msg("done removing from metabase")

		err = r.dataProductStorage.DeleteDataset(ctx, m.DatasetID)
		if err != nil {
			return fmt.Errorf("deleting dataset %s: %w", m.DatasetID, err)
		}

		log.Info().Fields(fields).Msg("done removing from data products")

		err = r.bqStorage.DeleteBigqueryDatasource(ctx, m.DatasetID)
		if err != nil {
			return fmt.Errorf("deleting bigquery datasource %s: %w", m.DatasetID, err)
		}

		log.Info().Fields(fields).Msg("done removing from bigquery")
	}

	return nil
}

func (r *Runner) removeFromMetabase(ctx context.Context, ds *service.BigQuery, log zerolog.Logger) error {
	_, err := r.metabaseStorage.GetMetadata(ctx, ds.DatasetID, true)
	if err != nil {
		if errs.KindIs(errs.NotExist, err) {
			log.Info().Msgf("no metadata found for dataset %s, skipping metabase removal", ds.DatasetID)

			return nil
		}

		return fmt.Errorf("getting metadata for dataset %s: %w", ds.DatasetID, err)
	}

	err = r.metabaseService.DeleteDatabase(ctx, ds.DatasetID)
	if err != nil {
		return fmt.Errorf("deleting dataset %s from metabase: %w", ds.DatasetID, err)
	}

	return nil
}

func New(
	bqOps bq.Operations,
	bqStorage service.BigQueryStorage,
	mbService service.MetabaseService,
	mbStorage service.MetabaseStorage,
	dpStorage service.DataProductsStorage,
) *Runner {
	return &Runner{
		bqOps:              bqOps,
		bqStorage:          bqStorage,
		metabaseService:    mbService,
		metabaseStorage:    mbStorage,
		dataProductStorage: dpStorage,
	}
}
