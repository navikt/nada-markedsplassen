package metabase_access

import (
	"context"
	"fmt"
	"strings"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/syncers"
	"github.com/rs/zerolog"
)

var _ syncers.Runner = &Runner{}

type Runner struct {
	api           service.MetabaseAPI
	mbStorage     service.RestrictedMetabaseStorage
	accessStorage service.AccessStorage
}

func New(api service.MetabaseAPI, mbStorage service.RestrictedMetabaseStorage, accessStorage service.AccessStorage) *Runner {
	return &Runner{
		api:           api,
		mbStorage:     mbStorage,
		accessStorage: accessStorage,
	}
}

func (r *Runner) Name() string {
	return "MetabaseAccessSyncer"
}

// RunOnce removes Metabase permission group members that no longer have an active
// Metabase access entry in the database.
func (r *Runner) RunOnce(ctx context.Context, log zerolog.Logger) error {
	allMeta, err := r.mbStorage.GetAllMetadata(ctx)
	if err != nil {
		return fmt.Errorf("getting all restricted metabase metadata: %w", err)
	}

	for _, meta := range allMeta {
		if meta.PermissionGroupID == nil || meta.SyncCompleted == nil {
			continue
		}

		if err := r.syncPermissionGroup(ctx, log, meta); err != nil {
			log.Error().Err(err).Msgf("syncing permission group for dataset %v", meta.DatasetID)
		}
	}

	return nil
}

func (r *Runner) syncPermissionGroup(ctx context.Context, log zerolog.Logger, meta *service.RestrictedMetabaseMetadata) error {
	groupMembers, err := r.api.GetPermissionGroup(ctx, *meta.PermissionGroupID)
	if err != nil {
		return fmt.Errorf("getting permission group %d: %w", *meta.PermissionGroupID, err)
	}

	activeAccesses, err := r.accessStorage.ListActiveAccessToDataset(ctx, meta.DatasetID)
	if err != nil {
		return fmt.Errorf("listing active access for dataset %v: %w", meta.DatasetID, err)
	}

	expectedEmails := make(map[string]struct{})
	for _, a := range activeAccesses {
		if a.Platform != service.AccessPlatformMetabase {
			continue
		}
		parts := strings.SplitN(a.Subject, ":", 2)
		if len(parts) == 2 && parts[0] == "user" {
			expectedEmails[parts[1]] = struct{}{}
		}
	}

	for _, member := range groupMembers {
		if _, ok := expectedEmails[member.Email]; ok {
			continue
		}
		log.Info().
			Str("email", member.Email).
			Str("dataset_id", meta.DatasetID.String()).
			Int("permission_group_id", *meta.PermissionGroupID).
			Msg("removing stale metabase permission group member")
		if err := r.api.RemovePermissionGroupMember(ctx, member.ID); err != nil {
			log.Error().Err(err).
				Str("email", member.Email).
				Str("dataset_id", meta.DatasetID.String()).
				Msg("removing permission group member")
		}
	}

	return nil
}
