package project_policy

import (
	"context"
	"fmt"
	"time"

	"github.com/navikt/nada-backend/pkg/leaderelection"
	"github.com/navikt/nada-backend/pkg/sa"

	"github.com/rs/zerolog"
)

type Syncer struct {
	gcpProject    string
	evaluateRoles []string
	ops           sa.Operations
	log           zerolog.Logger
}

func New(gcpProject string, evaluateRoles []string, ops sa.Operations, log zerolog.Logger) *Syncer {
	return &Syncer{
		gcpProject:    gcpProject,
		evaluateRoles: evaluateRoles,
		ops:           ops,
		log:           log,
	}
}

func (s *Syncer) Run(ctx context.Context, frequency time.Duration, initialDelaySec int) {
	isLeader, err := leaderelection.IsLeader()
	if err != nil {
		s.log.Error().Err(err).Msg("checking leader status")

		return
	}

	if isLeader {
		// Delay a little before starting
		time.Sleep(time.Duration(initialDelaySec) * time.Second)

		s.log.Info().Fields(map[string]interface{}{
			"project": s.gcpProject,
			"roles":   s.evaluateRoles,
		}).Msg("Starting google project iam policy cleaner")

		err := s.RunOnce(ctx)
		if err != nil {
			s.log.Error().Err(err).Fields(map[string]interface{}{
				"project": s.gcpProject,
			}).Msg("running google project iam policy cleaner")
		}
	}

	ticker := time.NewTicker(frequency)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := s.RunOnce(ctx)
			if err != nil {
				s.log.Error().Err(err).Fields(map[string]interface{}{
					"project": s.gcpProject,
				}).Msg("running google project iam policy cleaner")
			}
		}
	}
}

func (s *Syncer) RunOnce(ctx context.Context) error {
	isLeader, err := leaderelection.IsLeader()
	if err != nil {
		return fmt.Errorf("checking leader status: %w", err)
	}

	if !isLeader {
		s.log.Info().Msg("not leader, skipping cleaner")

		return nil
	}

	s.log.Info().Msg("Removing members from google iam project with deleted prefix...")

	err = s.ops.UpdateProjectPolicyBindingsMembers(ctx, s.gcpProject, sa.RemoveDeletedMembersWithRole(s.evaluateRoles, s.log))
	if err != nil {
		return fmt.Errorf("updating project policy bindings members: %w", err)
	}

	s.log.Info().Msg("Done cleaning up members.")

	return nil
}
