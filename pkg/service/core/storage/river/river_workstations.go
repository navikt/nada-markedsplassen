package river

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/exp/maps"
	"time"

	"github.com/navikt/nada-backend/pkg/worker/worker_args"

	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/riverqueue/river/rivertype"
)

var _ service.WorkstationsStorage = (*workstationsStorage)(nil)

type workstationsStorage struct {
	repo   *database.Repo
	config *river.Config
}

func workstationJobMetadata(ident string) string {
	return fmt.Sprintf(`{"ident": "%s"}`, ident)
}

func (s *workstationsStorage) GetWorkstationJobsForUser(ctx context.Context, ident string) ([]*service.WorkstationJob, error) {
	const op errs.Op = "workstationsStorage.GetRunningWorkstationJobsForUser"

	client, err := s.newClient()
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
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
		Metadata(workstationJobMetadata(ident))

	raw, err := client.JobList(ctx, params)
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}

	var jobs []*service.WorkstationJob
	for _, r := range raw.Jobs {
		job, err := fromRiverJob(r)
		if err != nil {
			return nil, errs.E(errs.Internal, op, err)
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (s *workstationsStorage) GetWorkstationJob(ctx context.Context, jobID int64) (*service.WorkstationJob, error) {
	const op errs.Op = "workstationsStorage.GetWorkstationJob"

	client, err := s.newClient()
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}

	j, err := client.JobGet(ctx, jobID)
	if err != nil {
		if errors.Is(err, river.ErrNotFound) {
			return nil, errs.E(errs.NotExist, op, err)
		}

		return nil, errs.E(errs.IO, op, err)
	}

	job, err := fromRiverJob(j)
	if err != nil {
		return nil, errs.E(errs.Internal, op, err)
	}

	return job, nil
}

func (s *workstationsStorage) CreateWorkstationJob(ctx context.Context, opts *service.WorkstationJobOpts) (*service.WorkstationJob, error) {
	const op errs.Op = "workstationsStorage.CreateWorkstationJob"

	client, err := s.newClient()
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}

	tx, err := s.repo.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return nil, errs.E(errs.Database, op, fmt.Errorf("starting workstations worker transaction: %w", err))
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
		Ident:           opts.User.Ident,
		Email:           opts.User.Email,
		Name:            opts.User.Name,
		MachineType:     opts.Input.MachineType,
		ContainerImage:  opts.Input.ContainerImage,
		URLAllowList:    opts.Input.URLAllowList,
		OnPremAllowList: opts.Input.OnPremAllowList,
	}, insertOpts)
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}

	job, err := fromRiverJob(raw.Job)
	if err != nil {
		return nil, errs.E(errs.Internal, op, err)
	}

	job.Duplicate = raw.UniqueSkippedAsDuplicate

	err = tx.Commit()
	if err != nil {
		return nil, errs.E(errs.Database, op, fmt.Errorf("committing workstations worker transaction: %w", err))
	}

	return job, nil
}

func (s *workstationsStorage) newClient() (*river.Client[*sql.Tx], error) {
	client, err := river.NewClient(riverdatabasesql.New(s.repo.GetDB()), s.config)
	if err != nil {
		return nil, fmt.Errorf("creating river client: %w", err)
	}

	return client, nil
}

func fromRiverJob(job *rivertype.JobRow) (*service.WorkstationJob, error) {
	a := &worker_args.WorkstationJob{}

	err := json.NewDecoder(bytes.NewReader(job.EncodedArgs)).Decode(a)
	if err != nil {
		return nil, fmt.Errorf("decoding workstation job args: %w", err)
	}

	var state service.WorkstationJobState

	switch job.State {
	case rivertype.JobStateAvailable, rivertype.JobStateRunning, rivertype.JobStateRetryable:
		state = service.WorkstationJobStateRunning
	case rivertype.JobStateCompleted:
		state = service.WorkstationJobStateCompleted
	case rivertype.JobStateDiscarded:
		state = service.WorkstationJobStateFailed
	}

	var allErrs map[string]struct{}

	for _, e := range job.Errors {
		allErrs[e.Error] = struct{}{}
	}

	return &service.WorkstationJob{
		ID:              job.ID,
		Name:            a.Name,
		Email:           a.Email,
		Ident:           a.Ident,
		MachineType:     a.MachineType,
		ContainerImage:  a.ContainerImage,
		URLAllowList:    a.URLAllowList,
		OnPremAllowList: a.OnPremAllowList,
		StartTime:       job.CreatedAt,
		State:           state,
		Duplicate:       false,
		Errors:          maps.Keys(allErrs),
	}, nil
}

func NewWorkstationsStorage(config *river.Config, repo *database.Repo) *workstationsStorage {
	return &workstationsStorage{
		config: config,
		repo:   repo,
	}
}
