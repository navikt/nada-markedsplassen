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

func addUserToWorkstation(member string) workstations.UpdateWorkstationIAMPolicyBindingsFn {
	return func(bindings []*workstations.Binding) []*workstations.Binding {
		for _, b := range bindings {
			if b.Role == service.WorkstationUserRole {
				for _, m := range b.Members {
					if m == member {
						return bindings
					}
				}

				b.Members = append(b.Members, member)
				return bindings
			}
		}

		return append(bindings, &workstations.Binding{
			Role:    service.WorkstationUserRole,
			Members: []string{member},
		})
	}
}

func (a *workstationsAPI) AddWorkstationUser(ctx context.Context, id *service.WorkstationIdentifier, email string) error {
	const op errs.Op = "workstationsAPI.AddWorkstationUser"

	err := a.ops.UpdateWorkstationIAMPolicyBindings(ctx,
		&workstations.WorkstationIdentifier{
			Slug:                  id.Slug,
			WorkstationConfigSlug: id.WorkstationConfigSlug,
		},
		addUserToWorkstation(fmt.Sprintf("user:%s", email)),
	)
	if err != nil {
		if errors.Is(err, workstations.ErrNotExist) {
			return errs.E(errs.NotExist, service.CodeGCPWorkstation, op, fmt.Errorf("workstation %s with config %s not found: %w", id.Slug, id.WorkstationConfigSlug, err), service.ParamWorkstation)
		}

		return errs.E(errs.IO, service.CodeGCPWorkstation, op, fmt.Errorf("adding user to workstation %s with config %s: %w", id.Slug, id.WorkstationConfigSlug, err))
	}

	return nil
}

func (a *workstationsAPI) StartWorkstation(ctx context.Context, id *service.WorkstationIdentifier) error {
	const op errs.Op = "workstationsAPI.StartWorkstation"

	err := a.ops.StartWorkstation(ctx, &workstations.WorkstationIdentifier{
		Slug:                  id.Slug,
		WorkstationConfigSlug: id.WorkstationConfigSlug,
	})
	if err != nil {
		if errors.Is(err, workstations.ErrNotExist) {
			return errs.E(errs.NotExist, service.CodeGCPWorkstation, op, fmt.Errorf("workstation %s with config %s not found: %w", id.Slug, id.WorkstationConfigSlug, err), service.ParamWorkstation)
		}

		return errs.E(errs.IO, service.CodeGCPWorkstation, op, fmt.Errorf("starting workstation %s with config %s: %w", id.Slug, id.WorkstationConfigSlug, err))
	}

	return nil
}

func (a *workstationsAPI) StopWorkstation(ctx context.Context, id *service.WorkstationIdentifier) error {
	const op errs.Op = "workstationsAPI.StopWorkstation"

	err := a.ops.StopWorkstation(ctx, &workstations.WorkstationIdentifier{
		Slug:                  id.Slug,
		WorkstationConfigSlug: id.WorkstationConfigSlug,
	})
	if err != nil {
		if errors.Is(err, workstations.ErrNotExist) {
			return errs.E(errs.NotExist, service.CodeGCPWorkstation, op, fmt.Errorf("workstation %s with config %s not found: %w", id.Slug, id.WorkstationConfigSlug, err), service.ParamWorkstation)
		}

		return errs.E(errs.IO, service.CodeGCPWorkstation, op, fmt.Errorf("stoping workstation %s with config %s: %w", id.Slug, id.WorkstationConfigSlug, err))
	}

	return nil
}

func (a *workstationsAPI) EnsureWorkstationWithConfig(ctx context.Context, opts *service.EnsureWorkstationOpts) (*service.WorkstationConfig, *service.Workstation, error) {
	const op errs.Op = "workstationsAPI.EnsureWorkstationWithConfig"

	err := opts.Config.Validate()
	if err != nil {
		return nil, nil, errs.E(errs.Invalid, service.CodeGCPWorkstation, op, err)
	}

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
			return nil, errs.E(errs.NotExist, service.CodeGCPWorkstation, op, fmt.Errorf("workstation config %s not found: %w", opts.Slug, err), service.ParamWorkstation)
		}

		return nil, errs.E(errs.IO, service.CodeGCPWorkstation, op, err)
	}

	return &service.WorkstationConfig{
		Slug:               c.Slug,
		FullyQualifiedName: c.FullyQualifiedName,
		DisplayName:        c.DisplayName,
		Annotations:        c.Annotations,
		Labels:             c.Labels,
		ServiceAccount:     c.ServiceAccount,
		CreateTime:         c.CreateTime,
		UpdateTime:         c.UpdateTime,
		IdleTimeout:        c.IdleTimeout,
		RunningTimeout:     c.RunningTimeout,
		ReplicaZones:       c.ReplicaZones,
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
		Annotations:         opts.Annotations,
		MachineType:         opts.MachineType,
		ServiceAccountEmail: opts.ServiceAccountEmail,
		SubjectEmail:        opts.SubjectEmail,
		Labels:              opts.Labels,
		Env:                 opts.Env,
		ContainerImage:      opts.ContainerImage,
	})
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeGCPWorkstation, op, err)
	}

	return &service.WorkstationConfig{
		Slug:               c.Slug,
		FullyQualifiedName: c.FullyQualifiedName,
		DisplayName:        c.DisplayName,
		Annotations:        c.Annotations,
		Labels:             c.Labels,
		CreateTime:         c.CreateTime,
		UpdateTime:         c.UpdateTime,
		IdleTimeout:        c.IdleTimeout,
		RunningTimeout:     c.RunningTimeout,
		ReplicaZones:       c.ReplicaZones,
		MachineType:        c.MachineType,
		ServiceAccount:     c.ServiceAccount,
		Image:              c.Image,
		Env:                c.Env,
	}, nil
}

func (a *workstationsAPI) UpdateWorkstationConfig(ctx context.Context, opts *service.WorkstationConfigUpdateOpts) (*service.WorkstationConfig, error) {
	const op errs.Op = "workstationsAPI.UpdateWorkstationConfig"

	c, err := a.ops.UpdateWorkstationConfig(ctx, &workstations.WorkstationConfigUpdateOpts{
		Slug:           opts.Slug,
		Annotations:    opts.Annotations,
		MachineType:    opts.MachineType,
		ContainerImage: opts.ContainerImage,
		Env:            opts.Env,
	})
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeGCPWorkstation, op, err)
	}

	return &service.WorkstationConfig{
		Slug:               c.Slug,
		FullyQualifiedName: c.FullyQualifiedName,
		DisplayName:        c.DisplayName,
		Annotations:        c.Annotations,
		Labels:             c.Labels,
		CreateTime:         c.CreateTime,
		UpdateTime:         c.UpdateTime,
		IdleTimeout:        c.IdleTimeout,
		RunningTimeout:     c.RunningTimeout,
		ReplicaZones:       c.ReplicaZones,
		MachineType:        c.MachineType,
		ServiceAccount:     c.ServiceAccount,
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
		return errs.E(errs.IO, service.CodeGCPWorkstation, op, err)
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
		return nil, errs.E(errs.IO, service.CodeGCPWorkstation, op, err)
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

func (a *workstationsAPI) GetWorkstation(ctx context.Context, id *service.WorkstationIdentifier) (*service.Workstation, error) {
	const op errs.Op = "workstationsAPI.GetWorkstation"

	w, err := a.ops.GetWorkstation(ctx, &workstations.WorkstationIdentifier{
		Slug:                  id.Slug,
		WorkstationConfigSlug: id.WorkstationConfigSlug,
	})
	if err != nil {
		if errors.Is(err, workstations.ErrNotExist) {
			return nil, errs.E(errs.NotExist, service.CodeGCPWorkstation, op, fmt.Errorf("workstation %s with config %s not found: %w", id.Slug, id.WorkstationConfigSlug, err), service.ParamWorkstation)
		}

		return nil, errs.E(errs.IO, service.CodeGCPWorkstation, op, err)
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
	_ = config
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
		Annotations:    opts.Annotations,
		ContainerImage: opts.ContainerImage,
		Env:            opts.Env,
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	return config, nil
}

func (a *workstationsAPI) ListWorkstationConfigs(ctx context.Context) ([]*service.WorkstationConfig, error) {
	const op errs.Op = "workstationsAPI.ListWorkstationConfigs"

	configs, err := a.ops.ListWorkstationConfigs(ctx)
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeGCPWorkstation, op, err)
	}

	var out []*service.WorkstationConfig
	for _, c := range configs {
		out = append(out, &service.WorkstationConfig{
			Slug:               c.Slug,
			FullyQualifiedName: c.FullyQualifiedName,
			DisplayName:        c.DisplayName,
			Annotations:        c.Annotations,
			Labels:             c.Labels,
			CreateTime:         c.CreateTime,
			UpdateTime:         c.UpdateTime,
			IdleTimeout:        c.IdleTimeout,
			RunningTimeout:     c.RunningTimeout,
			ReplicaZones:       c.ReplicaZones,
			MachineType:        c.MachineType,
			ServiceAccount:     c.ServiceAccount,
			Image:              c.Image,
			Env:                c.Env,
		})
	}

	return out, nil
}

func (a *workstationsAPI) ensureWorkstation(ctx context.Context, opts *service.WorkstationOpts) (*service.Workstation, error) {
	const op errs.Op = "workstationsAPI.ensureWorkstation"

	workstation, err := a.GetWorkstation(ctx, &service.WorkstationIdentifier{
		Slug:                  opts.Slug,
		WorkstationConfigSlug: opts.ConfigName,
	})
	if err != nil && !errors.Is(err, workstations.ErrNotExist) {
		return nil, errs.E(op, err)
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
