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

type MetabaseCreateRestrictedCollection struct {
	river.WorkerDefaults[worker_args.MetabaseCreateRestrictedCollectionJob]

	api   service.MetabaseAPI
	store service.MetabaseStorage

	repo *database.Repo
}

// FIXME: figure out how to write unit tests for this
func (w *MetabaseCreateRestrictedCollection) Work(ctx context.Context, job *river.Job[worker_args.MetabaseCreateRestrictedCollectionJob]) error {
	datasetID, err := uuid.Parse(job.Args.DatasetID)
	if err != nil {
		return fmt.Errorf("parsing dataset ID: %w", err)
	}

	meta, err := w.store.GetMetadata(ctx, datasetID, false)
	if err != nil {
		return fmt.Errorf("getting metadata: %w", err)
	}

	if meta.CollectionID == nil {
		colID, err := w.api.CreateCollectionWithAccess(ctx, *meta.PermissionGroupID, job.Args.CollectionName, true)
		if err != nil {
			return fmt.Errorf("creating collection with access: %w", err)
		}

		_, err = w.store.SetCollectionMetabaseMetadata(ctx, datasetID, colID)
		if err != nil {
			return fmt.Errorf("setting collection metadata: %w", err)
		}
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("metabase create restricted collection transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("metabase create restricted collection transaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("metabase create restricted collection transaction: %w", err)
	}

	return nil
}

type MetabaseEnsureServiceAccount struct {
	river.WorkerDefaults[worker_args.MetabaseEnsureServiceAccountJob]

	api   service.ServiceAccountAPI
	store service.MetabaseStorage

	repo *database.Repo
}

func (w *MetabaseEnsureServiceAccount) Work(ctx context.Context, job *river.Job[worker_args.MetabaseEnsureServiceAccountJob]) error {
	datasetID, err := uuid.Parse(job.Args.DatasetID)
	if err != nil {
		return fmt.Errorf("parsing dataset ID: %w", err)
	}

	meta, err := w.store.GetMetadata(ctx, datasetID, false)
	if err != nil {
		return fmt.Errorf("getting metadata: %w", err)
	}

	if meta.SAEmail == "" {
		sa, err := w.api.EnsureServiceAccount(ctx, &service.ServiceAccountRequest{
			AccountID:   job.Args.AccountID,
			ProjectID:   job.Args.ProjectID,
			DisplayName: job.Args.DisplayName,
			Description: job.Args.Description,
		})
		if err != nil {
			return fmt.Errorf("creating service account: %w", err)
		}

		_, err = w.store.SetServiceAccountMetabaseMetadata(ctx, datasetID, sa.Email)
		if err != nil {
			return fmt.Errorf("setting service account metadata: %w", err)
		}
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("metabase ensure service account transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("metabase ensure service account transaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("metabase ensure service account transaction: %w", err)
	}

	return nil
}

type MetabaseAddProjectIAMPolicyBindingJob struct {
	river.WorkerDefaults[worker_args.MetabaseAddProjectIAMPolicyBindingJob]

	api  service.CloudResourceManagerAPI
	repo *database.Repo
}

func (w *MetabaseAddProjectIAMPolicyBindingJob) Work(ctx context.Context, job *river.Job[worker_args.MetabaseAddProjectIAMPolicyBindingJob]) error {
	err := w.api.AddProjectIAMPolicyBinding(ctx, job.Args.ProjectID, &service.Binding{
		Role:    job.Args.Role,
		Members: []string{job.Args.Member},
	})
	if err != nil {
		return fmt.Errorf("adding project IAM policy binding: %w", err)
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("completing job: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commiting: %w", err)
	}

	return nil
}

type MetabaseCreateBigqueryDatabaseJob struct {
	river.WorkerDefaults[worker_args.MetabaseCreateRestrictedBigqueryDatabaseJob]

	service service.MetabaseService
	repo    *database.Repo
}

func (w *MetabaseCreateBigqueryDatabaseJob) Work(ctx context.Context, job *river.Job[worker_args.MetabaseCreateRestrictedBigqueryDatabaseJob]) error {
	datasetID, err := uuid.Parse(job.Args.DatasetID)
	if err != nil {
		return fmt.Errorf("parsing dataset ID: %w", err)
	}

	err = w.service.CreateRestrictedMetabaseBigqueryDatabase(ctx, datasetID)
	if err != nil {
		return fmt.Errorf("creating metabase bigquery database: %w", err)
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("completing job: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commiting: %w", err)
	}

	return nil
}

type MetabaseVerifyRestrictedBigqueryDatabaseJob struct {
	river.WorkerDefaults[worker_args.MetabaseVerifyRestrictedBigqueryDatabaseJob]

	service service.MetabaseService
	repo    *database.Repo
}

func (w *MetabaseVerifyRestrictedBigqueryDatabaseJob) Work(ctx context.Context, job *river.Job[worker_args.MetabaseVerifyRestrictedBigqueryDatabaseJob]) error {
	datasetID, err := uuid.Parse(job.Args.DatasetID)
	if err != nil {
		return fmt.Errorf("parsing dataset ID: %w", err)
	}

	err = w.service.VerifyRestrictedMetabaseBigqueryDatabase(ctx, datasetID)
	if err != nil {
		return fmt.Errorf("verifying metabase bigquery database: %w", err)
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("completing job: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commiting: %w", err)
	}

	return nil
}

type MetabaseFinalizeRestrictedBigqueryDatabaseJob struct {
	river.WorkerDefaults[worker_args.MetabaseFinalizeRestrictedBigqueryDatabaseJob]
	service service.MetabaseService
	repo    *database.Repo
}

func (w *MetabaseFinalizeRestrictedBigqueryDatabaseJob) Work(ctx context.Context, job *river.Job[worker_args.MetabaseFinalizeRestrictedBigqueryDatabaseJob]) error {
	datasetID, err := uuid.Parse(job.Args.DatasetID)
	if err != nil {
		return fmt.Errorf("parsing dataset ID: %w", err)
	}

	err = w.service.FinalizeRestrictedMetabaseBigqueryDatabase(ctx, datasetID)
	if err != nil {
		return fmt.Errorf("finalizing metabase bigquery database: %w", err)
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("completing job: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commiting: %w", err)
	}

	return nil
}

type MetabaseDeleteRestrictedBigqueryDatabaseJob struct {
	river.WorkerDefaults[worker_args.MetabaseDeleteRestrictedBigqueryDatabaseJob]

	service service.MetabaseService
	repo    *database.Repo
}

func (w *MetabaseDeleteRestrictedBigqueryDatabaseJob) Work(ctx context.Context, job *river.Job[worker_args.MetabaseDeleteRestrictedBigqueryDatabaseJob]) error {
	datasetID, err := uuid.Parse(job.Args.DatasetID)
	if err != nil {
		return fmt.Errorf("parsing dataset ID: %w", err)
	}

	err = w.service.DeleteRestrictedMetabaseBigqueryDatabase(ctx, datasetID)
	if err != nil {
		return fmt.Errorf("cleaning up failed metabase bigquery database: %w", err)
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("completing job: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commiting: %w", err)
	}

	return nil
}
