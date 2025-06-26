package gcp

import (
	"context"
	"errors"

	"github.com/navikt/nada-backend/pkg/computeengine"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

type computeAPI struct {
	client     computeengine.Operations
	gcpProject string
}

var _ service.ComputeAPI = &computeAPI{}

func (a *computeAPI) GetVirtualMachineByLabel(ctx context.Context, zones []string, label *service.Label) (*service.VirtualMachine, error) {
	const op errs.Op = "computeAPI.GetVirtualMachineByLabel"

	vms, err := a.GetVirtualMachinesByLabel(ctx, zones, label)
	if err != nil {
		return nil, errs.E(op, err)
	}

	if len(vms) == 0 {
		return nil, errs.E(errs.Other, service.CodeGCPComputeEngine, op, service.ErrNoVMs)
	}

	if len(vms) > 1 {
		return nil, errs.E(errs.Other, service.CodeGCPComputeEngine, op, service.ErrMultipleVMs)
	}

	return vms[0], nil
}

func (a *computeAPI) GetVirtualMachinesByLabel(ctx context.Context, zones []string, label *service.Label) ([]*service.VirtualMachine, error) {
	const op errs.Op = "computeAPI.GetVirtualMachinesByLabel"

	raw, err := a.client.GetVirtualMachinesByLabel(ctx, a.gcpProject, zones, &computeengine.Label{
		Key:   label.Key,
		Value: label.Value,
	})
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeGCPComputeEngine, op, err)
	}

	vms := make([]*service.VirtualMachine, len(raw))
	for i, r := range raw {
		vms[i] = &service.VirtualMachine{
			Name:               r.Name,
			ID:                 r.ID,
			Zone:               r.Zone,
			FullyQualifiedZone: r.FullyQualifiedZone,
			IPs:                r.IPs,
		}
	}

	return vms, nil
}

func (a *computeAPI) GetFirewallRulesForRegionalPolicy(ctx context.Context, project, region, name string) ([]*service.FirewallRule, error) {
	const op errs.Op = "computeAPI.GetFirewallRulesForRegionalPolicy"

	raw, err := a.client.GetFirewallRulesForRegionalPolicy(ctx, project, region, name)
	if err != nil {
		if errors.Is(err, computeengine.ErrNotExists) {
			return nil, errs.E(errs.NotExist, service.CodeGCPComputeEngine, op, err, service.ParamPolicy)
		}

		return nil, errs.E(errs.IO, service.CodeGCPComputeEngine, op, err)
	}

	frs := make([]*service.FirewallRule, len(raw))
	for i, r := range raw {
		frs[i] = &service.FirewallRule{
			Name:        r.Name,
			SecureTags:  r.SecureTags,
			Description: r.Description,
		}
	}

	return frs, nil
}

func NewComputeAPI(project string, client computeengine.Operations) *computeAPI {
	return &computeAPI{
		gcpProject: project,
		client:     client,
	}
}
