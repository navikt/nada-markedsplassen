package service

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type CloudResourceManagerAPI interface {
	// AddProjectIAMPolicyBinding adds a binding to a project IAM policy.
	AddProjectIAMPolicyBinding(ctx context.Context, project string, binding *Binding) error

	// RemoveProjectIAMPolicyBindingMember removes a member from all project IAM policy bindings.
	RemoveProjectIAMPolicyBindingMember(ctx context.Context, project string, member string) error

	// RemoveProjectIAMPolicyBindingMemberForRole removes a member from a specific role in all project IAM policy bindings.
	RemoveProjectIAMPolicyBindingMemberForRole(ctx context.Context, project, role, member string) error

	// ListProjectIAMPolicyBindings returns all project IAM policy bindings for a specific member.
	ListProjectIAMPolicyBindings(ctx context.Context, project, member string) ([]*Binding, error)

	// CreateZonalTagBinding creates a new tag binding
	CreateZonalTagBinding(ctx context.Context, zone, parentResource, tagNamespacedName string) error

	// DeleteZonalTagBinding deletes a tag binding
	DeleteZonalTagBinding(ctx context.Context, zone, parentResource, tagValue string) error

	// ListEffectiveTags returns all effective tags for a specific resource
	ListEffectiveTags(ctx context.Context, zone, parentResource string) ([]*EffectiveTag, error)
}

type EffectiveTags struct {
	Tags []*EffectiveTag `json:"tags,omitempty"`
}

type EffectiveTag struct {
	// NamespacedTagKey: The namespaced name of the TagKey. Can be in the form
	// `{organization_id}/{tag_key_short_name}` or
	// `{project_id}/{tag_key_short_name}` or
	// `{project_number}/{tag_key_short_name}`.
	NamespacedTagKey string `json:"namespacedTagKey,omitempty"`

	// NamespacedTagValue: The namespaced name of the TagValue. Can be in the form
	// `{organization_id}/{tag_key_short_name}/{tag_value_short_name}` or
	// `{project_id}/{tag_key_short_name}/{tag_value_short_name}` or
	// `{project_number}/{tag_key_short_name}/{tag_value_short_name}`.
	NamespacedTagValue string `json:"namespacedTagValue,omitempty"`

	// TagKey: The name of the TagKey, in the format `tagKeys/{id}`, such as
	// `tagKeys/123`.
	TagKey string `json:"tagKey,omitempty"`

	// TagKeyParentName: The parent name of the tag key. Must be in the format
	// `organizations/{organization_id}` or `projects/{project_number}`
	TagKeyParentName string `json:"tagKeyParentName,omitempty"`

	// TagValue: Resource name for TagValue in the format `tagValues/456`.
	TagValue string `json:"tagValue,omitempty"`
}

type Binding struct {
	Role    string
	Members []string
}

func (b Binding) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Role, validation.Required),
		validation.Field(&b.Members, validation.Required),
	)
}
