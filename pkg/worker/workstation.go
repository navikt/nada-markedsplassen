package worker

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"

	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/worker/worker_args"
	"riverqueue.com/riverpro"
)

type Workstation struct {
	river.WorkerDefaults[worker_args.WorkstationJob]

	service service.WorkstationsService
	repo    *database.Repo
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

func WorkstationAddWorkers(config *riverpro.Config, service service.WorkstationsService, repo *database.Repo) error {
	err := river.AddWorkerSafely[worker_args.WorkstationJob](config.Workers, &Workstation{
		WorkerDefaults: river.WorkerDefaults[worker_args.WorkstationJob]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return fmt.Errorf("adding workstation worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.WorkstationStart](config.Workers, &WorkstationStart{
		WorkerDefaults: river.WorkerDefaults[worker_args.WorkstationStart]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return fmt.Errorf("adding workstation start worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.WorkstationConnectJob](config.Workers, &WorkstationConnect{
		WorkerDefaults: river.WorkerDefaults[worker_args.WorkstationConnectJob]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return fmt.Errorf("adding workstation connect worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.WorkstationNotifyJob](config.Workers, &WorkstationNotify{
		WorkerDefaults: river.WorkerDefaults[worker_args.WorkstationNotifyJob]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return fmt.Errorf("adding workstation connect notify worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.WorkstationDisconnectJob](config.Workers, &WorkstationDisconnect{
		WorkerDefaults: river.WorkerDefaults[worker_args.WorkstationDisconnectJob]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return fmt.Errorf("adding workstation disconnect worker: %w", err)
	}

	return nil
}
