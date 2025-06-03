package river

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/worker/worker_args"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"riverqueue.com/riverpro"
)

var _ service.MetabaseQueue = &metabaseQueue{}

type metabaseQueue struct {
	repo   *database.Repo
	config *riverpro.Config
}

func (q *metabaseQueue) CreateOpenMetabaseBigqueryDatabaseWorkflow(ctx context.Context, datasetID uuid.UUID) (*service.MetabaseOpenBigqueryDatabaseWorkflowStatus, error) {
	const op errs.Op = "metabaseQueue.CreateOpenMetabaseBigqueryDatabaseWorkflow"

	client, err := NewClient(q.repo, q.config)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	tx, err := q.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}
	defer tx.Rollback(ctx)

	insertOpts := &river.InsertOpts{
		MaxAttempts: 5,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 10 * time.Minute,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable,
				rivertype.JobStatePending,
				rivertype.JobStateRunning,
				rivertype.JobStateRetryable,
				rivertype.JobStateScheduled,
			},
		},
		Queue:    worker_args.MetabaseQueue,
		Metadata: []byte(metabaseJobMetadata(datasetID)),
	}

	_, err = client.InsertManyTx(ctx, tx, []river.InsertManyParams{
		{
			Args: &worker_args.MetabasePreflightCheckOpenBigqueryDatabaseJob{
				DatasetID: datasetID.String(),
			},
			InsertOpts: insertOpts,
		},
		{
			Args: &worker_args.MetabaseCreateOpenBigqueryDatabaseJob{
				DatasetID: datasetID.String(),
			},
			InsertOpts: insertOpts,
		},
		{
			Args: &worker_args.MetabaseVerifyOpenBigqueryDatabaseJob{
				DatasetID: datasetID.String(),
			},
			InsertOpts: insertOpts,
		},
		{
			Args: &worker_args.MetabaseFinalizeOpenBigqueryDatabaseJob{
				DatasetID: datasetID.String(),
			},
			InsertOpts: insertOpts,
		},
	})
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("transaction: %w", err))
	}

	job, err := q.GetOpenMetabaseBigQueryDatabaseWorkflow(ctx, datasetID)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	return job, nil
}

func (q *metabaseQueue) GetOpenMetabaseBigQueryDatabaseWorkflow(ctx context.Context, datasetID uuid.UUID) (*service.MetabaseOpenBigqueryDatabaseWorkflowStatus, error) {
	const op errs.Op = "metabaseQueue.GetOpenMetabaseBigQueryDatabaseWorkflow"

	client, err := NewClient(q.repo, q.config)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	baseParams := river.NewJobListParams().
		Queues(worker_args.MetabaseQueue).
		States(
			rivertype.JobStateAvailable,
			rivertype.JobStateRunning,
			rivertype.JobStateRetryable,
			rivertype.JobStateCompleted,
			rivertype.JobStateDiscarded,
			rivertype.JobStatePending,
		).
		Metadata(metabaseJobMetadata(datasetID)).
		OrderBy("id", river.SortOrderDesc).
		First(1)

	raw, err := client.JobList(ctx, baseParams.Kinds(worker_args.MetabasePreflightCheckOpenBigqueryDatabaseJobKind))
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, fmt.Errorf("found no preflight check jobs for dataset: %s", datasetID.String()))
	}

	if len(raw.Jobs) != 1 {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("expected 1 preflight check job, got: %d", len(raw.Jobs)))
	}

	preflightJob, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.MetabasePreflightCheckOpenBigqueryDatabaseJob) *service.MetabasePreflightCheckOpenBigqueryDatabaseJob {
		return &service.MetabasePreflightCheckOpenBigqueryDatabaseJob{
			JobHeader: header,
			DatasetID: uuid.MustParse(args.DatasetID),
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeTransactionalQueue, op, err)
	}

	raw, err = client.JobList(ctx, baseParams.Kinds(worker_args.MetabaseCreateOpenBigqueryDatabaseJobKind))
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, fmt.Errorf("found no open jobs for dataset: %s", datasetID.String()))
	}

	if len(raw.Jobs) != 1 {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("expected 1 open job, got: %d", len(raw.Jobs)))
	}

	databaseJob, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.MetabaseCreateOpenBigqueryDatabaseJob) *service.MetabaseCreateOpenBigqueryDatabaseJob {
		return &service.MetabaseCreateOpenBigqueryDatabaseJob{
			JobHeader: header,
			DatasetID: uuid.MustParse(args.DatasetID),
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeTransactionalQueue, op, err)
	}

	raw, err = client.JobList(ctx, baseParams.Kinds(worker_args.MetabaseVerifyOpenBigqueryDatabaseJobKind))
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, fmt.Errorf("found no verify jobs for dataset: %s", datasetID.String()))
	}

	if len(raw.Jobs) != 1 {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("expected 1 verify job, got: %d", len(raw.Jobs)))
	}

	verifyJob, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.MetabaseVerifyOpenBigqueryDatabaseJob) *service.MetabaseVerifyOpenBigqueryDatabaseJob {
		return &service.MetabaseVerifyOpenBigqueryDatabaseJob{
			JobHeader: header,
			DatasetID: uuid.MustParse(args.DatasetID),
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeTransactionalQueue, op, err)
	}

	raw, err = client.JobList(ctx, baseParams.Kinds(worker_args.MetabaseFinalizeOpenBigqueryDatabaseJobKind))
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, fmt.Errorf("found no finalize jobs for dataset: %s", datasetID.String()))
	}

	if len(raw.Jobs) != 1 {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("expected 1 finalize job, got: %d", len(raw.Jobs)))
	}

	finalizeJob, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.MetabaseFinalizeOpenBigqueryDatabaseJob) *service.MetabaseFinalizeOpenBigqueryDatabaseJob {
		return &service.MetabaseFinalizeOpenBigqueryDatabaseJob{
			JobHeader: header,
			DatasetID: uuid.MustParse(args.DatasetID),
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeTransactionalQueue, op, err)
	}

	return &service.MetabaseOpenBigqueryDatabaseWorkflowStatus{
		PreflightCheckJob: preflightJob,
		DatabaseJob:       databaseJob,
		VerifyJob:         verifyJob,
		FinalizeJob:       finalizeJob,
	}, nil
}

func (q *metabaseQueue) CreateMetabaseBigqueryDatabaseDeleteJob(ctx context.Context, datasetID uuid.UUID) (*service.MetabaseBigqueryDatabaseDeleteJob, error) {
	const op errs.Op = "metabaseQueue.CreateMetabaseBigqueryDatabaseDeleteJob"

	client, err := NewClient(q.repo, q.config)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	tx, err := q.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}
	defer tx.Rollback(ctx)

	insertOpts := &river.InsertOpts{
		MaxAttempts: 5,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 10 * time.Minute,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable,
				rivertype.JobStatePending,
				rivertype.JobStateRunning,
				rivertype.JobStateRetryable,
				rivertype.JobStateScheduled,
			},
		},
		Queue:    worker_args.MetabaseQueue,
		Metadata: []byte(metabaseJobMetadata(datasetID)),
	}

	_, err = client.InsertTx(ctx, tx, &worker_args.MetabaseDeleteRestrictedBigqueryDatabaseJob{
		DatasetID: datasetID.String(),
	}, insertOpts)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("transaction: %w", err))
	}

	job, err := q.GetMetabaseBigqueryDatabaseDeleteJob(ctx, datasetID)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	return job, nil
}

func (q *metabaseQueue) GetMetabaseBigqueryDatabaseDeleteJob(ctx context.Context, datasetID uuid.UUID) (*service.MetabaseBigqueryDatabaseDeleteJob, error) {
	const op errs.Op = "metabaseQueue.GetMetabaseBigqueryDatabaseDeleteJob"

	client, err := NewClient(q.repo, q.config)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	params := river.NewJobListParams().
		Queues(worker_args.MetabaseQueue).
		States(
			rivertype.JobStateAvailable,
			rivertype.JobStateRunning,
			rivertype.JobStateRetryable,
			rivertype.JobStateCompleted,
			rivertype.JobStateDiscarded,
		).
		Kinds(worker_args.WorkstationJobKind).
		Metadata(metabaseJobMetadata(datasetID)).
		OrderBy("id", river.SortOrderDesc).
		First(1)

	raw, err := client.JobList(ctx, params)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, fmt.Errorf("found no cleanup jobs for dataset: %s", datasetID.String()))
	}

	job, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.MetabaseDeleteRestrictedBigqueryDatabaseJob) *service.MetabaseBigqueryDatabaseDeleteJob {
		return &service.MetabaseBigqueryDatabaseDeleteJob{
			JobHeader: header,
			DatasetID: uuid.MustParse(args.DatasetID),
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeTransactionalQueue, op, err)
	}

	return job, nil
}

func (q *metabaseQueue) CreateRestrictedMetabaseBigqueryDatabaseWorkflow(ctx context.Context, opts *service.MetabaseRestrictedBigqueryDatabaseWorkflowOpts) (*service.MetabaseRestrictedBigqueryDatabaseWorkflowStatus, error) {
	const op errs.Op = "metabaseQueue.CreateRestrictedMetabaseBigqueryDatabaseWorkflow"

	client, err := NewClient(q.repo, q.config)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	tx, err := q.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}
	defer tx.Rollback(ctx)

	insertOpts := &river.InsertOpts{
		MaxAttempts: 5,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 10 * time.Minute,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable,
				rivertype.JobStatePending,
				rivertype.JobStateRunning,
				rivertype.JobStateRetryable,
				rivertype.JobStateScheduled,
			},
		},
		Queue:    worker_args.MetabaseQueue,
		Metadata: []byte(metabaseJobMetadata(opts.DatasetID)),
	}

	_, err = client.InsertManyTx(ctx, tx, []river.InsertManyParams{
		{
			Args: &worker_args.MetabasePreflightCheckRestrictedBigqueryDatabaseJob{
				DatasetID: opts.DatasetID.String(),
			},
			InsertOpts: insertOpts,
		},
		{
			Args: &worker_args.MetabaseCreatePermissionGroupJob{
				DatasetID:           opts.DatasetID.String(),
				PermissionGroupName: opts.PermissionGroupName,
			},
			InsertOpts: insertOpts,
		},
		{
			Args: &worker_args.MetabaseCreateRestrictedCollectionJob{
				DatasetID:      opts.DatasetID.String(),
				CollectionName: opts.CollectionName,
			},
			InsertOpts: insertOpts,
		},
		{
			Args: &worker_args.MetabaseEnsureServiceAccountJob{
				DatasetID:   opts.DatasetID.String(),
				AccountID:   opts.AccountID,
				ProjectID:   opts.ProjectID,
				DisplayName: opts.DisplayName,
				Description: opts.Description,
			},
			InsertOpts: insertOpts,
		},
		{
			Args: &worker_args.MetabaseCreateServiceAccountKeyJob{
				DatasetID: opts.DatasetID.String(),
			},
			InsertOpts: insertOpts,
		},
		{
			Args: &worker_args.MetabaseAddProjectIAMPolicyBindingJob{
				DatasetID: opts.DatasetID.String(),
				ProjectID: opts.ProjectID,
				Role:      opts.Role,
				Member:    opts.Member,
			},
			InsertOpts: insertOpts,
		},
		{
			Args: &worker_args.MetabaseCreateRestrictedBigqueryDatabaseJob{
				DatasetID: opts.DatasetID.String(),
			},
			InsertOpts: insertOpts,
		},
		{
			Args: &worker_args.MetabaseVerifyRestrictedBigqueryDatabaseJob{
				DatasetID: opts.DatasetID.String(),
			},
			InsertOpts: insertOpts,
		},
		{
			Args: &worker_args.MetabaseFinalizeRestrictedBigqueryDatabaseJob{
				DatasetID: opts.DatasetID.String(),
			},
			InsertOpts: insertOpts,
		},
	})
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("transaction: %w", err))
	}

	status, err := q.GetRestrictedMetabaseBigqueryDatabaseWorkflow(ctx, opts.DatasetID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return status, nil
}

func (q *metabaseQueue) GetRestrictedMetabaseBigqueryDatabaseWorkflow(ctx context.Context, datasetID uuid.UUID) (*service.MetabaseRestrictedBigqueryDatabaseWorkflowStatus, error) {
	const op errs.Op = "metabaseQueue.GetRestrictedMetabaseBigQueryDatabaseWorkflow"

	client, err := NewClient(q.repo, q.config)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	baseParams := river.NewJobListParams().
		Queues(worker_args.MetabaseQueue).
		States(
			rivertype.JobStateAvailable,
			rivertype.JobStateRunning,
			rivertype.JobStateRetryable,
			rivertype.JobStateCompleted,
			rivertype.JobStateDiscarded,
			rivertype.JobStatePending,
		).
		Metadata(metabaseJobMetadata(datasetID)).
		OrderBy("id", river.SortOrderDesc)

	raw, err := client.JobList(ctx, baseParams.Kinds(worker_args.MetabasePreflightCheckRestrictedBigqueryDatabaseJobKind).First(1))
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, fmt.Errorf("found no preflight check jobs for dataset: %s", datasetID.String()))
	}

	if len(raw.Jobs) != 1 {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("expected 1 preflight check job, got: %d", len(raw.Jobs)))
	}

	preflightJob, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.MetabasePreflightCheckRestrictedBigqueryDatabaseJob) *service.MetabasePreflightCheckRestrictedBigqueryDatabaseJob {
		return &service.MetabasePreflightCheckRestrictedBigqueryDatabaseJob{
			JobHeader: header,
			DatasetID: uuid.MustParse(args.DatasetID),
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeTransactionalQueue, op, err)
	}

	raw, err = client.JobList(ctx, baseParams.Kinds(worker_args.MetabaseCreatePermissionGroupJobKind).First(1))
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, fmt.Errorf("found no permission group jobs for dataset: %s", datasetID.String()))
	}

	if len(raw.Jobs) != 1 {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("expected 1 permission group job, got: %d", len(raw.Jobs)))
	}

	permissionGroupJob, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.MetabaseCreatePermissionGroupJob) *service.MetabaseCreatePermissionGroupJob {
		return &service.MetabaseCreatePermissionGroupJob{
			JobHeader:           header,
			DatasetID:           uuid.MustParse(args.DatasetID),
			PermissionGroupName: args.PermissionGroupName,
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeTransactionalQueue, op, err)
	}

	raw, err = client.JobList(ctx, baseParams.Kinds(worker_args.MetabaseCreateRestrictedCollectionJobKind).First(1))
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, fmt.Errorf("found no collection jobs for dataset: %s", datasetID.String()))
	}

	if len(raw.Jobs) != 1 {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("expected 1 collection job, got: %d", len(raw.Jobs)))
	}

	collectionJob, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.MetabaseCreateRestrictedCollectionJob) *service.MetabaseCreateCollectionJob {
		return &service.MetabaseCreateCollectionJob{
			JobHeader:      header,
			DatasetID:      uuid.MustParse(args.DatasetID),
			CollectionName: args.CollectionName,
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeTransactionalQueue, op, err)
	}

	raw, err = client.JobList(ctx, baseParams.Kinds(worker_args.MetabaseEnsureServiceAccountJobKind).First(1))
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, fmt.Errorf("found no service account jobs for dataset: %s", datasetID.String()))
	}

	if len(raw.Jobs) != 1 {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("expected 1 service account job, got: %d", len(raw.Jobs)))
	}

	serviceAccountJob, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.MetabaseEnsureServiceAccountJob) *service.MetabaseEnsureServiceAccountJob {
		return &service.MetabaseEnsureServiceAccountJob{
			JobHeader:   header,
			DatasetID:   uuid.MustParse(args.DatasetID),
			AccountID:   args.AccountID,
			ProjectID:   args.ProjectID,
			DisplayName: args.DisplayName,
			Description: args.Description,
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeTransactionalQueue, op, err)
	}

	raw, err = client.JobList(ctx, baseParams.Kinds(worker_args.MetabaseCreateServiceAccountKeyJobKind).First(1))
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, fmt.Errorf("found no service account key jobs for dataset: %s", datasetID.String()))
	}

	if len(raw.Jobs) != 1 {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("expected 1 service account key job, got: %d", len(raw.Jobs)))
	}

	serviceAccountKeyJob, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.MetabaseCreateServiceAccountKeyJob) *service.MetabaseCreateServiceAccountKeyJob {
		return &service.MetabaseCreateServiceAccountKeyJob{
			JobHeader: header,
			DatasetID: uuid.MustParse(args.DatasetID),
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeTransactionalQueue, op, err)
	}

	raw, err = client.JobList(ctx, baseParams.Kinds(worker_args.MetabaseAddProjectIAMPolicyBindingJobKind).First(1))
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, fmt.Errorf("found no iam policy jobs for dataset: %s", datasetID.String()))
	}

	if len(raw.Jobs) != 1 {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("expected 1 project IAM job, got: %d", len(raw.Jobs)))
	}

	projectIAMJob, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.MetabaseAddProjectIAMPolicyBindingJob) *service.MetabaseAddProjectIAMPolicyBindingJob {
		return &service.MetabaseAddProjectIAMPolicyBindingJob{
			JobHeader: header,
			DatasetID: uuid.MustParse(args.DatasetID),
			ProjectID: args.ProjectID,
			Role:      args.Role,
			Member:    args.Member,
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeTransactionalQueue, op, err)
	}

	raw, err = client.JobList(ctx, baseParams.Kinds(worker_args.MetabaseCreateRestrictedBigqueryDatabaseJobKind).First(1))
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, fmt.Errorf("found no create database jobs for dataset: %s", datasetID.String()))
	}

	if len(raw.Jobs) != 1 {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("expected 1 create database job, got: %d", len(raw.Jobs)))
	}

	databaseJob, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.MetabaseCreateRestrictedBigqueryDatabaseJob) *service.MetabaseBigqueryCreateDatabaseJob {
		return &service.MetabaseBigqueryCreateDatabaseJob{
			JobHeader: header,
			DatasetID: uuid.MustParse(args.DatasetID),
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeTransactionalQueue, op, err)
	}

	raw, err = client.JobList(ctx, baseParams.Kinds(worker_args.MetabaseVerifyRestrictedBigqueryDatabaseJobKind).First(1))
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, fmt.Errorf("found no verify database jobs for dataset: %s", datasetID.String()))
	}

	if len(raw.Jobs) != 1 {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("expected 1 verify database job, got: %d", len(raw.Jobs)))
	}

	verifyJob, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.MetabaseVerifyRestrictedBigqueryDatabaseJob) *service.MetabaseBigqueryVerifyDatabaseJob {
		return &service.MetabaseBigqueryVerifyDatabaseJob{
			JobHeader: header,
			DatasetID: uuid.MustParse(args.DatasetID),
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeTransactionalQueue, op, err)
	}

	raw, err = client.JobList(ctx, baseParams.Kinds(worker_args.MetabaseFinalizeRestrictedBigqueryDatabaseJobKind).First(1))
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) != 1 {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("expected 1 finalize database job, got: %d", len(raw.Jobs)))
	}

	if len(raw.Jobs) == 0 {
		return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, fmt.Errorf("found no finalize database jobs for dataset: %s", datasetID.String()))
	}

	finalizeJob, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.MetabaseFinalizeRestrictedBigqueryDatabaseJob) *service.MetabaseBigqueryFinalizeDatabaseJob {
		return &service.MetabaseBigqueryFinalizeDatabaseJob{
			JobHeader: header,
			DatasetID: uuid.MustParse(args.DatasetID),
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeTransactionalQueue, op, err)
	}

	return &service.MetabaseRestrictedBigqueryDatabaseWorkflowStatus{
		PreflightCheckJob:    preflightJob,
		PermissionGroupJob:   permissionGroupJob,
		CollectionJob:        collectionJob,
		ServiceAccountJob:    serviceAccountJob,
		ServiceAccountKeyJob: serviceAccountKeyJob,
		ProjectIAMJob:        projectIAMJob,
		DatabaseJob:          databaseJob,
		VerifyJob:            verifyJob,
		FinalizeJob:          finalizeJob,
	}, nil
}

func metabaseJobMetadata(datasetID uuid.UUID) string {
	return fmt.Sprintf(`{"datasetID": "%s"}`, datasetID)
}

func NewMetabaseQueue(repo *database.Repo, config *riverpro.Config) *metabaseQueue {
	return &metabaseQueue{
		repo:   repo,
		config: config,
	}
}
