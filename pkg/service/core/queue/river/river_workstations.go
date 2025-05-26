package river

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"riverqueue.com/riverpro"

	"github.com/navikt/nada-backend/pkg/worker/worker_args"

	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/riverqueue/river/rivertype"
)

var _ service.WorkstationsQueue = (*workstationsQueue)(nil)

type workstationsQueue struct {
	repo   *database.Repo
	config *riverpro.Config
}

func (s *workstationsQueue) GetWorkstationDisconnectJob(ctx context.Context, ident string) (*service.WorkstationDisconnectJob, error) {
	const op errs.Op = "workstationQueue.GetWorkstationDisconnectJob"

	client, err := NewClient(s.repo, s.config)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	params := river.NewJobListParams().
		Queues(worker_args.WorkstationConnectivityQueue).
		States(
			rivertype.JobStateAvailable,
			rivertype.JobStateRunning,
			rivertype.JobStateRetryable,
			rivertype.JobStateCompleted,
			rivertype.JobStateDiscarded,
		).
		Kinds(worker_args.WorkstationDisconnectKind).
		Metadata(workstationJobMetadata(ident)).
		OrderBy("id", river.SortOrderDesc)

	raw, err := client.JobList(ctx, params)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, nil
	}

	job, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.WorkstationDisconnectJob) *service.WorkstationDisconnectJob {
		return &service.WorkstationDisconnectJob{
			JobHeader: header,
			Ident:     args.Ident,
			Hosts:     args.Hosts,
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err)
	}

	return job, nil
}

func (s *workstationsQueue) GetWorkstationConnectJobs(ctx context.Context, ident string) ([]*service.WorkstationConnectJob, error) {
	const op errs.Op = "workstationsQueue.GetWorkstationConnectJob"

	client, err := NewClient(s.repo, s.config)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	params := river.NewJobListParams().
		Queues(worker_args.WorkstationConnectivityQueue).
		States(
			rivertype.JobStateAvailable,
			rivertype.JobStateRunning,
			rivertype.JobStateRetryable,
			rivertype.JobStateCompleted,
			rivertype.JobStateDiscarded,
		).
		Kinds(worker_args.WorkstationConnectKind).
		Metadata(workstationJobMetadata(ident)).
		OrderBy("id", river.SortOrderDesc)

	raw, err := client.JobList(ctx, params)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, nil
	}

	jobs := make([]*service.WorkstationConnectJob, len(raw.Jobs))

	for i, r := range raw.Jobs {
		job, err := FromRiverJob(r, func(header service.JobHeader, args worker_args.WorkstationConnectJob) *service.WorkstationConnectJob {
			return &service.WorkstationConnectJob{
				JobHeader: header,
				Ident:     args.Ident,
				Host:      args.Host,
			}
		})
		if err != nil {
			return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err)
		}

		jobs[i] = job
	}

	return jobs, nil
}

func (s *workstationsQueue) GetWorkstationNotifyJob(ctx context.Context, ident string) (*service.WorkstationNotifyJob, error) {
	const op errs.Op = "workstationsQueue.GetWorkstationNotifyJob"

	client, err := NewClient(s.repo, s.config)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	params := river.NewJobListParams().
		Queues(worker_args.WorkstationConnectivityQueue).
		States(
			rivertype.JobStateAvailable,
			rivertype.JobStateRunning,
			rivertype.JobStateRetryable,
			rivertype.JobStateCompleted,
			rivertype.JobStateDiscarded,
		).
		Kinds(worker_args.WorkstationNotifyKind).
		Metadata(workstationJobMetadata(ident)).
		OrderBy("id", river.SortOrderDesc)

	raw, err := client.JobList(ctx, params)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	if len(raw.Jobs) == 0 {
		return nil, nil
	}

	job, err := FromRiverJob(raw.Jobs[0], func(header service.JobHeader, args worker_args.WorkstationNotifyJob) *service.WorkstationNotifyJob {
		return &service.WorkstationNotifyJob{
			JobHeader: header,
			Ident:     args.Ident,
			RequestID: args.RequestID,
			Hosts:     args.Hosts,
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err)
	}

	return job, nil
}

func (s *workstationsQueue) CreateWorkstationConnectivityWorkflow(ctx context.Context, ident string, requestID string, hosts []string) error {
	const op errs.Op = "workstationsQueue.CreateWorkstationConnectivityWorkflow"

	client, err := NewClient(s.repo, s.config)
	if err != nil {
		return errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	tx, err := s.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}
	defer tx.Rollback(ctx)

	insertOpts := &river.InsertOpts{
		MaxAttempts: 5,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 15 * time.Minute,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable,
				rivertype.JobStatePending,
				rivertype.JobStateRunning,
				rivertype.JobStateRetryable,
				rivertype.JobStateScheduled,
			},
		},
		Queue:    worker_args.WorkstationConnectivityQueue,
		Metadata: []byte(workstationJobMetadata(ident)),
	}

	for _, host := range hosts {
		_, err = client.InsertTx(ctx, tx, &worker_args.WorkstationConnectJob{
			Ident: ident,
			Host:  host,
		}, insertOpts)
		if err != nil {
			return errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
		}
	}

	_, err = client.InsertTx(ctx, tx, &worker_args.WorkstationDisconnectJob{
		Ident: ident,
		Hosts: hosts,
	}, insertOpts)
	if err != nil {
		return errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	_, err = client.InsertTx(ctx, tx, &worker_args.WorkstationNotifyJob{
		Ident:     ident,
		RequestID: requestID,
		Hosts:     hosts,
	}, insertOpts)
	if err != nil {
		return errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("committing workstations worker transaction: %w", err))
	}

	return nil
}

func (s *workstationsQueue) GetWorkstationStartJob(ctx context.Context, id int64) (*service.WorkstationStartJob, error) {
	const op errs.Op = "workstationsQueue.GetWorkstationStartJob"

	client, err := NewClient(s.repo, s.config)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	j, err := client.JobGet(ctx, id)
	if err != nil {
		if errors.Is(err, river.ErrNotFound) {
			return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, err, service.ParamJob)
		}

		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	job, err := FromRiverJob(j, func(header service.JobHeader, args worker_args.WorkstationStart) *service.WorkstationStartJob {
		return &service.WorkstationStartJob{
			JobHeader: header,
			Ident:     args.Ident,
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err, service.ParamJob)
	}

	return job, nil
}

func (s *workstationsQueue) CreateWorkstationStartJob(ctx context.Context, ident string) (*service.WorkstationStartJob, error) {
	const op errs.Op = "workstationsQueue.CreateWorkstationStartJob"

	client, err := NewClient(s.repo, s.config)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	tx, err := s.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}
	defer tx.Rollback(ctx)

	insertOpts := &river.InsertOpts{
		MaxAttempts: 5,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 15 * time.Minute,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable,
				rivertype.JobStatePending,
				rivertype.JobStateRunning,
				rivertype.JobStateRetryable,
				rivertype.JobStateScheduled,
			},
		},
		Queue:    worker_args.WorkstationQueue,
		Metadata: []byte(workstationJobMetadata(ident)),
	}

	raw, err := client.InsertTx(ctx, tx, &worker_args.WorkstationStart{
		Ident: ident,
	}, insertOpts)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	job, err := FromRiverJob(raw.Job, func(header service.JobHeader, args worker_args.WorkstationStart) *service.WorkstationStartJob {
		return &service.WorkstationStartJob{
			JobHeader: header,
			Ident:     args.Ident,
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err, service.ParamJob)
	}

	job.Duplicate = raw.UniqueSkippedAsDuplicate

	err = tx.Commit(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("committing workstations worker transaction: %w", err))
	}

	return job, nil
}

func (s *workstationsQueue) GetWorkstationStartJobsForUser(ctx context.Context, ident string) ([]*service.WorkstationStartJob, error) {
	const op errs.Op = "workstationsQueue.GetWorkstationStartJobsForUser"

	client, err := NewClient(s.repo, s.config)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	params := river.NewJobListParams().
		Queues(worker_args.WorkstationQueue).
		States(
			rivertype.JobStateAvailable,
			rivertype.JobStateRunning,
			rivertype.JobStateRetryable,
			rivertype.JobStateCompleted,
			rivertype.JobStateDiscarded,
		).
		Kinds(worker_args.WorkstationStartKind).
		Metadata(workstationJobMetadata(ident)).
		OrderBy("id", river.SortOrderDesc).
		First(5)

	raw, err := client.JobList(ctx, params)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	var jobs []*service.WorkstationStartJob
	for _, r := range raw.Jobs {
		job, err := FromRiverJob(r, func(header service.JobHeader, args worker_args.WorkstationStart) *service.WorkstationStartJob {
			return &service.WorkstationStartJob{
				JobHeader: header,
				Ident:     args.Ident,
			}
		})
		if err != nil {
			return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err)
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (s *workstationsQueue) GetWorkstationJobsForUser(ctx context.Context, ident string) ([]*service.WorkstationJob, error) {
	const op errs.Op = "workstationsQueue.GetRunningWorkstationJobsForUser"

	client, err := NewClient(s.repo, s.config)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	params := river.NewJobListParams().
		Queues(worker_args.WorkstationQueue).
		States(
			rivertype.JobStateAvailable,
			rivertype.JobStateRunning,
			rivertype.JobStateRetryable,
			rivertype.JobStateCompleted,
			rivertype.JobStateDiscarded,
		).
		Kinds(worker_args.WorkstationJobKind).
		Metadata(workstationJobMetadata(ident)).
		OrderBy("id", river.SortOrderDesc).
		First(10)

	raw, err := client.JobList(ctx, params)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	var jobs []*service.WorkstationJob
	for _, r := range raw.Jobs {
		job, err := FromRiverJob(r, func(header service.JobHeader, args worker_args.WorkstationJob) *service.WorkstationJob {
			return &service.WorkstationJob{
				JobHeader:      header,
				Name:           args.Name,
				Email:          args.Email,
				Ident:          args.Ident,
				MachineType:    args.MachineType,
				ContainerImage: args.ContainerImage,
			}
		})
		if err != nil {
			return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err)
		}

		jobs = append(jobs, job)
	}

	// We need at least two jobs to compare
	if len(jobs) < 2 {
		return jobs, nil
	}

	for i := 1; i < len(jobs); i++ {
		diff := JobDifference(jobs[i], jobs[i-1])
		jobs[i-1].Diff = diff
	}

	return jobs, nil
}

func (s *workstationsQueue) GetWorkstationJob(ctx context.Context, jobID int64) (*service.WorkstationJob, error) {
	const op errs.Op = "workstationsQueue.GetWorkstationJob"

	client, err := NewClient(s.repo, s.config)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	j, err := client.JobGet(ctx, jobID)
	if err != nil {
		if errors.Is(err, river.ErrNotFound) {
			return nil, errs.E(errs.NotExist, service.CodeTransactionalQueue, op, err, service.ParamJob)
		}

		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	job, err := FromRiverJob(j, func(header service.JobHeader, args worker_args.WorkstationJob) *service.WorkstationJob {
		return &service.WorkstationJob{
			JobHeader:      header,
			Name:           args.Name,
			Email:          args.Email,
			Ident:          args.Ident,
			MachineType:    args.MachineType,
			ContainerImage: args.ContainerImage,
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err, service.ParamJob)
	}

	return job, nil
}

func (s *workstationsQueue) CreateWorkstationJob(ctx context.Context, opts *service.WorkstationJobOpts) (*service.WorkstationJob, error) {
	const op errs.Op = "workstationsQueue.CreateWorkstationJob"

	client, err := NewClient(s.repo, s.config)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	tx, err := s.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("starting workstations worker transaction: %w", err))
	}

	insertOpts := &river.InsertOpts{
		MaxAttempts: 5,
		UniqueOpts: river.UniqueOpts{
			ByArgs:   true,
			ByPeriod: 20 * time.Minute,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable,
				rivertype.JobStatePending,
				rivertype.JobStateRunning,
				rivertype.JobStateRetryable,
				rivertype.JobStateScheduled,
			},
		},
		Queue:    worker_args.WorkstationQueue,
		Metadata: []byte(workstationJobMetadata(opts.User.Ident)),
	}

	raw, err := client.InsertTx(ctx, tx, &worker_args.WorkstationJob{
		Ident:          opts.User.Ident,
		Email:          opts.User.Email,
		Name:           opts.User.Name,
		MachineType:    opts.Input.MachineType,
		ContainerImage: opts.Input.ContainerImage,
	}, insertOpts)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, err)
	}

	job, err := FromRiverJob(raw.Job, func(header service.JobHeader, args worker_args.WorkstationJob) *service.WorkstationJob {
		return &service.WorkstationJob{
			JobHeader:      header,
			Name:           args.Name,
			Email:          args.Email,
			Ident:          args.Ident,
			MachineType:    args.MachineType,
			ContainerImage: args.ContainerImage,
		}
	})
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err, service.ParamJob)
	}

	job.Duplicate = raw.UniqueSkippedAsDuplicate

	err = tx.Commit(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeTransactionalQueue, op, fmt.Errorf("committing workstations worker transaction: %w", err))
	}

	return job, nil
}

func workstationJobMetadata(ident string) string {
	return fmt.Sprintf(`{"ident": "%s"}`, ident)
}

func JobDifference(a, b *service.WorkstationJob) map[string]*service.Diff {
	diff := map[string]*service.Diff{}

	if a.ContainerImage != b.ContainerImage {
		diff[service.WorkstationDiffContainerImage] = &service.Diff{
			Added:   []string{b.ContainerImage},
			Removed: []string{a.ContainerImage},
		}
	}

	if a.MachineType != b.MachineType {
		diff[service.WorkstationDiffMachineType] = &service.Diff{
			Added:   []string{b.MachineType},
			Removed: []string{a.MachineType},
		}
	}

	return diff
}

func NewWorkstationsQueue(config *riverpro.Config, repo *database.Repo) *workstationsQueue {
	return &workstationsQueue{
		config: config,
		repo:   repo,
	}
}
