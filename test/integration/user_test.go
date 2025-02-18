package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/config/v2"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/routes"
	"github.com/navikt/nada-backend/pkg/service/core/storage"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nolint: tparallel,gocognit,gocyclo
func TestUserDataService(t *testing.T) {
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

	r := TestRouter(log)

	stores := storage.NewStores(nil, repo, config.Config{}, log)

	StorageCreateProductAreasAndTeams(t, stores.ProductAreaStorage)
	StorageCreateNaisConsoleTeamsAndProjects(t, stores.NaisConsoleStorage, []*service.NaisTeamMapping{
		{
			Slug:       GroupNameReef,
			GroupEmail: GroupEmailReef,
			ProjectID:  "gcp-project-team1",
		},
	})
	fuel := StorageCreateDataproduct(t, stores.DataProductsStorage, NewDataProductBiofuelProduction(GroupEmailNada, TeamSeagrassID))
	barriers := StorageCreateDataproduct(t, stores.DataProductsStorage, NewDataProductProtectiveBarriers(GroupEmailReef, TeamReefID))
	feed := StorageCreateDataproduct(t, stores.DataProductsStorage, NewDataProductAquacultureFeed(GroupEmailNada, TeamSeagrassID))
	reef := StorageCreateDataproduct(t, stores.DataProductsStorage, NewDataProductReefMonitoring(GroupEmailReef, TeamReefID))

	knastADGroup := "550e8400-e29b-41d4-a716-446655440000"
	user := &service.User{
		Email: "bob.the.builder@example.com",
		GoogleGroups: []service.GoogleGroup{
			{
				Name:  GroupNameNada,
				Email: GroupEmailNada,
			},
			{
				Name:  GroupNameReef,
				Email: GroupEmailReef,
			},
		},
		IsKnastUser: true,
		AzureGroups: service.AzureGroups{
			{
				Name:     GroupNameNada,
				Email:    GroupEmailNada,
				ObjectID: knastADGroup,
			},
		},
	}
	fuelInsights := StorageCreateInsightProduct(t, user.Email, stores.InsightProductStorage, NewInsightProductBiofuelProduction(GroupEmailNada, TeamSeagrassID))
	barriersInsights := StorageCreateInsightProduct(t, user.Email, stores.InsightProductStorage, NewInsightProductProtectiveBarriers(GroupEmailReef, TeamReefID))
	feedInsights := StorageCreateInsightProduct(t, user.Email, stores.InsightProductStorage, NewInsightProductAquacultureFeed(GroupEmailNada, TeamSeagrassID))
	reefInsights := StorageCreateInsightProduct(t, user.Email, stores.InsightProductStorage, NewInsightProductReefMonitoring(GroupEmailReef, TeamReefID))

	fuelStory := StorageCreateStory(t, stores.StoryStorage, user.Email, NewStoryBiofuelProduction(GroupEmailNada, &TeamSeagrassID))
	barriersStory := StorageCreateStory(t, stores.StoryStorage, user.Email, NewStoryProtectiveBarriers(GroupEmailReef, &TeamReefID))
	feedStory := StorageCreateStory(t, stores.StoryStorage, user.Email, NewStoryAquacultureFeed(GroupEmailNada, &TeamSeagrassID))
	reefStory := StorageCreateStory(t, stores.StoryStorage, user.Email, NewStoryReefMonitoring(GroupEmailReef, &TeamReefID))

	{
		s := core.NewUserService(stores.AccessStorage, stores.PollyStorage, stores.TokenStorage, stores.StoryStorage, stores.DataProductsStorage,
			stores.InsightProductStorage, stores.NaisConsoleStorage, log)
		h := handlers.NewUserHandler(s)
		e := routes.NewUserEndpoints(log, h)
		f := routes.NewUserRoutes(e, injectUser(user))
		f(r)
	}

	server := httptest.NewServer(r)
	defer server.Close()
	t.Run("User data products are sorted alphabetically by team_name and dp_name", func(t *testing.T) {
		got := &service.UserInfo{}
		expect := []service.Dataproduct{
			{ID: barriers.ID, Name: barriers.Name, TeamName: &TeamReefName, Owner: &service.DataproductOwner{Group: GroupEmailReef}},
			{ID: reef.ID, Name: reef.Name, TeamName: &TeamReefName, Owner: &service.DataproductOwner{Group: GroupEmailReef}},
			{ID: feed.ID, Name: feed.Name, TeamName: &TeamSeagrassName, Owner: &service.DataproductOwner{Group: GroupEmailNada}},
			{ID: fuel.ID, Name: fuel.Name, TeamName: &TeamSeagrassName, Owner: &service.DataproductOwner{Group: GroupEmailNada}},
		}

		NewTester(t, server).Get(ctx, "/api/userData").
			HasStatusCode(http.StatusOK).Value(got)

		if len(got.Dataproducts) != len(expect) {
			t.Fatalf("got %d, expected %d", len(got.Dataproducts), len(expect))
		}
		for i := 0; i < len(got.Dataproducts); i++ {
			if got.Dataproducts[i].ID != expect[i].ID {
				t.Errorf("got %s, expected %s", got.Dataproducts[i].ID, expect[i].ID)
			}
			if got.Dataproducts[i].Name != expect[i].Name {
				t.Errorf("got %s, expected %s", got.Dataproducts[i].Name, expect[i].Name)
			}
			if *got.Dataproducts[i].TeamName != *expect[i].TeamName {
				t.Errorf("got %s, expected %s", *got.Dataproducts[i].TeamName, *expect[i].TeamName)
			}
			if got.Dataproducts[i].Owner.Group != expect[i].Owner.Group {
				t.Errorf("got %s, expected %s", got.Dataproducts[i].Owner.Group, expect[i].Owner.Group)
			}
		}
	})
	t.Run("User insight products are sorted alphabetically by team_name and name", func(t *testing.T) {
		got := &service.UserInfo{}
		expect := []service.InsightProduct{
			{ID: barriersInsights.ID, Name: barriersInsights.Name, TeamName: &TeamReefName, Group: GroupEmailReef},
			{ID: reefInsights.ID, Name: reefInsights.Name, TeamName: &TeamReefName, Group: GroupEmailReef},
			{ID: feedInsights.ID, Name: feedInsights.Name, TeamName: &TeamSeagrassName, Group: GroupEmailNada},
			{ID: fuelInsights.ID, Name: fuelInsights.Name, TeamName: &TeamSeagrassName, Group: GroupEmailNada},
		}

		NewTester(t, server).Get(ctx, "/api/userData").
			HasStatusCode(http.StatusOK).Value(got)

		if len(got.InsightProducts) != len(expect) {
			t.Fatalf("got %d, expected %d", len(got.InsightProducts), len(expect))
		}
		for i := 0; i < len(got.InsightProducts); i++ {
			if got.InsightProducts[i].ID != expect[i].ID {
				t.Errorf("got %s, expected %s", got.InsightProducts[i].ID, expect[i].ID)
			}
			if got.InsightProducts[i].Name != expect[i].Name {
				t.Errorf("got %s, expected %s", got.InsightProducts[i].Name, expect[i].Name)
			}
			if *got.InsightProducts[i].TeamName != *expect[i].TeamName {
				t.Errorf("got %s, expected %s", *got.InsightProducts[i].TeamName, *expect[i].TeamName)
			}
			if got.InsightProducts[i].Group != expect[i].Group {
				t.Errorf("got %s, expected %s", got.InsightProducts[i].Group, expect[i].Group)
			}
		}
	})

	t.Run("User stories are sorted alphabetically by team_name and name", func(t *testing.T) {
		got := &service.UserInfo{}
		expect := []service.Story{
			{ID: barriersStory.ID, Name: barriersStory.Name, TeamName: &TeamReefName, Group: GroupEmailReef},
			{ID: reefStory.ID, Name: reefStory.Name, TeamName: &TeamReefName, Group: GroupEmailReef},
			{ID: feedStory.ID, Name: feedStory.Name, TeamName: &TeamSeagrassName, Group: GroupEmailNada},
			{ID: fuelStory.ID, Name: fuelStory.Name, TeamName: &TeamSeagrassName, Group: GroupEmailNada},
		}

		NewTester(t, server).Get(ctx, "/api/userData").
			HasStatusCode(http.StatusOK).Value(got)

		if len(got.Stories) != len(expect) {
			t.Fatalf("got %d, expected %d", len(got.Stories), len(expect))
		}
		for i := 0; i < len(got.Stories); i++ {
			if got.Stories[i].ID != expect[i].ID {
				t.Errorf("got %s, expected %s", got.Stories[i].ID, expect[i].ID)
			}
			if got.Stories[i].Name != expect[i].Name {
				t.Errorf("got %s, expected %s", got.Stories[i].Name, expect[i].Name)
			}
			if *got.Stories[i].TeamName != *expect[i].TeamName {
				t.Errorf("got %s, expected %s", *got.Stories[i].TeamName, *expect[i].TeamName)
			}
			if got.Stories[i].Group != expect[i].Group {
				t.Errorf("got %s, expected %s", got.Stories[i].Group, expect[i].Group)
			}
		}
	})

	t.Run("Access requests in userData only contains polly object when user has specified a polly reference", func(t *testing.T) {
		pollyRef := service.QueryPolly{
			ExternalID: "6f865942-e3a6-4c23-948f-6f372ee67037",
			Name:       "behandling",
			URL:        "https://polly-catalog.com/process/purpose/behandling/6f865942-e3a6-4c23-948f-6f372ee67037",
		}

		polly, err := stores.PollyStorage.CreatePollyDocumentation(ctx, service.PollyInput{
			QueryPolly: pollyRef,
		})
		assert.NoError(t, err)

		dp := StorageCreateDataproduct(t, stores.DataProductsStorage, service.NewDataproduct{
			Name:  "Dataprodukt",
			Group: GroupEmailNada,
		})

		ds, err := stores.DataProductsStorage.CreateDataset(ctx, service.NewDataset{
			DataproductID: dp.ID,
			Name:          "Datasett",
			Pii:           service.PiiLevelNone,
		}, nil, user)
		assert.NoError(t, err)

		arWithPollyRef, err := stores.AccessStorage.CreateAccessRequestForDataset(ctx, ds.ID, uuid.NullUUID{UUID: polly.ID, Valid: true}, "user:"+user.Email, "user:"+user.Email, nil)
		assert.NoError(t, err)

		arWithoutPollyRef, err := stores.AccessStorage.CreateAccessRequestForDataset(ctx, ds.ID, uuid.NullUUID{}, "user:"+user.Email, "user:"+user.Email, nil)
		assert.NoError(t, err)

		got := &service.UserInfo{}
		expected := []service.AccessRequest{
			{
				ID:          arWithoutPollyRef.ID,
				DatasetID:   ds.ID,
				Polly:       nil,
				Subject:     user.Email,
				SubjectType: string(service.SubjectTypeUser),
				Owner:       user.Email,
				Status:      service.AccessRequestStatusPending,
				Granter:     nil,
				Reason:      nil,
			},
			{
				ID:        arWithPollyRef.ID,
				DatasetID: ds.ID,
				Polly: &service.Polly{
					ID:         polly.ID,
					QueryPolly: pollyRef,
				},
				Subject:     user.Email,
				SubjectType: string(service.SubjectTypeUser),
				Owner:       user.Email,
				Status:      service.AccessRequestStatusPending,
				Granter:     nil,
				Reason:      nil,
			},
		}

		NewTester(t, server).Get(ctx, "/api/userData").
			HasStatusCode(http.StatusOK).Value(got)

		require.Len(t, got.AccessRequests, 2)
		diff := cmp.Diff(expected, got.AccessRequests, cmpopts.IgnoreFields(service.AccessRequest{}, "Created", "Closed", "Expires"))
		assert.Empty(t, diff)

		require.Len(t, got.AccessRequestsAsGranter, 2)
		for _, arg := range got.AccessRequestsAsGranter {
			err := ContainsAccessRequest(t, expected, arg.AccessRequest)
			assert.NoError(t, err)
		}
	})

	t.Run("User is member of an AD group allowed to create Knast", func(t *testing.T) {
		got := &service.UserInfo{}
		NewTester(t, server).Get(ctx, "/api/userData").
			HasStatusCode(http.StatusOK).Value(got)

		assert.True(t, got.IsKnastUser)
	})

	noKnastUser := &service.User{
		Name:  "No Knast",
		Email: "no-knast@nav.no",
	}

	noKnastRouter := TestRouter(log)

	{
		s := core.NewUserService(stores.AccessStorage, stores.PollyStorage, stores.TokenStorage, stores.StoryStorage, stores.DataProductsStorage,
			stores.InsightProductStorage, stores.NaisConsoleStorage, log)
		h := handlers.NewUserHandler(s)
		e := routes.NewUserEndpoints(log, h)
		noKnastRoutes := routes.NewUserRoutes(e, injectUser(noKnastUser))
		noKnastRoutes(noKnastRouter)
	}

	noKnastServer := httptest.NewServer(noKnastRouter)

	t.Run("User is not allowed to create a Knast", func(t *testing.T) {
		got := &service.UserInfo{}
		NewTester(t, noKnastServer).Get(ctx, "/api/userData").
			HasStatusCode(http.StatusOK).Value(got)

		assert.False(t, got.IsKnastUser)
	})
}
