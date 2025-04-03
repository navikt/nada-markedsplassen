package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"

	"riverqueue.com/riverpro"
	"riverqueue.com/riverpro/driver/riverpropgxv5"

	"github.com/navikt/nada-backend/pkg/worker/worker_args"
	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"

	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service"
)

type Workstation struct {
	river.WorkerDefaults[worker_args.WorkstationJob]

	service service.WorkstationsService
	repo    *database.Repo
}

func WorkstationConfig(log *zerolog.Logger, workers *river.Workers) *riverpro.Config {
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

	return &riverpro.Config{
		Config: river.Config{
			Queues: map[string]river.QueueConfig{
				worker_args.WorkstationQueue: {
					MaxWorkers: 10,
				},
				worker_args.WorkstationConnectivityQueue: {
					MaxWorkers: 10,
				},
			},
			Logger:     logger,
			Workers:    workers,
			JobTimeout: 10 * time.Minute,
		},
	}
}

func (w *Workstation) Work(ctx context.Context, job *river.Job[worker_args.WorkstationJob]) error {
	user := &service.User{
		Name:  job.Args.Name,
		Email: job.Args.Email,
		Ident: job.Args.Ident,
	}

	input := &service.WorkstationInput{
		MachineType:    job.Args.MachineType,
		ContainerImage: job.Args.ContainerImage,
	}

	_, err := w.service.EnsureWorkstation(ctx, user, input)
	if err != nil {
		return fmt.Errorf("ensuring workstation: %w", err)
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("starting workstation worker transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("completing workstations worker job: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("committing workstation worker transaction: %w", err)
	}

	return nil
}

type WorkstationStart struct {
	river.WorkerDefaults[worker_args.WorkstationStart]

	service service.WorkstationsService
	repo    *database.Repo
}

func (w *WorkstationStart) Work(ctx context.Context, job *river.Job[worker_args.WorkstationStart]) error {
	user := &service.User{
		Ident: job.Args.Ident,
	}

	err := w.service.StartWorkstation(ctx, user)
	if err != nil {
		return fmt.Errorf("starting workstation: %w", err)
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("workstation start worker transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("completing workstation start worker job: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("committing workstation start worker transaction: %w", err)
	}

	return nil
}

type WorkstationZonalTagBindings struct {
	river.WorkerDefaults[worker_args.WorkstationZonalTagBindingsJob]

	service service.WorkstationsService
	repo    *database.Repo
}

func (w *WorkstationZonalTagBindings) Work(ctx context.Context, job *river.Job[worker_args.WorkstationZonalTagBindingsJob]) error {
	err := w.service.UpdateWorkstationZonalTagBindingsForUser(ctx, job.Args.Ident, job.Args.RequestID, job.Args.Hosts)
	if err != nil {
		return fmt.Errorf("updating workstation zonal tag bindings for user: %w", err)
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("workstation zonal tag binding transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("completing workstation zonal tag binding job: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("committing workstation zonal tag binding transaction: %w", err)
	}

	return nil
}

type WorkstationConnect struct {
	river.WorkerDefaults[worker_args.WorkstationConnectJob]

	service service.WorkstationsService
	repo    *database.Repo
}

func (w *WorkstationConnect) Work(ctx context.Context, job *river.Job[worker_args.WorkstationConnectJob]) error {
	err := w.service.ConnectWorkstation(ctx, job.Args.Ident, job.Args.Host)
	if err != nil {
		return fmt.Errorf("connecting workstation: %w", err)
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("workstation connect transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("completing workstation connect job: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("committing workstation connect transaction: %w", err)
	}

	return nil
}

type WorkstationDisconnect struct {
	river.WorkerDefaults[worker_args.WorkstationDisconnectJob]

	service service.WorkstationsService
	repo    *database.Repo
}

func (w *WorkstationDisconnect) Work(ctx context.Context, job *river.Job[worker_args.WorkstationDisconnectJob]) error {
	err := w.service.DisconnectWorkstation(ctx, job.Args.Ident, job.Args.Hosts)
	if err != nil {
		return fmt.Errorf("disconnecting workstation: %w", err)
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("workstation disconnect transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("completing workstation disconnect job: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("committing workstation disconnect transaction: %w", err)
	}

	return nil
}

type WorkstationNotify struct {
	river.WorkerDefaults[worker_args.WorkstationNotifyJob]

	service service.WorkstationsService
	repo    *database.Repo
}

func (w *WorkstationNotify) Work(ctx context.Context, job *river.Job[worker_args.WorkstationNotifyJob]) error {
	err := w.service.NotifyWorkstation(ctx, job.Args.Ident, job.Args.RequestID, job.Args.Hosts)
	if err != nil {
		return fmt.Errorf("notifying workstation connect: %w", err)
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("workstation connect notify transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("completing workstation connect notify job: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("committing workstation connect notify transaction: %w", err)
	}

	return nil
}

func NewWorkstationWorker(config *riverpro.Config, service service.WorkstationsService, repo *database.Repo) (*riverpro.Client[pgx.Tx], error) {
	err := river.AddWorkerSafely[worker_args.WorkstationJob](config.Workers, &Workstation{
		WorkerDefaults: river.WorkerDefaults[worker_args.WorkstationJob]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return nil, fmt.Errorf("adding workstation worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.WorkstationStart](config.Workers, &WorkstationStart{
		WorkerDefaults: river.WorkerDefaults[worker_args.WorkstationStart]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return nil, fmt.Errorf("adding workstation start worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.WorkstationZonalTagBindingsJob](config.Workers, &WorkstationZonalTagBindings{
		WorkerDefaults: river.WorkerDefaults[worker_args.WorkstationZonalTagBindingsJob]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return nil, fmt.Errorf("adding workstation zonal tag binding worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.WorkstationConnectJob](config.Workers, &WorkstationConnect{
		WorkerDefaults: river.WorkerDefaults[worker_args.WorkstationConnectJob]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return nil, fmt.Errorf("adding workstation connect worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.WorkstationNotifyJob](config.Workers, &WorkstationNotify{
		WorkerDefaults: river.WorkerDefaults[worker_args.WorkstationNotifyJob]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return nil, fmt.Errorf("adding workstation connect notify worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.WorkstationDisconnectJob](config.Workers, &WorkstationDisconnect{
		WorkerDefaults: river.WorkerDefaults[worker_args.WorkstationDisconnectJob]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return nil, fmt.Errorf("adding workstation disconnect worker: %w", err)
	}

	client, err := riverpro.NewClient[pgx.Tx](riverpropgxv5.New(repo.GetDBX()), config)
	if err != nil {
		return nil, fmt.Errorf("creating river client: %w", err)
	}

	return client, nil
}
