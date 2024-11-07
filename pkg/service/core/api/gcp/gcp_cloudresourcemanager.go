package gcp

import (
	"context"
	"fmt"

	"github.com/navikt/nada-backend/pkg/cloudresourcemanager"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.CloudResourceManagerAPI = &cloudResourceManagerAPI{}

type cloudResourceManagerAPI struct {
	ops cloudresourcemanager.Operations
}

const (
	tagBindingMaxNumRetries     = 5
	tagBindingRetryDelaySeconds = 3
)

func (c *cloudResourceManagerAPI) AddProjectIAMPolicyBinding(ctx context.Context, project string, binding *service.Binding) error {
	const op errs.Op = "cloudResourceManagerAPI.AddProjectIAMPolicyBinding"

	if err := binding.Validate(); err != nil {
		return errs.E(errs.Invalid, service.CodeGCPCloudResourceManager, op, err)
	}

	err := c.ops.AddProjectIAMPolicyBinding(ctx, project, &cloudresourcemanager.Binding{
		Role:    binding.Role,
		Members: binding.Members,
	})
	if err != nil {
		return errs.E(errs.IO, service.CodeGCPCloudResourceManager, op, fmt.Errorf("adding binding (role: %s, members: %v): %w", binding.Role, binding.Members, err))
	}

	return nil
}

func (c *cloudResourceManagerAPI) RemoveProjectIAMPolicyBindingMember(ctx context.Context, project string, member string) error {
	const op errs.Op = "cloudResourceManagerAPI.RemoveProjectIAMPolicyBindingMember"

	err := c.ops.RemoveProjectIAMPolicyBindingMember(ctx, project, member)
	if err != nil {
		return errs.E(errs.IO, service.CodeGCPCloudResourceManager, op, err)
	}

	return nil
}

func (c *cloudResourceManagerAPI) RemoveProjectIAMPolicyBindingMemberForRole(ctx context.Context, project, role, member string) error {
	const op errs.Op = "cloudResourceManagerAPI.RemoveProjectIAMPolicyBindingMemberForRole"

	err := c.ops.RemoveProjectIAMPolicyBindingMemberForRole(ctx, project, role, member)
	if err != nil {
		return errs.E(errs.IO, service.CodeGCPCloudResourceManager, op, err)
	}

	return nil
}

func (c *cloudResourceManagerAPI) ListProjectIAMPolicyBindings(ctx context.Context, project, member string) ([]*service.Binding, error) {
	const op errs.Op = "cloudResourceManagerAPI.ListProjectIAMPolicyBindings"

	raw, err := c.ops.ListProjectIAMPolicyBindings(ctx, project, member)
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeGCPCloudResourceManager, op, err)
	}

	var bindings []*service.Binding

	for _, b := range raw {
		bindings = append(bindings, &service.Binding{
			Role:    b.Role,
			Members: b.Members,
		})
	}

	return bindings, nil
}

func (c *cloudResourceManagerAPI) CreateZonalTagBinding(ctx context.Context, zone, parentResource, tagNamespacedName string) error {
	const op errs.Op = "cloudResourceManagerAPI.CreateZonalTagBinding"

	err := c.ops.CreateZonalTagBindingWithRetries(ctx, zone, parentResource, tagNamespacedName, tagBindingMaxNumRetries, tagBindingRetryDelaySeconds)
	if err != nil {
		return errs.E(errs.IO, service.CodeGCPCloudResourceManager, op, err)
	}

	return nil
}

func NewCloudResourceManagerAPI(ops cloudresourcemanager.Operations) *cloudResourceManagerAPI {
	return &cloudResourceManagerAPI{
		ops: ops,
	}
}
