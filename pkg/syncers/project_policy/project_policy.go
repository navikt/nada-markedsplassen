package project_policy

import (
	"context"
	"fmt"

	"github.com/navikt/nada-backend/pkg/cloudresourcemanager"

	"github.com/navikt/nada-backend/pkg/syncers"

	"github.com/rs/zerolog"
)

var _ syncers.Runner = &Runner{}

type Runner struct {
	gcpProject    string
	evaluateRoles []string
	ops           cloudresourcemanager.Operations
}

func (r *Runner) Name() string {
	return "GoogleProjectIamPolicyCleaner"
}

func (r *Runner) RunOnce(ctx context.Context, log zerolog.Logger) error {
	log.Info().Msg("Removing members from google iam project with deleted prefix...")

	err := r.ops.UpdateProjectIAMPolicyBindingsMembers(ctx, r.gcpProject, cloudresourcemanager.RemoveDeletedMembersWithRole(r.evaluateRoles, log))
	if err != nil {
		return fmt.Errorf("updating project policy bindings members: %w", err)
	}

	log.Info().Msg("Done cleaning up members.")

	return nil
}

func New(gcpProject string, evaluateRoles []string, ops cloudresourcemanager.Operations) *Runner {
	return &Runner{
		gcpProject:    gcpProject,
		evaluateRoles: evaluateRoles,
		ops:           ops,
	}
}
