package worker

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/worker/worker_args"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type MetabaseCreatePermissionGroup struct {
	river.WorkerDefaults[worker_args.MetabaseCreatePermissionGroupJob]

	api   service.MetabaseAPI
	store service.MetabaseStorage

	repo *database.Repo
}

func (w *MetabaseCreatePermissionGroup) Work(ctx context.Context, job *river.Job[worker_args.MetabaseCreatePermissionGroupJob]) error {
	datasetID, err := uuid.Parse(job.Args.DatasetID)
	if err != nil {
		return fmt.Errorf("parsing dataset ID: %w", err)
	}

	meta, err := w.store.GetMetadata(ctx, datasetID, false)
	if err != nil {
		return fmt.Errorf("getting metadata: %w", err)
	}

	if meta.PermissionGroupID == nil {
		groupID, err := w.api.GetOrCreatePermissionGroup(ctx, job.Args.PermissionGroupName)
		if err != nil {
			return fmt.Errorf("creating permission group: %w", err)
		}

		// FIXME: figure out how to do this in a transaction
		_, err = w.store.SetPermissionGroupMetabaseMetadata(ctx, datasetID, groupID)
		if err != nil {
			return fmt.Errorf("setting permission group metadata: %w", err)
		}
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("metabase create permission group transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("metabase create permission group transaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("metabase create permission group transaction: %w", err)
	}

	return nil
}
