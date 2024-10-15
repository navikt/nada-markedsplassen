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

func (a *computeAPI) GetVirtualMachinesByLabel(ctx context.Context, zones []string, label *service.Label) ([]*service.VirtualMachine, error) {
	const op errs.Op = "computeAPI.GetVirtualMachinesByLabel"

	raw, err := a.client.GetVirtualMachinesByLabel(ctx, a.gcpProject, zones, &computeengine.Label{
		Key:   label.Key,
		Value: label.Value,
	})
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}

	vms := make([]*service.VirtualMachine, len(raw))
	for i, r := range raw {
		vms[i] = &service.VirtualMachine{
			Name:               r.Name,
			ID:                 r.ID,
			Zone:               r.Zone,
			FullyQualifiedZone: r.FullyQualifiedZone,
		}
	}

	return vms, nil
}

func (a *computeAPI) GetFirewallRulesForRegionalPolicy(ctx context.Context, project, region, name string) ([]*service.FirewallRule, error) {
	const op errs.Op = "computeAPI.GetFirewallRulesForRegionalPolicy"

	raw, err := a.client.GetFirewallRulesForRegionalPolicy(ctx, project, region, name)
	if err != nil {
		if errors.Is(err, computeengine.ErrNotExists) {
			return nil, errs.E(errs.NotExist, op, err)
		}

		return nil, errs.E(errs.IO, op, err)
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
