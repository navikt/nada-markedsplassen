package gcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/workstations"
)

var _ service.WorkstationsAPI = &workstationsAPI{}

type workstationsAPI struct {
	ops workstations.Operations
}

func (a *workstationsAPI) EnsureWorkstationWithConfig(ctx context.Context, opts *service.EnsureWorkstationOpts) (*service.WorkstationConfig, *service.Workstation, error) {
	const op errs.Op = "workstationsAPI.EnsureWorkstationWithConfig"

	// FIXME: Do we need to stop and start the workstation before updating the configuration?
	config, err := a.ensureWorkstationConfig(ctx, &opts.Config)
	if err != nil {
		return nil, nil, errs.E(op, err)
	}

	workstation, err := a.ensureWorkstation(ctx, &opts.Workstation)
	if err != nil {
		return nil, nil, errs.E(op, err)
	}

	return config, workstation, nil
}

func (a *workstationsAPI) GetWorkstationConfig(ctx context.Context, opts *service.WorkstationConfigGetOpts) (*service.WorkstationConfig, error) {
	const op errs.Op = "workstationsAPI.GetWorkstationConfig"

	c, err := a.ops.GetWorkstationConfig(ctx, &workstations.WorkstationConfigGetOpts{
		Slug: opts.Slug,
	})
	if err != nil {
		if errors.Is(err, workstations.ErrNotExist) {
			return nil, errs.E(errs.NotExist, op, fmt.Errorf("workstation config %s not found: %w", opts.Slug, err))
		}

		return nil, errs.E(errs.IO, op, err)
	}

	return &service.WorkstationConfig{
		Slug:               c.Slug,
		FullyQualifiedName: c.FullyQualifiedName,
		DisplayName:        c.DisplayName,
		Labels:             c.Labels,
		ServiceAccount:     c.ServiceAccount,
		CreateTime:         c.CreateTime,
		UpdateTime:         c.UpdateTime,
		IdleTimeout:        c.IdleTimeout,
		RunningTimeout:     c.RunningTimeout,
		MachineType:        c.MachineType,
		Image:              c.Image,
		Env:                c.Env,
	}, nil
}

func (a *workstationsAPI) CreateWorkstationConfig(ctx context.Context, opts *service.WorkstationConfigOpts) (*service.WorkstationConfig, error) {
	const op errs.Op = "workstationsAPI.CreateWorkstationConfig"

	c, err := a.ops.CreateWorkstationConfig(ctx, &workstations.WorkstationConfigOpts{
		Slug:                opts.Slug,
		DisplayName:         opts.DisplayName,
		Labels:              opts.Labels,
		MachineType:         opts.MachineType,
		ServiceAccountEmail: opts.ServiceAccountEmail,
		SubjectEmail:        opts.SubjectEmail,
		ContainerImage:      opts.ContainerImage,
	})
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}

	return &service.WorkstationConfig{
		Slug:               c.Slug,
		FullyQualifiedName: c.FullyQualifiedName,
		DisplayName:        c.DisplayName,
		Labels:             c.Labels,
		ServiceAccount:     c.ServiceAccount,
		CreateTime:         c.CreateTime,
		UpdateTime:         c.UpdateTime,
		IdleTimeout:        c.IdleTimeout,
		RunningTimeout:     c.RunningTimeout,
		MachineType:        c.MachineType,
		Image:              c.Image,
		Env:                c.Env,
	}, nil
}

func (a *workstationsAPI) UpdateWorkstationConfig(ctx context.Context, opts *service.WorkstationConfigUpdateOpts) (*service.WorkstationConfig, error) {
	const op errs.Op = "workstationsAPI.UpdateWorkstationConfig"

	c, err := a.ops.UpdateWorkstationConfig(ctx, &workstations.WorkstationConfigUpdateOpts{
		Slug:           opts.Slug,
		MachineType:    opts.MachineType,
		ContainerImage: opts.ContainerImage,
	})
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}

	return &service.WorkstationConfig{
		Slug:               c.Slug,
		FullyQualifiedName: c.FullyQualifiedName,
		DisplayName:        c.DisplayName,
		Labels:             c.Labels,
		ServiceAccount:     c.ServiceAccount,
		CreateTime:         c.CreateTime,
		UpdateTime:         c.UpdateTime,
		IdleTimeout:        c.IdleTimeout,
		RunningTimeout:     c.RunningTimeout,
		MachineType:        c.MachineType,
		Image:              c.Image,
		Env:                c.Env,
	}, nil
}

func (a *workstationsAPI) DeleteWorkstationConfig(ctx context.Context, opts *service.WorkstationConfigDeleteOpts) error {
	const op errs.Op = "workstationsAPI.DeleteWorkstationConfig"

	err := a.ops.DeleteWorkstationConfig(ctx, &workstations.WorkstationConfigDeleteOpts{
		Slug: opts.Slug,
	})
	if err != nil {
		return errs.E(errs.IO, op, err)
	}

	return nil
}

func (a *workstationsAPI) CreateWorkstation(ctx context.Context, opts *service.WorkstationOpts) (*service.Workstation, error) {
	const op errs.Op = "workstationsAPI.CreateWorkstation"

	w, err := a.ops.CreateWorkstation(ctx, &workstations.WorkstationOpts{
		Slug:                  opts.Slug,
		DisplayName:           opts.DisplayName,
		Labels:                opts.Labels,
		WorkstationConfigSlug: opts.ConfigName,
	})
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}

	return &service.Workstation{
		Slug:               w.Slug,
		FullyQualifiedName: w.FullyQualifiedName,
		DisplayName:        w.DisplayName,
		Reconciling:        w.Reconciling,
		CreateTime:         w.CreateTime,
		UpdateTime:         w.UpdateTime,
		StartTime:          w.StartTime,
		State:              service.WorkstationState(w.State),
		Host:               w.Host,
	}, nil
}

func (a *workstationsAPI) GetWorkstation(ctx context.Context, opts *service.WorkstationGetOpts) (*service.Workstation, error) {
	const op errs.Op = "workstationsAPI.GetWorkstation"

	w, err := a.ops.GetWorkstation(ctx, &workstations.WorkstationGetOpts{
		Slug:                  opts.Slug,
		WorkstationConfigSlug: opts.ConfigName,
	})
	if err != nil {
		if errors.Is(err, workstations.ErrNotExist) {
			return nil, errs.E(errs.NotExist, op, fmt.Errorf("workstation %s with config %s not found: %w", opts.Slug, opts.ConfigName, err))
		}

		return nil, errs.E(errs.IO, op, err)
	}

	return &service.Workstation{
		Slug:               w.Slug,
		FullyQualifiedName: w.FullyQualifiedName,
		DisplayName:        w.DisplayName,
		Reconciling:        w.Reconciling,
		CreateTime:         w.CreateTime,
		UpdateTime:         w.UpdateTime,
		StartTime:          w.StartTime,
		State:              service.WorkstationState(w.State),
		Host:               w.Host,
	}, nil
}

func (a *workstationsAPI) ensureWorkstationConfig(ctx context.Context, opts *service.WorkstationConfigOpts) (*service.WorkstationConfig, error) {
	const op errs.Op = "workstationsAPI.ensureWorkstationConfig"

	config, err := a.GetWorkstationConfig(ctx, &service.WorkstationConfigGetOpts{
		Slug: opts.Slug,
	})

	if errs.KindIs(errs.NotExist, err) {
		config, err = a.CreateWorkstationConfig(ctx, opts)
		if err != nil {
			return nil, errs.E(op, err)
		}

		return config, nil
	}

	config, err = a.UpdateWorkstationConfig(ctx, &service.WorkstationConfigUpdateOpts{
		Slug:           opts.Slug,
		MachineType:    opts.MachineType,
		ContainerImage: opts.ContainerImage,
	})
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}

	return config, nil
}

func (a *workstationsAPI) ensureWorkstation(ctx context.Context, opts *service.WorkstationOpts) (*service.Workstation, error) {
	const op errs.Op = "workstationsAPI.ensureWorkstation"

	workstation, err := a.GetWorkstation(ctx, &service.WorkstationGetOpts{
		Slug:       opts.Slug,
		ConfigName: opts.ConfigName,
	})
	if err != nil && !errors.Is(err, workstations.ErrNotExist) {
		return nil, errs.E(errs.IO, op, err)
	}

	if errors.Is(err, workstations.ErrNotExist) {
		workstation, err = a.CreateWorkstation(ctx, opts)
		if err != nil {
			return nil, errs.E(op, err)
		}

		return workstation, nil
	}

	return workstation, nil
}

func NewWorkstationsAPI(ops workstations.Operations) *workstationsAPI {
	return &workstationsAPI{
		ops: ops,
	}
}
