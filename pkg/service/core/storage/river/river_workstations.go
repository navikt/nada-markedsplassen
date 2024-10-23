package river

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/workers/workstations/args"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/riverqueue/river/rivertype"
)

var _ service.WorkstationsStorage = (*workstationsStorage)(nil)

type workstationsStorage struct {
	workers *river.Workers
	repo    *database.Repo
}

func fromRiverJob(job *rivertype.JobRow) (*service.WorkstationJob, error) {
	a := &args.WorkstationJob{}

	err := json.NewDecoder(bytes.NewReader(job.EncodedArgs)).Decode(a)
	if err != nil {
		return nil, fmt.Errorf("decoding workstation job args: %w", err)
	}

	var state service.WorkstationJobState
	var allErrs []string

	switch job.State {
	case rivertype.JobStateAvailable, rivertype.JobStateRunning, rivertype.JobStateRetryable:
		state = service.WorkstationJobStateRunning
	case rivertype.JobStateCompleted:
		state = service.WorkstationJobStateCompleted
	case rivertype.JobStateDiscarded:
		state = service.WorkstationJobStateFailed
		for _, e := range job.Errors {
			allErrs = append(allErrs, e.Error)
		}
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
		Errors:          allErrs,
	}, nil
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
		Queue: river.QueueDefault,
	}

	raw, err := client.InsertTx(ctx, tx, &args.WorkstationJob{
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
	// Create an insert-only client
	// - https://github.com/riverqueue/river?tab=readme-ov-file#insert-only-clients
	// FIXME: Should receive the config
	client, err := river.NewClient(riverdatabasesql.New(s.repo.GetDB()),
		&river.Config{
			TestOnly: true,
			Workers:  s.workers,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("creating river client: %w", err)
	}

	return client, nil
}

func NewWorkstationsStorage(workers *river.Workers, repo *database.Repo) *workstationsStorage {
	return &workstationsStorage{
		workers: workers,
		repo:    repo,
	}
}
