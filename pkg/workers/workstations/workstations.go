package workstations

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/workers/workstations/args"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/riverqueue/river/rivershared/util/slogutil"
)

type WorkstationWorker struct {
	river.WorkerDefaults[args.WorkstationJob]

	service service.WorkstationsService
	repo    *database.Repo
}

func (w *WorkstationWorker) Work(ctx context.Context, job *river.Job[args.WorkstationJob]) error {
	user := &service.User{
		Name:  job.Args.Name,
		Email: job.Args.Email,
		Ident: job.Args.Ident,
	}

	input := &service.WorkstationInput{
		MachineType:     job.Args.MachineType,
		ContainerImage:  job.Args.ContainerImage,
		URLAllowList:    job.Args.URLAllowList,
		OnPremAllowList: job.Args.OnPremAllowList,
	}

	_, err := w.service.EnsureWorkstation(ctx, user, input)
	if err != nil {
		return err
	}

	tx, err := w.repo.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting workstations worker transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = river.JobCompleteTx[*riverdatabasesql.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("completing workstations worker job: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("committing workstations worker transaction: %w", err)
	}

	return nil
}

func New(workers *river.Workers, service service.WorkstationsService, repo *database.Repo) (*river.Client[*sql.Tx], error) {
	err := river.AddWorkerSafely(workers, &WorkstationWorker{
		WorkerDefaults: river.WorkerDefaults[args.WorkstationJob]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return nil, fmt.Errorf("adding workstation worker: %w", err)
	}

	// FIXME: Should pass in the config
	client, err := river.NewClient(riverdatabasesql.New(repo.GetDB()),
		&river.Config{
			Queues: map[string]river.QueueConfig{
				river.QueueDefault: {MaxWorkers: 5},
			},
			Logger:   slog.New(&slogutil.SlogMessageOnlyHandler{Level: slog.LevelDebug}),
			Workers:  workers,
			TestOnly: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("creating river client: %w", err)
	}

	return client, nil
}
