package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/navikt/nada-backend/pkg/worker/worker_args"
	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"

	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
)

type Workstation struct {
	river.WorkerDefaults[worker_args.WorkstationJob]

	service service.WorkstationsService
	repo    *database.Repo
}

func WorkstationConfig(log *zerolog.Logger, workers *river.Workers) *river.Config {
	var level slog.Level

	switch log.GetLevel() {
	case zerolog.DebugLevel:
		level = slog.LevelDebug
	case zerolog.InfoLevel:
		level = slog.LevelInfo
	case zerolog.WarnLevel:
		level = slog.LevelWarn
	case zerolog.ErrorLevel:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slogzerolog.Option{Level: level, Logger: log}.NewZerologHandler())

	return &river.Config{
		Queues: map[string]river.QueueConfig{
			worker_args.WorkstationQueue: {
				MaxWorkers: 5,
			},
		},
		Logger:  logger,
		Workers: workers,
	}
}

func (w *Workstation) Work(ctx context.Context, job *river.Job[worker_args.WorkstationJob]) error {
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
		return fmt.Errorf("ensuring workstation: %w", err)
	}

	tx, err := w.repo.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting workstation worker transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = river.JobCompleteTx[*riverdatabasesql.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("completing workstations worker job: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("committing workstation worker transaction: %w", err)
	}

	return nil
}

func NewWorkstationWorker(config *river.Config, service service.WorkstationsService, repo *database.Repo) (*river.Client[*sql.Tx], error) {
	err := river.AddWorkerSafely(config.Workers, &Workstation{
		WorkerDefaults: river.WorkerDefaults[worker_args.WorkstationJob]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return nil, fmt.Errorf("adding workstation worker: %w", err)
	}

	client, err := river.NewClient(riverdatabasesql.New(repo.GetDB()), config)
	if err != nil {
		return nil, fmt.Errorf("creating river client: %w", err)
	}

	return client, nil
}
