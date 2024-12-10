package gcp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/navikt/nada-backend/pkg/cloudresourcemanager"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"google.golang.org/api/googleapi"
)

var _ service.CloudResourceManagerAPI = &cloudResourceManagerAPI{}

type cloudResourceManagerAPI struct {
	ops cloudresourcemanager.Operations
}

func (c *cloudResourceManagerAPI) DeleteZonalTagBinding(ctx context.Context, zone, parentResource, tagValue string) error {
	const op errs.Op = "cloudResourceManagerAPI.DeleteZonalTagBinding"

	err := c.ops.DeleteZonalTagBinding(ctx, zone, parentResource, tagValue)
	if err != nil {
		if errors.Is(err, cloudresourcemanager.ErrNotFound) {
			return errs.E(errs.NotExist, service.CodeGCPCloudResourceManager, op, err)
		}

		return errs.E(errs.IO, service.CodeGCPCloudResourceManager, op, err)
	}

	return nil
}

func (c *cloudResourceManagerAPI) ListEffectiveTags(ctx context.Context, zone, parentResource string) ([]*service.EffectiveTag, error) {
	const op errs.Op = "cloudResourceManagerAPI.ListEffectiveTags"

	raw, err := c.ops.ListEffectiveTags(ctx, zone, parentResource)
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeGCPCloudResourceManager, op, err)
	}

	var tags []*service.EffectiveTag

	for _, t := range raw {
		tags = append(tags, &service.EffectiveTag{
			NamespacedTagKey:   t.NamespacedTagKey,
			NamespacedTagValue: t.NamespacedTagValue,
			TagKey:             t.TagKey,
			TagKeyParentName:   t.TagKeyParentName,
			TagValue:           t.TagValue,
		})
	}

	return tags, nil
}

func (c *cloudResourceManagerAPI) AddProjectIAMPolicyBinding(ctx context.Context, project string, binding *service.Binding) error {
	const op errs.Op = "cloudResourceManagerAPI.AddProjectIAMPolicyBinding"

	if err := binding.Validate(); err != nil {
		return errs.E(errs.Invalid, service.CodeGCPCloudResourceManager, op, err)
	}

	var err error
	for i := 0; i < 60; i++ {
		err := c.ops.AddProjectIAMPolicyBinding(ctx, project, &cloudresourcemanager.Binding{
			Role:    binding.Role,
			Members: binding.Members,
		})
		if err != nil {
			var gerr *googleapi.Error
			if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
				time.Sleep(time.Second)
				continue
			}
			return errs.E(errs.IO, service.CodeGCPCloudResourceManager, op, fmt.Errorf("adding binding (role: %s, members: %v): %w", binding.Role, binding.Members, err))
		}
		return nil
	}

	return errs.E(errs.IO, service.CodeGCPCloudResourceManager, op, fmt.Errorf("adding binding (role: %s, members: %v): %w", binding.Role, binding.Members, err))
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

	err := c.ops.CreateZonalTagBinding(ctx, zone, parentResource, tagNamespacedName)
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
