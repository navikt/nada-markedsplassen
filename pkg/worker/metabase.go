package worker

import (
	"context"
	"fmt"
	"riverqueue.com/riverpro"

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

	service service.MetabaseService
	repo    *database.Repo
}

func (w *MetabaseCreatePermissionGroup) Work(ctx context.Context, job *river.Job[worker_args.MetabaseCreatePermissionGroupJob]) error {
	datasetID, err := uuid.Parse(job.Args.DatasetID)
	if err != nil {
		return fmt.Errorf("parsing dataset ID: %w", err)
	}

	err = w.service.EnsurePermissionGroup(ctx, datasetID, job.Args.PermissionGroupName)
	if err != nil {
		return fmt.Errorf("ensuring permission group: %w", err)
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

	service service.MetabaseService

	repo *database.Repo
}

func (w *MetabaseCreateRestrictedCollection) Work(ctx context.Context, job *river.Job[worker_args.MetabaseCreateRestrictedCollectionJob]) error {
	datasetID, err := uuid.Parse(job.Args.DatasetID)
	if err != nil {
		return fmt.Errorf("parsing dataset ID: %w", err)
	}

	err = w.service.CreateRestrictedCollection(ctx, datasetID, job.Args.CollectionName)
	if err != nil {
		return fmt.Errorf("creating restricted collection: %w", err)
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

	service service.MetabaseService
	repo    *database.Repo
}

func (w *MetabaseEnsureServiceAccount) Work(ctx context.Context, job *river.Job[worker_args.MetabaseEnsureServiceAccountJob]) error {
	datasetID, err := uuid.Parse(job.Args.DatasetID)
	if err != nil {
		return fmt.Errorf("parsing dataset ID: %w", err)
	}

	err = w.service.CreateMetabaseServiceAccount(ctx, datasetID, &service.ServiceAccountRequest{
		AccountID:   job.Args.AccountID,
		ProjectID:   job.Args.ProjectID,
		DisplayName: job.Args.DisplayName,
		Description: job.Args.Description,
	})
	if err != nil {
		return fmt.Errorf("creating metabase service account: %w", err)
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

	service service.MetabaseService
	repo    *database.Repo
}

func (w *MetabaseAddProjectIAMPolicyBindingJob) Work(ctx context.Context, job *river.Job[worker_args.MetabaseAddProjectIAMPolicyBindingJob]) error {
	err := w.service.CreateMetabaseProjectIAMPolicyBinding(ctx, job.Args.ProjectID, job.Args.Role, job.Args.Member)
	if err != nil {
		return fmt.Errorf("creating metabase project IAM policy binding: %w", err)
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

func MetabaseAddWorkers(config *riverpro.Config, service service.MetabaseService, repo *database.Repo) error {
	err := river.AddWorkerSafely[worker_args.MetabaseCreatePermissionGroupJob](config.Workers, &MetabaseCreatePermissionGroup{
		service: service,
		repo:    repo,
	})
	if err != nil {
		return fmt.Errorf("adding workstation worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.MetabaseCreateRestrictedCollectionJob](config.Workers, &MetabaseCreateRestrictedCollection{
		service: service,
		repo:    repo,
	})
	if err != nil {
		return fmt.Errorf("adding workstation worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.MetabaseEnsureServiceAccountJob](config.Workers, &MetabaseEnsureServiceAccount{
		service: service,
		repo:    repo,
	})
	if err != nil {
		return fmt.Errorf("adding workstation worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.MetabaseAddProjectIAMPolicyBindingJob](config.Workers, &MetabaseAddProjectIAMPolicyBindingJob{
		service: service,
		repo:    repo,
	})
	if err != nil {
		return fmt.Errorf("adding workstation worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.MetabaseCreateRestrictedBigqueryDatabaseJob](config.Workers, &MetabaseCreateBigqueryDatabaseJob{
		service: service,
		repo:    repo,
	})
	if err != nil {
		return fmt.Errorf("adding workstation worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.MetabaseVerifyRestrictedBigqueryDatabaseJob](config.Workers, &MetabaseVerifyRestrictedBigqueryDatabaseJob{
		service: service,
		repo:    repo,
	})
	if err != nil {
		return fmt.Errorf("adding workstation worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.MetabaseFinalizeRestrictedBigqueryDatabaseJob](config.Workers, &MetabaseFinalizeRestrictedBigqueryDatabaseJob{
		service: service,
		repo:    repo,
	})
	if err != nil {
		return fmt.Errorf("adding workstation worker: %w", err)
	}

	return nil
}
