package empty_stories

import (
	"context"
	"fmt"
	"time"

	"github.com/navikt/nada-backend/pkg/syncers"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
)

const (
	Name = "EmptyStoriesCleaner"
)

var _ syncers.Runner = &Runner{}

type Runner struct {
	daysWithNoContent int
	store             service.StoryStorage
	api               service.CloudStorageAPI
	bucket            string
}

func New(daysWithNoContent int, store service.StoryStorage, api service.CloudStorageAPI, bucket string) *Runner {
	return &Runner{
		daysWithNoContent: daysWithNoContent,
		store:             store,
		api:               api,
		bucket:            bucket,
	}
}

func (r *Runner) Name() string {
	return Name
}

func (r *Runner) RunOnce(ctx context.Context, log zerolog.Logger) error {
	stories, err := r.store.GetStories(ctx)
	if err != nil {
		return fmt.Errorf("getting stories: %w", err)
	}

	var removed int

	for _, story := range stories {
		n, err := r.api.GetNumberOfObjectsWithPrefix(ctx, r.bucket, story.ID.String())
		if err != nil {
			return fmt.Errorf("getting number of objects for story '%s': %w", story.ID, err)
		}

		if n == 0 && story.Created.Before(time.Now().AddDate(0, 0, -r.daysWithNoContent)) {
			log.Info().Fields(map[string]interface{}{
				"story_id": story.ID.String(),
				"created":  story.Created.String(),
				"name":     story.Name,
			}).Msg("removing story with no content")

			err = r.store.DeleteStory(ctx, story.ID)
			if err != nil {
				return fmt.Errorf("deleting story '%s': %w", story.ID, err)
			}

			ed := r.api.DeleteObjectsWithPrefix(ctx, r.bucket, story.ID.String())
			if ed != nil {
				return fmt.Errorf("deleting story folder: %w", err)
			}

			removed++
		}
	}

	log.Info().Int("removed_count", removed).Msg("done cleaning up empty stories")

	return nil
}
