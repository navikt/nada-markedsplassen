package metabase_mapper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/navikt/nada-backend/pkg/syncers"

	"github.com/navikt/nada-backend/pkg/errs"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
)

var _ syncers.Runner = &Runner{}

type Work struct {
	DatasetID uuid.UUID
	Services  []string
}

type Runner struct {
	queue                    chan Work
	mappingDeadlineSec       int
	metabaseService          service.MetabaseService
	thirdPartyMappingStorage service.ThirdPartyMappingStorage
	log                      zerolog.Logger
}

func New(
	metabaseService service.MetabaseService,
	thirdPartyMappingStorage service.ThirdPartyMappingStorage,
	mappingDeadlineSec int,
	queue chan Work,
	log zerolog.Logger,
) *Runner {
	return &Runner{
		mappingDeadlineSec:       mappingDeadlineSec,
		metabaseService:          metabaseService,
		queue:                    queue,
		thirdPartyMappingStorage: thirdPartyMappingStorage,
		log:                      log,
	}
}

func (r *Runner) Name() string {
	return "MetabaseMapper"
}

func (r *Runner) ProcessQueue(ctx context.Context) {
	r.log.Info().Msg("Starting metabase mapper queue processor")

	for {
		select {
		case work := <-r.queue:
			r.MapDataset(ctx, work.DatasetID, work.Services)
		case <-ctx.Done():
			r.log.Info().Msg("Shutting down metabase mapper queue processor")
			return
		}
	}
}

func (r *Runner) RunOnce(ctx context.Context, _ zerolog.Logger) error {
	add, err := r.thirdPartyMappingStorage.GetAddMetabaseDatasetMappings(ctx)
	if err != nil {
		return fmt.Errorf("getting add metabase mappings: %w", err)
	}

	for _, datasetID := range add {
		r.queue <- Work{
			DatasetID: datasetID,
			Services: []string{
				service.MappingServiceMetabase,
			},
		}
	}

	remove, err := r.thirdPartyMappingStorage.GetRemoveMetabaseDatasetMappings(ctx)
	if err != nil {
		return fmt.Errorf("getting remove metabase mappings: %w", err)
	}

	for _, datasetID := range remove {
		r.queue <- Work{
			DatasetID: datasetID,
			Services:  []string{},
		}
	}

	return nil
}

func (r *Runner) MapDataset(ctx context.Context, datasetID uuid.UUID, services []string) {
	deadline := time.Duration(r.mappingDeadlineSec) * time.Second

	r.log.Info().Fields(map[string]interface{}{
		"dataset_id":       datasetID.String(),
		"services":         strings.Join(services, ","),
		"deadline_seconds": deadline.Seconds(),
	}).Msg("mapping dataset")

	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.mappingDeadlineSec)*time.Second)
	defer cancel()

	err := r.metabaseService.MapDataset(ctx, datasetID, services)
	if err != nil {
		r.log.Error().Err(err).Fields(map[string]interface{}{
			"dataset_id": datasetID.String(),
			"services":   services,
			"stack":      errs.OpStack(err),
		}).Msg("mapping dataset")
	}
}
