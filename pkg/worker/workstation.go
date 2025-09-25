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

	ws, err := w.service.GetWorkstation(ctx, user)
	if err != nil {
		return fmt.Errorf("getting workstation: %w", err)
	}

	switch ws.State {
	case service.Workstation_STATE_STARTING, service.Workstation_STATE_RUNNING:
		return nil
	case service.Workstation_STATE_STOPPING:
		return fmt.Errorf("workstation is still stopping, need to wait: %s", ws.Slug)
	case service.Workstation_STATE_STOPPED:
		break
	}

	err = w.service.StartWorkstation(ctx, user)
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

type WorkstationStop struct {
	river.WorkerDefaults[worker_args.WorkstationStop]

	service service.WorkstationsService
	repo    *database.Repo
}

func (w *WorkstationStop) Work(ctx context.Context, job *river.Job[worker_args.WorkstationStop]) error {
	user := &service.User{
		Ident: job.Args.Ident,
	}

	err := w.service.StopWorkstation(ctx, user, job.Args.RequestID)
	if err != nil {
		return fmt.Errorf("stopping workstation: %w", err)
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("workstation stop worker transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("completing workstation stop worker job: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("committing workstation stop worker transaction: %w", err)
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

type WorkstationResync struct {
	river.WorkerDefaults[worker_args.WorkstationResync]

	service service.WorkstationsService
	repo    *database.Repo
}

func (w *WorkstationResync) Work(ctx context.Context, job *river.Job[worker_args.WorkstationResync]) error {
	user := &service.User{
		Ident: job.Args.Ident,
	}

	ws, err := w.service.GetWorkstation(ctx, user)
	if err != nil {
		return fmt.Errorf("getting workstation: %w", err)
	}

	user.Name = ws.DisplayName

	userEmail, ok := ws.Config.Env["WORKSTATION_USER_EMAIL"]
	if !ok {
		return fmt.Errorf("resyncing workstation, unable to retrieve user email for ident %s", user.Ident)
	}
	user.Email = userEmail

	input := &service.WorkstationInput{
		MachineType:    ws.Config.MachineType,
		ContainerImage: ws.Config.Image,
	}

	_, err = w.service.EnsureWorkstation(ctx, user, input)
	if err != nil {
		return fmt.Errorf("resyncing workstation: %w", err)
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("workstation resync worker transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("completing workstation resync worker job: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("committing workstation resync worker transaction: %w", err)
	}

	return nil
}

type ConfigWorkstationSSH struct {
	river.WorkerDefaults[worker_args.ConfigWorkstationSSH]

	service service.WorkstationsService
	repo    *database.Repo
}

func (w *ConfigWorkstationSSH) Work(ctx context.Context, job *river.Job[worker_args.ConfigWorkstationSSH]) error {
	err := w.service.ConfigWorkstationSSH(ctx, job.Args.Ident, job.Args.Allow)
	if err != nil {
		return fmt.Errorf("configuring workstation SSH: %w", err)
	}

	tx, err := w.repo.GetDBX().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("configuring workstation SSH transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return fmt.Errorf("configuring workstation SSH worker job: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("committing configuring workstation SSH worker transaction: %w", err)
	}

	return nil
}

type WorkstationEnsureURLList struct {
	river.WorkerDefaults[worker_args.WorkstationEnsureURLList]

	service service.WorkstationsService
	repo    *database.Repo
}

func (w *WorkstationEnsureURLList) Work(ctx context.Context, job *river.Job[worker_args.WorkstationEnsureURLList]) error {
	idents, err := w.service.GetWorkstationURLListUsers(ctx)
	if err != nil {
		return fmt.Errorf("getting all workstation idents: %w", err)
	}

	for _, ident := range idents {
		urlList, err := w.service.GetWorkstationActiveURLListForIdent(ctx, &service.User{Ident: ident.NavIdent})
		if err != nil {
			return fmt.Errorf("getting workstation active urllist for ident %s: %w", ident, err)
		}

		err = w.service.EnsureWorkstationURLList(ctx, &service.WorkstationActiveURLListForIdent{
			Slug:                 ident.NavIdent,
			URLList:              urlList.URLList,
			DisableGlobalURLList: urlList.DisableGlobalURLList,
		})
		if err != nil {
			return fmt.Errorf("ensuring workstation urllist for ident %s: %w", ident, err)
		}
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

	err = river.AddWorkerSafely[worker_args.WorkstationStop](config.Workers, &WorkstationStop{
		WorkerDefaults: river.WorkerDefaults[worker_args.WorkstationStop]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return fmt.Errorf("adding workstation stop worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.WorkstationResync](config.Workers, &WorkstationResync{
		WorkerDefaults: river.WorkerDefaults[worker_args.WorkstationResync]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return fmt.Errorf("adding workstation resync worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.ConfigWorkstationSSH](config.Workers, &ConfigWorkstationSSH{
		WorkerDefaults: river.WorkerDefaults[worker_args.ConfigWorkstationSSH]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return fmt.Errorf("adding configure workstation SSH worker: %w", err)
	}

	err = river.AddWorkerSafely[worker_args.WorkstationEnsureURLList](config.Workers, &WorkstationEnsureURLList{
		WorkerDefaults: river.WorkerDefaults[worker_args.WorkstationEnsureURLList]{},
		service:        service,
		repo:           repo,
	})
	if err != nil {
		return fmt.Errorf("adding workstation ensure urllist worker: %w", err)
	}

	return nil
}
