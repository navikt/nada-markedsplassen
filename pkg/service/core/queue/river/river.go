package river

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/riverqueue/river/rivertype"
	"golang.org/x/exp/maps"
	"riverqueue.com/riverpro"
	"riverqueue.com/riverpro/driver/riverpropgxv5"
)

func NewClient(repo *database.Repo, config *riverpro.Config) (*riverpro.Client[pgx.Tx], error) {
	client, err := riverpro.NewClient[pgx.Tx](riverpropgxv5.New(repo.GetDBX()), config)
	if err != nil {
		return nil, fmt.Errorf("creating river client: %w", err)
	}

	return client, nil
}

func FromRiverJobToHeader(job *rivertype.JobRow) service.JobHeader {
	var state service.JobState

	switch job.State {
	case rivertype.JobStateAvailable, rivertype.JobStateRunning, rivertype.JobStateRetryable:
		state = service.JobStateRunning
	case rivertype.JobStateCompleted:
		state = service.JobStateCompleted
	case rivertype.JobStateDiscarded:
		state = service.JobStateFailed
	}

	allErrs := map[string]struct{}{}

	for _, e := range job.Errors {
		allErrs[e.Error] = struct{}{}
	}

	return service.JobHeader{
		ID:        job.ID,
		StartTime: job.CreatedAt,
		EndTime:   job.FinalizedAt,
		State:     state,
		Duplicate: false,
		Errors:    maps.Keys(allErrs),
	}
}

func FromRiverJob[T, R any](job *rivertype.JobRow, processFn func(service.JobHeader, T) R) (R, error) {
	var args T
	var result R

	err := json.NewDecoder(bytes.NewReader(job.EncodedArgs)).Decode(&args)
	if err != nil {
		return result, fmt.Errorf("decoding job args: %w", err)
	}

	return processFn(FromRiverJobToHeader(job), args), nil
}
