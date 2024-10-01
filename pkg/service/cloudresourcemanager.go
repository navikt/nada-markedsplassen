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
