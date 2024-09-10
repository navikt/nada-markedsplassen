package empty_stories

import (
	"context"
	"fmt"
	"github.com/navikt/nada-backend/pkg/leaderelection"
	"time"

	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
)

type Syncer struct {
	daysWithNoContent int

	store service.StoryStorage
	api   service.StoryAPI

	log zerolog.Logger
}

func New(daysWithNoContent int, store service.StoryStorage, api service.StoryAPI, log zerolog.Logger) *Syncer {
	return &Syncer{
		daysWithNoContent: daysWithNoContent,
		store:             store,
		api:               api,
		log:               log,
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
		s.log.Info().Msg("Starting new stories cleaner")

		err := s.RunOnce(ctx)
		if err != nil {
			s.log.Error().Err(err).Fields(map[string]interface{}{
				"stack": errs.OpStack(err),
			}).Msg("running story cleaner")
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
					"stack": errs.OpStack(err),
				}).Msg("running story cleaner")
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

	s.log.Info().Msg("Removing stories with no content...")

	stories, err := s.store.GetStories(ctx)
	if err != nil {
		return fmt.Errorf("getting stories: %w", err)
	}

	s.log.Info().Int("empty_stories", len(stories)).Msg("Found empty stories")

	for _, story := range stories {
		n, err := s.api.GetNumberOfObjectsWithPrefix(ctx, story.ID.String())
		if err != nil {
			return fmt.Errorf("getting number of objects for story '%s': %w", story.ID, err)
		}

		if n == 0 && story.Created.Before(time.Now().AddDate(0, 0, -s.daysWithNoContent)) {
			s.log.Info().Fields(map[string]interface{}{
				"story_id": story.ID.String(),
				"created":  story.Created.String(),
				"name":     story.Name,
			}).Msg("Removing story with no content")

			err = s.store.DeleteStory(ctx, story.ID)
			if err != nil {
				return fmt.Errorf("deleting story '%s': %w", story.ID, err)
			}

			if err := s.api.DeleteStoryFolder(ctx, story.ID.String()); err != nil {
				return fmt.Errorf("deleting story folder: %w", err)
			}
		}
	}

	s.log.Info().Msg("Done removing stories with no content.")

	return nil
}
