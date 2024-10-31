package river

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/exp/maps"
	"slices"
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

func (s *workstationsStorage) CreateWorkstationStartJob(ctx context.Context, ident string) (*service.WorkstationStartJob, error) {
	const op errs.Op = "workstationsStorage.CreateWorkstationStartJob"

	client, err := s.newClient()
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}

	tx, err := s.repo.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}
	defer tx.Rollback()

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
		return nil, errs.E(errs.IO, op, err)
	}

	job, err := fromRiverStartJob(raw.Job)
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

func (s *workstationsStorage) GetWorkstationStartJobsForUser(ctx context.Context, ident string) ([]*service.WorkstationStartJob, error) {
	const op errs.Op = "workstationsStorage.GetWorkstationStartJobsForUser"

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
		Kinds(worker_args.WorkstationStartKind).
		Metadata(workstationJobMetadata(ident)).
		OrderBy("id", river.SortOrderDesc).
		First(5)

	raw, err := client.JobList(ctx, params)
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}

	var jobs []*service.WorkstationStartJob
	for _, r := range raw.Jobs {
		job, err := fromRiverStartJob(r)
		if err != nil {
			return nil, errs.E(errs.Internal, op, err)
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
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
		Metadata(workstationJobMetadata(ident)).
		OrderBy("id", river.SortOrderDesc).
		First(10)

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

func workstationJobMetadata(ident string) string {
	return fmt.Sprintf(`{"ident": "%s"}`, ident)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func JobDifference(a, b *service.WorkstationJob) map[string]*service.Diff {
	diff := map[string]*service.Diff{}

	if a.ContainerImage != b.ContainerImage {
		diff[service.WorkstationDiffContainerImage] = &service.Diff{
			Value: b.ContainerImage,
		}
	}

	if a.MachineType != b.MachineType {
		diff[service.WorkstationDiffMachineType] = &service.Diff{
			Value: b.MachineType,
		}
	}

	if !slices.Equal(a.URLAllowList, b.URLAllowList) {
		d := &service.Diff{}

		// Find elements that are new in b and add them to added
		added := []string{}
		for _, url := range b.URLAllowList {
			if !contains(a.URLAllowList, url) {
				added = append(added, url)
			}
		}
		if len(added) > 0 {
			d.Added = added
		}

		// Find elements that are in a but not in b and add them to removed
		removed := []string{}
		for _, url := range a.URLAllowList {
			if !contains(b.URLAllowList, url) {
				removed = append(removed, url)
			}
		}
		if len(removed) > 0 {
			d.Removed = removed
		}

		diff[service.WorkstationDiffURLAllowList] = d
	}

	if !slices.Equal(a.OnPremAllowList, b.OnPremAllowList) {
		d := &service.Diff{}

		// Find elements that are new in b and add them to added
		added := []string{}
		for _, url := range b.OnPremAllowList {
			if !contains(a.OnPremAllowList, url) {
				added = append(added, url)
			}
		}
		if len(added) > 0 {
			d.Added = added
		}

		// Find elements that are in a but not in b and add them to removed
		removed := []string{}
		for _, url := range a.OnPremAllowList {
			if !contains(b.OnPremAllowList, url) {
				removed = append(removed, url)
			}
		}
		if len(removed) > 0 {
			d.Removed = removed
		}

		diff[service.WorkstationDiffOnPremAllowList] = d
	}

	return diff
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

	allErrs := map[string]struct{}{}

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

func fromRiverStartJob(job *rivertype.JobRow) (*service.WorkstationStartJob, error) {
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

	allErrs := map[string]struct{}{}

	for _, e := range job.Errors {
		allErrs[e.Error] = struct{}{}
	}

	return &service.WorkstationStartJob{
		ID:        job.ID,
		Ident:     a.Ident,
		StartTime: job.CreatedAt,
		State:     state,
		Duplicate: false,
		Errors:    maps.Keys(allErrs),
	}, nil
}

func NewWorkstationsStorage(config *river.Config, repo *database.Repo) *workstationsStorage {
	return &workstationsStorage{
		config: config,
		repo:   repo,
	}
}
