package gcp

import (
	"context"
	"errors"

	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/workstations"
)

var _ service.WorkstationsAPI = &workstationsAPI{}

type workstationsAPI struct {
	ops workstations.Operations
}

func (a *workstationsAPI) EnsureWorkstationWithConfig(ctx context.Context, opts *service.EnsureWorkstationOpts) error {
	const op errs.Op = "workstationsAPI.EnsureWorkstationWithConfig"

	// FIXME: Do we need to stop and start the workstation before updating the configuration?
	err := a.ensureWorkstationConfig(ctx, &opts.Config)
	if err != nil {
		return errs.E(op, err)
	}

	err = a.ensureWorkstation(ctx, &opts.Workstation)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (a *workstationsAPI) CreateWorkstationConfig(ctx context.Context, opts *service.WorkstationConfigOpts) (*service.WorkstationConfig, error) {
	const op errs.Op = "workstationsAPI.CreateWorkstationConfig"

	w, err := a.ops.CreateWorkstationConfig(ctx, &workstations.WorkstationConfigOpts{
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
		Name:        w.Name,
		DisplayName: w.DisplayName,
	}, nil
}

func (a *workstationsAPI) UpdateWorkstationConfig(ctx context.Context, opts *service.WorkstationConfigUpdateOpts) (*service.WorkstationConfig, error) {
	const op errs.Op = "workstationsAPI.UpdateWorkstationConfig"

	w, err := a.ops.UpdateWorkstationConfig(ctx, &workstations.WorkstationConfigUpdateOpts{
		Slug:           opts.Slug,
		MachineType:    opts.MachineType,
		ContainerImage: opts.ContainerImage,
	})
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}

	return &service.WorkstationConfig{
		Name:        w.Name,
		DisplayName: w.DisplayName,
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
		Slug:        opts.Slug,
		DisplayName: opts.DisplayName,
		Labels:      opts.Labels,
		ConfigName:  opts.ConfigName,
	})
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}

	return &service.Workstation{
		Name: w.Name,
	}, nil
}

func (a *workstationsAPI) ensureWorkstationConfig(ctx context.Context, opts *service.WorkstationConfigOpts) error {
	const op errs.Op = "workstationsAPI.ensureWorkstationConfig"

	_, err := a.ops.GetWorkstationConfig(ctx, &workstations.WorkstationConfigGetOpts{
		Slug: opts.Slug,
	})
	if err != nil && !errors.Is(err, workstations.ErrNotExist) {
		return errs.E(errs.IO, op, err)
	}

	if errors.Is(err, workstations.ErrNotExist) {
		_, err := a.CreateWorkstationConfig(ctx, opts)
		if err != nil {
			return errs.E(op, err)
		}

		return nil
	}

	_, err = a.ops.UpdateWorkstationConfig(ctx, &workstations.WorkstationConfigUpdateOpts{
		Slug:           opts.Slug,
		MachineType:    opts.MachineType,
		ContainerImage: opts.ContainerImage,
	})
	if err != nil {
		return errs.E(errs.IO, op, err)
	}

	return nil
}

func (a *workstationsAPI) ensureWorkstation(ctx context.Context, opts *service.WorkstationOpts) error {
	const op errs.Op = "workstationsAPI.ensureWorkstation"

	_, err := a.ops.GetWorkstation(ctx, &workstations.WorkstationGetOpts{
		Slug:       opts.Slug,
		ConfigName: opts.ConfigName,
	})
	if err != nil && !errors.Is(err, workstations.ErrNotExist) {
		return errs.E(errs.IO, op, err)
	}

	if errors.Is(err, workstations.ErrNotExist) {
		_, err := a.CreateWorkstation(ctx, opts)
		if err != nil {
			return errs.E(op, err)
		}

		return nil
	}

	return nil
}
