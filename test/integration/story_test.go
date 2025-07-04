package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/cloudstorage"
	"github.com/navikt/nada-backend/pkg/cloudstorage/emulator"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core"
	"github.com/navikt/nada-backend/pkg/service/core/api/gcp"
	httpapi "github.com/navikt/nada-backend/pkg/service/core/api/http"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/routes"
	"github.com/navikt/nada-backend/pkg/service/core/storage/postgres"
	"github.com/navikt/nada-backend/pkg/tk"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const (
	defaultHtml = "<html><h1>Story</h1></html>"
)

// nolint: tparallel,maintidx,goconst
func TestStory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	log := zerolog.New(os.Stdout)

	c := NewContainers(t, log)
	defer c.Cleanup()

	pgCfg := c.RunPostgres(NewPostgresConfig())

	repo, err := database.New(
		pgCfg.ConnectionURL(),
		10,
		10,
	)
	assert.NoError(t, err)

	pa1 := uuid.MustParse("00000000-1111-0000-0000-000000000000")
	pa2 := uuid.MustParse("00000000-2222-0000-0000-000000000000")
	team1 := uuid.MustParse("00000000-0000-1111-0000-000000000000")
	team2 := uuid.MustParse("00000000-0000-2222-0000-000000000000")
	team3 := uuid.MustParse("00000000-0000-3333-0000-000000000000")
	nada := uuid.MustParse("00000000-0000-4444-0000-000000000000")

	pas := []*tk.ProductArea{
		{
			ID:   pa1,
			Name: "Product area 1",
		},
		{
			ID:   pa2,
			Name: "Product area 2",
		},
	}

	teams := []*tk.Team{
		{
			ID:            team1,
			Name:          "Team1",
			Description:   "This is the first team",
			ProductAreaID: pa1,
		},
		{
			ID:            team2,
			Name:          "Team 2",
			Description:   "This is the second team",
			ProductAreaID: pa2,
		},
		{
			ID:            team3,
			Name:          "Team 3",
			Description:   "This is the third team",
			NaisTeams:     []string{"team3"},
			ProductAreaID: pa2,
		},
		{
			ID:            nada,
			Name:          "nada",
			Description:   "Nada team",
			NaisTeams:     []string{"nada"},
			ProductAreaID: pa1,
		},
	}

	staticFetcher := tk.NewStatic("http://example.com", pas, teams)

	router := TestRouter(log)

	e := emulator.New(t, nil)
	defer e.Cleanup()

	e.CreateBucket("nada-backend-stories")

	user := &service.User{
		Name:        "Bob the Builder",
		Email:       "bob.the.builder@nav.no",
		AzureGroups: nil,
		GoogleGroups: []service.GoogleGroup{
			{
				Email: GroupEmailNada,
				Name:  NaisTeamNada,
			},
		},
		AllGoogleGroups: nil,
	}

	naisConsoleStorage := postgres.NewNaisConsoleStorage(repo)
	err = naisConsoleStorage.UpdateAllTeamProjects(context.Background(), []*service.NaisTeamMapping{
		{
			Slug:       NaisTeamNada,
			GroupEmail: GroupEmailNada,
			ProjectID:  "gcp-project-team1",
		},
	})
	assert.NoError(t, err)

	tokenStorage := postgres.NewTokenStorage(repo)

	{
		teamKatalogenAPI := httpapi.NewTeamKatalogenAPI(staticFetcher, log)
		cs := cloudstorage.NewFromClient(e.Client())
		cloudStorageAPI := gcp.NewCloudStorageAPI(cs, log)
		tokenService := core.NewTokenService(tokenStorage)
		storyService := core.NewStoryService(postgres.NewStoryStorage(repo), teamKatalogenAPI, cloudStorageAPI, false, "nada-backend-stories", log)
		h := handlers.NewStoryHandler("@nav.no", storyService, tokenService, log)
		e := routes.NewStoryEndpoints(log, h)
		f := routes.NewStoryRoutes(e, InjectUser(user), h.NadaTokenMiddleware)
		f(router)
	}

	server := httptest.NewServer(router)

	story := &service.Story{}

	t.Run("Create story with oauth", func(t *testing.T) {
		newStory := &service.NewStory{
			Name:          "My new story",
			Description:   strToStrPtr("This is my story, and it is pretty bad"),
			Keywords:      []string{"story", "bad"},
			ProductAreaID: &pa1,
			TeamID:        &nada,
			Group:         GroupEmailNada,
		}

		expect := &service.Story{
			Name:             "My new story",
			Creator:          "bob.the.builder@nav.no",
			Description:      "This is my story, and it is pretty bad",
			Keywords:         []string{"story", "bad"},
			TeamkatalogenURL: nil,
			TeamID:           &nada,
			Group:            GroupEmailNada,
			// FIXME: can't set these from CreateStory, should they be?
			// TeamName:         strToStrPtr("Team1"),
			// ProductAreaName:  "Product area 1",
		}

		files := map[string]string{
			"index.html": defaultHtml,
		}
		objects := map[string]string{
			"nada-backend-new-story": string(Marshal(t, newStory)),
		}

		req := CreateMultipartFormRequest(ctx, t, http.MethodPost, server.URL+"/api/stories/new", files, objects, nil)

		NewTester(t, server).Send(req).
			HasStatusCode(http.StatusOK).
			Expect(expect, story, cmpopts.IgnoreFields(service.Story{}, "ID", "Created", "LastModified"))
	})

	t.Run("Update story with oauth", func(t *testing.T) {
		update := &service.UpdateStoryDto{
			Name:             story.Name,
			Description:      "This is a better description",
			Keywords:         story.Keywords,
			TeamkatalogenURL: story.TeamkatalogenURL,
			ProductAreaID:    &pa1,
			TeamID:           &nada,
			Group:            story.Group,
		}

		story.Description = update.Description

		got := &service.Story{}

		NewTester(t, server).
			Put(ctx, update, "/api/stories/"+story.ID.String()).
			HasStatusCode(http.StatusOK).
			Expect(story, got, cmpopts.IgnoreFields(service.Story{}, "LastModified"))

		story = got
	})

	t.Run("Get story with oauth", func(t *testing.T) {
		got := &service.Story{}

		NewTester(t, server).
			Get(ctx, "/api/stories/"+story.ID.String()).
			HasStatusCode(http.StatusOK).
			Expect(story, got)
	})

	t.Run("Get index", func(t *testing.T) {
		data := NewTester(t, server).
			Get(ctx, "/quarto/"+story.ID.String()).
			HasStatusCode(http.StatusOK).
			Body()

		assert.Equal(t, defaultHtml, data)
	})

	t.Run("Create story with multiple index.html files", func(t *testing.T) {
		newStory := &service.NewStory{
			Name:          "My story with multiple index.html files",
			Description:   strToStrPtr("This is my story, and it is pretty bad, and it has several index.html files"),
			Keywords:      []string{"story", "bad"},
			ProductAreaID: &pa1,
			TeamID:        &nada,
			Group:         "nada@nav.no",
		}

		files := map[string]string{
			"index.html":             defaultHtml,
			"a-subfolder/index.html": "<html><h1>Subfolder</h1></html>",
		}
		objects := map[string]string{
			"nada-backend-new-story": string(Marshal(t, newStory)),
		}

		req := CreateMultipartFormRequest(ctx, t, http.MethodPost, server.URL+"/api/stories/new", files, objects, nil)

		expect := &service.Story{
			Name:             "My story with multiple index.html files",
			Creator:          "bob.the.builder@nav.no",
			Description:      "This is my story, and it is pretty bad, and it has several index.html files",
			Keywords:         []string{"story", "bad"},
			TeamkatalogenURL: nil,
			TeamID:           &nada,
			Group:            "nada@nav.no",
			// FIXME: can't set these from CreateStory, should they be?
			// TeamName:         strToStrPtr("Team1"),
			// ProductAreaName:  "Product area 1",
		}

		NewTester(t, server).Send(req).
			HasStatusCode(http.StatusOK).
			Expect(expect, story, cmpopts.IgnoreFields(service.Story{}, "ID", "Created", "LastModified"))
	})

	t.Run("Get index - ensure correct top level index html is returned for story", func(t *testing.T) {
		data := NewTester(t, server).
			Get(ctx, "/quarto/"+story.ID.String()).
			HasStatusCode(http.StatusOK).
			Body()

		assert.Equal(t, defaultHtml, data)
	})

	t.Run("Delete story with oauth", func(t *testing.T) {
		NewTester(t, server).
			Delete(ctx, "/api/stories/"+story.ID.String()).
			HasStatusCode(http.StatusOK)
	})

	t.Run("Get story with oauth after delete", func(t *testing.T) {
		NewTester(t, server).
			Get(ctx, "/api/stories/"+story.ID.String()).
			HasStatusCode(http.StatusNotFound)
	})

	token, err := tokenStorage.GetNadaTokenFromGroupEmail(context.Background(), GroupEmailNada)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Create story without token", func(t *testing.T) {
		NewTester(t, server).
			Post(ctx, &service.NewStory{}, "/quarto/create").
			HasStatusCode(http.StatusUnauthorized)
	})

	t.Run("Create story for team with token", func(t *testing.T) {
		newStory := &service.NewStory{
			Name:          "My new story",
			Description:   strToStrPtr("This is my story, and it is pretty bad"),
			Keywords:      []string{"story", "bad"},
			ProductAreaID: &pa1,
			TeamID:        &nada,
			Group:         GroupEmailNada,
		}

		expect := &service.Story{
			Name:             "My new story",
			Creator:          GroupEmailNada,
			Description:      "This is my story, and it is pretty bad",
			Keywords:         []string{"story", "bad"},
			TeamkatalogenURL: strToStrPtr("http://example.com/team/00000000-0000-4444-0000-000000000000"),
			TeamID:           &nada,
			Group:            GroupEmailNada,
			// FIXME: can't set these from CreateStory, should they be?
			// TeamName:         strToStrPtr("Team1"),
			// ProductAreaName:  "Product area 1",
		}

		NewTester(t, server).
			Headers(map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)}).
			Post(ctx, newStory, "/quarto/create").
			HasStatusCode(http.StatusOK).
			Expect(expect, story, cmpopts.IgnoreFields(service.Story{}, "ID", "Created", "LastModified"))
	})

	var lastModified time.Time
	t.Run("Get story metadata after creating with team token", func(t *testing.T) {
		got := &service.Story{}

		NewTester(t, server).
			Get(ctx, "/api/stories/"+story.ID.String()).
			HasStatusCode(http.StatusOK).
			Value(got)

		expect := &service.Story{
			Name:             "My new story",
			Creator:          GroupEmailNada,
			Description:      "This is my story, and it is pretty bad",
			Keywords:         []string{"story", "bad"},
			TeamkatalogenURL: strToStrPtr("http://example.com/team/00000000-0000-4444-0000-000000000000"),
			TeamID:           &nada,
			Group:            GroupEmailNada,
		}

		diff := cmp.Diff(expect, got, cmpopts.IgnoreFields(service.Story{}, "ID", "Created", "LastModified"))
		assert.Empty(t, diff)

		// Update last modified for next test
		lastModified = *got.LastModified
	})

	t.Run("Recreate story files with token", func(t *testing.T) {
		files := map[string]string{
			"index.html":                   defaultHtml,
			"asubpage/index.html":          "<html><h1>Subpage</h1></html>",
			"subsubsubpage/something.html": "<html><h1>Subsubsubpage</h1></html>",
		}

		req := CreateMultipartFormRequest(
			ctx,
			t,
			http.MethodPut,
			server.URL+"/quarto/update/"+story.ID.String(),
			files,
			nil,
			map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", token),
			},
		)

		NewTester(t, server).
			Send(req).
			HasStatusCode(http.StatusNoContent)

		for path, content := range files {
			got := NewTester(t, server).
				Get(ctx, "/quarto/"+story.ID.String()+"/"+path).
				HasStatusCode(http.StatusOK).
				Body()

			assert.Equal(t, content, got)
		}

		got := &service.Story{}
		NewTester(t, server).
			Get(ctx, "/api/stories/"+story.ID.String()).
			HasStatusCode(http.StatusOK).
			Value(got)

		expect := &service.Story{
			Name:             "My new story",
			Creator:          GroupEmailNada,
			Description:      "This is my story, and it is pretty bad",
			Keywords:         []string{"story", "bad"},
			TeamkatalogenURL: strToStrPtr("http://example.com/team/00000000-0000-4444-0000-000000000000"),
			TeamID:           &nada,
			Group:            GroupEmailNada,
		}

		diff := cmp.Diff(expect, got, cmpopts.IgnoreFields(service.Story{}, "ID", "Created", "LastModified"))
		assert.Empty(t, diff)

		// Ensure last modified is updated when recreating story files
		assert.True(t, got.LastModified.After(lastModified))
		// Update last modified for next test
		lastModified = *got.LastModified
	})

	t.Run("Append story files with token", func(t *testing.T) {
		files := map[string]string{
			"newpage/test.html": "<html><h1>New page</h1></html>",
		}

		req := CreateMultipartFormRequest(
			ctx,
			t,
			http.MethodPatch,
			server.URL+"/quarto/update/"+story.ID.String(),
			files,
			nil,
			map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", token),
			},
		)

		NewTester(t, server).
			Send(req).
			HasStatusCode(http.StatusNoContent)

		got := NewTester(t, server).
			Get(ctx, "/quarto/"+story.ID.String()+"/newpage/test.html").
			HasStatusCode(http.StatusOK).
			Body()

		assert.Equal(t, "<html><h1>New page</h1></html>", got)

		gotStory := &service.Story{}
		NewTester(t, server).
			Get(ctx, "/api/stories/"+story.ID.String()).
			HasStatusCode(http.StatusOK).
			Value(gotStory)

		expect := &service.Story{
			Name:             "My new story",
			Creator:          GroupEmailNada,
			Description:      "This is my story, and it is pretty bad",
			Keywords:         []string{"story", "bad"},
			TeamkatalogenURL: strToStrPtr("http://example.com/team/00000000-0000-4444-0000-000000000000"),
			TeamID:           &nada,
			Group:            GroupEmailNada,
		}

		diff := cmp.Diff(expect, gotStory, cmpopts.IgnoreFields(service.Story{}, "ID", "Created", "LastModified"))
		assert.Empty(t, diff)

		// Ensure last modified is updated when appending story files
		assert.True(t, gotStory.LastModified.After(lastModified))
	})
}
