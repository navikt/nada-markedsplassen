package integration

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service/core"
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

// blockingCloudStorageAPI blocks the first file write until proceed is closed,
// allowing tests to run code while a story upload is in progress.
type blockingCloudStorageAPI struct {
	service.CloudStorageAPI
	writeStarted chan struct{}
	proceed      chan struct{}
	once         sync.Once
}

func newBlockingCloudStorageAPI(api service.CloudStorageAPI) *blockingCloudStorageAPI {
	return &blockingCloudStorageAPI{
		CloudStorageAPI: api,
		writeStarted:    make(chan struct{}),
		proceed:         make(chan struct{}),
	}
}

func (b *blockingCloudStorageAPI) WriteFileToBucket(ctx context.Context, bucket, pathPrefix string, file *service.UploadFile) error {
	b.once.Do(func() {
		close(b.writeStarted)
		<-b.proceed
	})

	return b.CloudStorageAPI.WriteFileToBucket(ctx, bucket, pathPrefix, file)
}

// staleStoryStorage returns a fixed list of stories, simulating a cleaner
// that read the stories before a concurrent update started.
type staleStoryStorage struct {
	service.StoryStorage
	stories []*service.Story
}

func (s *staleStoryStorage) GetStories(_ context.Context) ([]*service.Story, error) {
	return s.stories, nil
}

// nolint: tparallel,maintidx
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
	cs := cloudstorage.NewFromClient(e.Client())
	cloudStorageAPI := gcp.NewCloudStorageAPI(cs, log)

	backdateStory := func(t *testing.T, id uuid.UUID) {
		t.Helper()

		// Disable the trigger that sets last_modified on update, so we can backdate it
		_, err := repo.GetDB().Exec("ALTER TABLE stories DISABLE TRIGGER quarto_story_modified")
		require.NoError(t, err)

		_, err = repo.GetDB().Exec("UPDATE stories SET created = NOW() - INTERVAL '8 day', last_modified = NOW() - INTERVAL '8 day' WHERE id = $1", id)
		require.NoError(t, err)

		_, err = repo.GetDB().Exec("ALTER TABLE stories ENABLE TRIGGER quarto_story_modified")
		require.NoError(t, err)
	}

	uploadFiles := func() []*service.UploadFile {
		return []*service.UploadFile{
			{
				Path:       "index.html",
				ReadCloser: io.NopCloser(strings.NewReader("<html><h1>Story</h1></html>")),
			},
		}
	}

	t.Run("Deleting stories with no content after deadline", func(t *testing.T) {
		storyToBeDeleted, err := storyStorage.CreateStory(context.Background(), "nada@nav.no", &service.NewStory{
			Name:        "This empty story will be deleted",
			Description: strToStrPtr("This is my story"),
			Keywords:    []string{},
			Group:       "nada@nav.no",
		})
		require.NoError(t, err)

		backdateStory(t, storyToBeDeleted.ID)

		story, err := storyStorage.CreateStory(context.Background(), "nada@nav.no", &service.NewStory{
			Name:        "This empty story will not be deleted",
			Description: strToStrPtr("This is my story"),
			Keywords:    []string{},
			Group:       "nada@nav.no",
		})
		require.NoError(t, err)

		cleaner := empty_stories.New(7, storyStorage, cloudStorageAPI, "nada-backend-stories")
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

		backdateStory(t, storyToBeDeleted.ID)

		e.CreateObject("nada-backend-stories", storyToBeDeleted.ID.String()+"/file.txt", "content")

		cleaner := empty_stories.New(7, storyStorage, cloudStorageAPI, "nada-backend-stories")
		err = cleaner.RunOnce(context.Background(), log)
		require.NoError(t, err)

		story, err := storyStorage.GetStory(context.Background(), storyToBeDeleted.ID)
		require.NoError(t, err)
		assert.Equal(t, storyToBeDeleted.ID, story.ID)
	})

	t.Run("Does not delete empty story with recent modifications", func(t *testing.T) {
		story, err := storyStorage.CreateStory(context.Background(), "nada@nav.no", &service.NewStory{
			Name:        "This empty story was recently modified and should not be deleted",
			Description: strToStrPtr("This is my story"),
			Keywords:    []string{},
			Group:       "nada@nav.no",
		})
		require.NoError(t, err)

		backdateStory(t, story.ID)

		err = storyStorage.UpdateStoryLastModified(context.Background(), story.ID)
		require.NoError(t, err)

		cleaner := empty_stories.New(7, storyStorage, cloudStorageAPI, "nada-backend-stories")
		err = cleaner.RunOnce(context.Background(), log)
		require.NoError(t, err)

		_, err = storyStorage.GetStory(context.Background(), story.ID)
		require.NoError(t, err)
	})

	t.Run("Does not delete story while its files are being recreated", func(t *testing.T) {
		story, err := storyStorage.CreateStory(context.Background(), "nada@nav.no", &service.NewStory{
			Name:        "This story is being updated and should not be deleted",
			Description: strToStrPtr("This is my story"),
			Keywords:    []string{},
			Group:       "nada@nav.no",
		})
		require.NoError(t, err)

		e.CreateObject("nada-backend-stories", story.ID.String()+"/index.html", "old content")
		backdateStory(t, story.ID)

		staleStory, err := storyStorage.GetStory(context.Background(), story.ID)
		require.NoError(t, err)

		blockingAPI := newBlockingCloudStorageAPI(cloudStorageAPI)
		storyService := core.NewStoryService(storyStorage, nil, blockingAPI, false, "nada-backend-stories", log)

		errCh := make(chan error, 1)
		go func() {
			errCh <- storyService.RecreateStoryFiles(context.Background(), story.ID, "nada@nav.no", uploadFiles())
		}()

		// The existing files have been removed, and the new files are being uploaded
		<-blockingAPI.writeStarted

		n, err := cloudStorageAPI.GetNumberOfObjectsWithPrefix(context.Background(), "nada-backend-stories", story.ID.String())
		require.NoError(t, err)
		require.Equal(t, 0, n)

		cleaner := empty_stories.New(7, storyStorage, cloudStorageAPI, "nada-backend-stories")
		err = cleaner.RunOnce(context.Background(), log)
		require.NoError(t, err)

		// Cleaner that read the story before the update started
		staleCleaner := empty_stories.New(7, &staleStoryStorage{
			StoryStorage: storyStorage,
			stories:      []*service.Story{staleStory},
		}, cloudStorageAPI, "nada-backend-stories")
		err = staleCleaner.RunOnce(context.Background(), log)
		require.NoError(t, err)

		close(blockingAPI.proceed)
		require.NoError(t, <-errCh)

		got, err := storyStorage.GetStory(context.Background(), story.ID)
		require.NoError(t, err)
		assert.Equal(t, story.ID, got.ID)

		n, err = cloudStorageAPI.GetNumberOfObjectsWithPrefix(context.Background(), "nada-backend-stories", story.ID.String())
		require.NoError(t, err)
		assert.Equal(t, 1, n)
	})

	t.Run("Recreating story files fails if the story is removed during upload", func(t *testing.T) {
		story, err := storyStorage.CreateStory(context.Background(), "nada@nav.no", &service.NewStory{
			Name:        "This story is removed while being updated",
			Description: strToStrPtr("This is my story"),
			Keywords:    []string{},
			Group:       "nada@nav.no",
		})
		require.NoError(t, err)

		blockingAPI := newBlockingCloudStorageAPI(cloudStorageAPI)
		storyService := core.NewStoryService(storyStorage, nil, blockingAPI, false, "nada-backend-stories", log)

		errCh := make(chan error, 1)
		go func() {
			errCh <- storyService.RecreateStoryFiles(context.Background(), story.ID, "nada@nav.no", uploadFiles())
		}()

		<-blockingAPI.writeStarted

		err = storyStorage.DeleteStory(context.Background(), story.ID)
		require.NoError(t, err)

		close(blockingAPI.proceed)

		err = <-errCh
		require.Error(t, err)
		assert.True(t, errs.KindIs(errs.NotExist, err))

		n, err := cloudStorageAPI.GetNumberOfObjectsWithPrefix(context.Background(), "nada-backend-stories", story.ID.String())
		require.NoError(t, err)
		assert.Equal(t, 0, n)
	})
}
