package river

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"riverqueue.com/riverpro"
	"riverqueue.com/riverpro/driver/riverpropgxv5"

	"golang.org/x/exp/maps"

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

	client, err := s.newClient()
	if err != nil {
		return nil, errs.E(op, err)
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

	job, err := fromRiverDisconnectJob(raw.Jobs[0])
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err)
	}

	return job, nil
}

func (s *workstationsQueue) GetWorkstationConnectJobs(ctx context.Context, ident string) ([]*service.WorkstationConnectJob, error) {
	const op errs.Op = "workstationsQueue.GetWorkstationConnectJob"

	client, err := s.newClient()
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
		job, err := fromRiverConnectJob(r)
		if err != nil {
			return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err)
		}

		jobs[i] = job
	}

	return jobs, nil
}

func (s *workstationsQueue) GetWorkstationNotifyJob(ctx context.Context, ident string) (*service.WorkstationNotifyJob, error) {
	const op errs.Op = "workstationsQueue.GetWorkstationNotifyJob"

	client, err := s.newClient()
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

	job, err := fromRiverConnectNotifyJob(raw.Jobs[0])
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err)
	}

	return job, nil
}

func (s *workstationsQueue) CreateWorkstationConnectivityWorkflow(ctx context.Context, ident string, requestID string, hosts []string) error {
	const op errs.Op = "workstationsQueue.CreateWorkstationConnectivityWorkflow"

	client, err := s.newClient()
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

	client, err := s.newClient()
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

	job, err := fromRiverStartJob(j)
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err, service.ParamJob)
	}

	return job, nil
}

func (s *workstationsQueue) CreateWorkstationStartJob(ctx context.Context, ident string) (*service.WorkstationStartJob, error) {
	const op errs.Op = "workstationsQueue.CreateWorkstationStartJob"

	client, err := s.newClient()
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

	job, err := fromRiverStartJob(raw.Job)
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

	client, err := s.newClient()
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
		job, err := fromRiverStartJob(r)
		if err != nil {
			return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err, service.ParamJob)
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (s *workstationsQueue) GetWorkstationJobsForUser(ctx context.Context, ident string) ([]*service.WorkstationJob, error) {
	const op errs.Op = "workstationsQueue.GetRunningWorkstationJobsForUser"

	client, err := s.newClient()
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
		job, err := fromRiverJob(r)
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

	client, err := s.newClient()
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

	job, err := fromRiverJob(j)
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err, service.ParamJob)
	}

	return job, nil
}

func (s *workstationsQueue) CreateWorkstationJob(ctx context.Context, opts *service.WorkstationJobOpts) (*service.WorkstationJob, error) {
	const op errs.Op = "workstationsQueue.CreateWorkstationJob"

	client, err := s.newClient()
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

	job, err := fromRiverJob(raw.Job)
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

func (s *workstationsQueue) newClient() (*riverpro.Client[pgx.Tx], error) {
	client, err := riverpro.NewClient[pgx.Tx](riverpropgxv5.New(s.repo.GetDBX()), s.config)
	if err != nil {
		return nil, fmt.Errorf("creating river client: %w", err)
	}

	return client, nil
}

func fromRiverJobHeader(job *rivertype.JobRow) service.JobHeader {
	var state service.WorkstationJobState

	switch job.State {
	case rivertype.JobStateAvailable, rivertype.JobStateRunning, rivertype.JobStateRetryable:
		state = service.WorkstationJobStateRunning
	case rivertype.JobStateCompleted:
		state = service.WorkstationJobStateCompleted
	case rivertype.JobStateDiscarded:
		state = service.WorkstationJobStateFailed
	}

	allErrs := map[string]struct{}{}

	for _, e := range job.Errors {
		allErrs[e.Error] = struct{}{}
	}

	return service.JobHeader{
		ID:        job.ID,
		StartTime: job.CreatedAt,
		State:     state,
		Duplicate: false,
		Errors:    maps.Keys(allErrs),
	}
}

func fromRiverJob(job *rivertype.JobRow) (*service.WorkstationJob, error) {
	a := &worker_args.WorkstationJob{}

	err := json.NewDecoder(bytes.NewReader(job.EncodedArgs)).Decode(a)
	if err != nil {
		return nil, fmt.Errorf("decoding workstation job args: %w", err)
	}

	return &service.WorkstationJob{
		JobHeader:      fromRiverJobHeader(job),
		Name:           a.Name,
		Email:          a.Email,
		Ident:          a.Ident,
		MachineType:    a.MachineType,
		ContainerImage: a.ContainerImage,
		Diff:           nil,
	}, nil
}

func fromRiverStartJob(job *rivertype.JobRow) (*service.WorkstationStartJob, error) {
	a := &worker_args.WorkstationStart{}

	err := json.NewDecoder(bytes.NewReader(job.EncodedArgs)).Decode(a)
	if err != nil {
		return nil, fmt.Errorf("decoding workstation start job args: %w", err)
	}

	return &service.WorkstationStartJob{
		JobHeader: fromRiverJobHeader(job),
		Ident:     a.Ident,
	}, nil
}

func fromRiverConnectJob(job *rivertype.JobRow) (*service.WorkstationConnectJob, error) {
	a := &worker_args.WorkstationConnectJob{}

	err := json.NewDecoder(bytes.NewReader(job.EncodedArgs)).Decode(a)
	if err != nil {
		return nil, fmt.Errorf("decoding workstation connect job args: %w", err)
	}

	return &service.WorkstationConnectJob{
		JobHeader: fromRiverJobHeader(job),
		Ident:     a.Ident,
		Host:      a.Host,
	}, nil
}

func fromRiverDisconnectJob(job *rivertype.JobRow) (*service.WorkstationDisconnectJob, error) {
	a := &worker_args.WorkstationDisconnectJob{}

	err := json.NewDecoder(bytes.NewReader(job.EncodedArgs)).Decode(a)
	if err != nil {
		return nil, fmt.Errorf("decoding workstation disconnect job args: %w", err)
	}

	return &service.WorkstationDisconnectJob{
		JobHeader: fromRiverJobHeader(job),
		Ident:     a.Ident,
		Hosts:     a.Hosts,
	}, nil
}

func fromRiverConnectNotifyJob(job *rivertype.JobRow) (*service.WorkstationNotifyJob, error) {
	a := &worker_args.WorkstationNotifyJob{}

	err := json.NewDecoder(bytes.NewReader(job.EncodedArgs)).Decode(a)
	if err != nil {
		return nil, fmt.Errorf("decoding workstation connect notify job args: %w", err)
	}

	return &service.WorkstationNotifyJob{
		JobHeader: fromRiverJobHeader(job),
		Ident:     a.Ident,
		RequestID: a.RequestID,
		Hosts:     a.Hosts,
	}, nil
}

func NewWorkstationsQueue(config *riverpro.Config, repo *database.Repo) *workstationsQueue {
	return &workstationsQueue{
		config: config,
		repo:   repo,
	}
}
