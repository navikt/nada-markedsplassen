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

var _ syncers.Runner = &Syncer{}

type Syncer struct {
	daysWithNoContent int
	store             service.StoryStorage
	api               service.StoryAPI
}

func New(daysWithNoContent int, store service.StoryStorage, api service.StoryAPI) *Syncer {
	return &Syncer{
		daysWithNoContent: daysWithNoContent,
		store:             store,
		api:               api,
	}
}

func (s *Syncer) Name() string {
	return Name
}

func (s *Syncer) RunOnce(ctx context.Context, log zerolog.Logger) error {
	stories, err := s.store.GetStories(ctx)
	if err != nil {
		return fmt.Errorf("getting stories: %w", err)
	}

	log.Info().Int("empty_stories", len(stories)).Msg("found empty stories")

	for _, story := range stories {
		n, err := s.api.GetNumberOfObjectsWithPrefix(ctx, story.ID.String())
		if err != nil {
			return fmt.Errorf("getting number of objects for story '%s': %w", story.ID, err)
		}

		if n == 0 && story.Created.Before(time.Now().AddDate(0, 0, -s.daysWithNoContent)) {
			log.Info().Fields(map[string]interface{}{
				"story_id": story.ID.String(),
				"created":  story.Created.String(),
				"name":     story.Name,
			}).Msg("removing story with no content")

			err = s.store.DeleteStory(ctx, story.ID)
			if err != nil {
				return fmt.Errorf("deleting story '%s': %w", story.ID, err)
			}

			if err := s.api.DeleteStoryFolder(ctx, story.ID.String()); err != nil {
				return fmt.Errorf("deleting story folder: %w", err)
			}
		}
	}

	return nil
}
