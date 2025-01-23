package integration

import (
	"context"
	"os"
	"testing"

	"github.com/navikt/nada-backend/pkg/syncers/empty_stories"
	"github.com/stretchr/testify/require"

	"github.com/navikt/nada-backend/pkg/cloudstorage"
	"github.com/navikt/nada-backend/pkg/cloudstorage/emulator"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core/api/gcp"
	"github.com/navikt/nada-backend/pkg/service/core/storage/postgres"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// nolint: tparallel
func TestStoryCleaner(t *testing.T) {
	t.Parallel()

	log := zerolog.New(os.Stdout)

	c := NewContainers(t, log)
	defer c.Cleanup()

	pgCfg := c.RunPostgres(NewPostgresConfig())

	repo, err := database.New(
		pgCfg.ConnectionURL(),
		10,
		10,
	)
	require.NoError(t, err)

	e := emulator.New(t, nil)
	defer e.Cleanup()

	e.CreateBucket("nada-backend-stories")

	naisConsoleStorage := postgres.NewNaisConsoleStorage(repo)
	err = naisConsoleStorage.UpdateAllTeamProjects(context.Background(), []*service.NaisTeamMapping{
		{
			Slug:       NaisTeamNada,
			GroupEmail: GroupEmailNada,
			ProjectID:  "gcp-project-team1",
		},
	})
	require.NoError(t, err)

	storyStorage := postgres.NewStoryStorage(repo)
	cs := cloudstorage.NewFromClient("nada-backend-stories", e.Client())
	storyAPI := gcp.NewStoryAPI(cs, log)

	t.Run("Deleting stories with no content after deadline", func(t *testing.T) {
		storyToBeDeleted, err := storyStorage.CreateStory(context.Background(), "nada@nav.no", &service.NewStory{
			Name:        "This empty story will be deleted",
			Description: strToStrPtr("This is my story"),
			Keywords:    []string{},
			Group:       "nada@nav.no",
		})
		require.NoError(t, err)

		_, err = repo.GetDB().Exec("UPDATE stories SET created = NOW() - INTERVAL '8 day' WHERE id = $1", storyToBeDeleted.ID)
		require.NoError(t, err)

		story, err := storyStorage.CreateStory(context.Background(), "nada@nav.no", &service.NewStory{
			Name:        "This empty story will not be deleted",
			Description: strToStrPtr("This is my story"),
			Keywords:    []string{},
			Group:       "nada@nav.no",
		})
		require.NoError(t, err)

		cleaner := empty_stories.New(7, storyStorage, storyAPI)
		err = cleaner.RunOnce(context.Background(), log)
		require.NoError(t, err)

		stories, err := storyStorage.GetStories(context.Background())
		require.NoError(t, err)

		assert.Len(t, stories, 1)
		assert.Equal(t, story.ID, stories[0].ID)
	})

	t.Run("Does not delete stories with content after deadline", func(t *testing.T) {
		storyToBeDeleted, err := storyStorage.CreateStory(context.Background(), "nada@nav.no", &service.NewStory{
			Name:        "This story with content after deadline should not be deleted",
			Description: strToStrPtr("This is my story"),
			Keywords:    []string{},
			Group:       "nada@nav.no",
		})
		require.NoError(t, err)

		_, err = repo.GetDB().Exec("UPDATE stories SET created = NOW() - INTERVAL '8 day' WHERE id = $1", storyToBeDeleted.ID)
		require.NoError(t, err)

		e.CreateObject("nada-backend-stories", storyToBeDeleted.ID.String()+"/file.txt", "content")

		cleaner := empty_stories.New(7, storyStorage, storyAPI)
		err = cleaner.RunOnce(context.Background(), log)
		require.NoError(t, err)

		story, err := storyStorage.GetStory(context.Background(), storyToBeDeleted.ID)
		require.NoError(t, err)
		assert.Equal(t, storyToBeDeleted.ID, story.ID)
	})
}
