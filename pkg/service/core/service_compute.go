package core

import (
	"context"

	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.ComputeService = (*computeService)(nil)

type computeService struct {
	project    string
	policyName string

	computeAPI service.ComputeAPI
}

func (s *computeService) GetAllowedFirewallTags(ctx context.Context) ([]string, error) {
	const op errs.Op = "computeService.GetAllowedFirewallTags"

	frs, err := s.computeAPI.GetFirewallRulesForPolicy(ctx, s.policyName)
	if err != nil {
		return nil, errs.E(op, err)
	}

	tags := []string{}
	for _, fr := range frs {
		tags = append(tags, fr.SecureTags...)
	}

	return tags, nil
}

func NewComputeService(project, policyName string, computeAPI service.ComputeAPI) *computeService {
	return &computeService{
		project:    project,
		policyName: policyName,
		computeAPI: computeAPI,
	}
}
