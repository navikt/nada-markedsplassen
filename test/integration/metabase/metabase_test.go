package integration_metabase

import (
	"context"
	"fmt"
	httpapi "net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/kms"
	"github.com/navikt/nada-backend/pkg/kms/emulator"
	river2 "github.com/navikt/nada-backend/pkg/service/core/queue/river"
	"github.com/navikt/nada-backend/pkg/syncers/metabase_collections"
	"github.com/navikt/nada-backend/pkg/worker"
	"github.com/riverqueue/river"
	"golang.org/x/oauth2/google"

	"github.com/navikt/nada-backend/pkg/bq"
	crm "github.com/navikt/nada-backend/pkg/cloudresourcemanager"
	"github.com/navikt/nada-backend/pkg/config/v2"
	"github.com/navikt/nada-backend/pkg/database"
	dmpSA "github.com/navikt/nada-backend/pkg/sa"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core"
	"github.com/navikt/nada-backend/pkg/service/core/api/gcp"
	"github.com/navikt/nada-backend/pkg/service/core/api/http"
	"github.com/navikt/nada-backend/pkg/service/core/api/static"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/routes"
	"github.com/navikt/nada-backend/pkg/service/core/storage"
	"github.com/navikt/nada-backend/test/integration"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	metabaseAdminGroupID          = 2
	metabaseDashboardCollectionID = 2
)

// nolint: tparallel,maintidx
func TestMetabaseOpenDataset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(20*time.Minute))
	defer cancel()

	log := zerolog.New(zerolog.NewConsoleWriter())
	log.Level(zerolog.DebugLevel)

	gcpHelper := NewGCPHelper(t, log)
	cleanupFn := gcpHelper.Start(ctx)
	defer cleanupFn(ctx)

	c := integration.NewContainers(t, log)
	defer c.Cleanup()

	pgCfg := c.RunPostgres(integration.NewPostgresConfig())

	repo, err := database.New(
		pgCfg.ConnectionURL(),
		10,
		10,
	)
	assert.NoError(t, err)

	mbCfg := c.RunMetabase(integration.NewMetabaseConfig(), "../../../.metabase_version")

	bqClient := bq.NewClient("", true, log)
	saClient := dmpSA.NewClient("", false)
	crmClient := crm.NewClient("", false, nil)

	kmsEmulator := emulator.New(log)
	kmsEmulator.AddSymmetricKey(integration.MetabaseProject, integration.Location, integration.Keyring, integration.MetabaseKeyName, []byte("7b483b28d6e67cfd3b9b5813a286c763"))
	kmsURL := kmsEmulator.Run()

	kmsClient := kms.NewClient(kmsURL, true)

	stores := storage.NewStores(nil, repo, config.Config{}, log)

	zlog := zerolog.New(os.Stdout)
	r := integration.TestRouter(zlog)

	crmapi := gcp.NewCloudResourceManagerAPI(crmClient)
	saapi := gcp.NewServiceAccountAPI(saClient)
	bqapi := gcp.NewBigQueryAPI(integration.MetabaseProject, integration.Location, integration.PseudoDataSet, bqClient)
	kmsapi := gcp.NewKMSAPI(kmsClient)

	mbapi := http.NewMetabaseHTTP(
		mbCfg.ConnectionURL()+"/api",
		mbCfg.Email,
		mbCfg.Password,
		"",
		false,
		false,
		log,
	)

	credBytes, err := os.ReadFile("../../../tests-metabase-all-users-sa-creds.json")
	assert.NoError(t, err)

	_, err = google.CredentialsFromJSON(ctx, credBytes)
	if err != nil {
		t.Fatalf("Failed to parse Metabase service account credentials: %v", err)
	}

	workers := river.NewWorkers()
	riverConfig := worker.RiverConfig(&zlog, workers)
	riverConfig.PeriodicJobs = []*river.PeriodicJob{}
	mbqueue := river2.NewMetabaseQueue(repo, riverConfig)

	mbService := core.NewMetabaseService(
		integration.MetabaseProject,
		integration.Location,
		integration.Keyring,
		integration.MetabaseKeyName,
		string(credBytes),
		integration.MetabaseAllUsersServiceAccount,
		"group:"+integration.GroupEmailAllUsers,
		integration.GroupEmailAllUsers,
		mbqueue,
		kmsapi,
		mbapi,
		bqapi,
		saapi,
		crmapi,
		stores.OpenMetabaseStorage,
		stores.RestrictedMetaBaseStorage,
		stores.BigQueryStorage,
		stores.DataProductsStorage,
		stores.AccessStorage,
		zlog,
	)

	err = worker.MetabaseAddWorkers(riverConfig, mbService, repo)
	require.NoError(t, err)

	riverClient, err := worker.RiverClient(riverConfig, repo)
	require.NoError(t, err)

	err = riverClient.Start(ctx)
	require.NoError(t, err)

	defer riverClient.Stop(ctx)

	err = stores.NaisConsoleStorage.UpdateAllTeamProjects(ctx, []*service.NaisTeamMapping{
		{
			Slug:       integration.NaisTeamNada,
			GroupEmail: integration.GroupEmailNada,
			ProjectID:  MetabaseProject,
		},
	})
	assert.NoError(t, err)

	dataproductService := core.NewDataProductsService(
		stores.DataProductsStorage,
		stores.BigQueryStorage,
		bqapi,
		stores.NaisConsoleStorage,
		integration.GroupEmailAllUsers,
	)

	{
		h := handlers.NewMetabaseHandler(mbService)
		e := routes.NewMetabaseEndpoints(zlog, h)
		f := routes.NewMetabaseRoutes(e, integration.InjectUser(integration.UserOne))

		f(r)
	}

	slack := static.NewSlackAPI(log)
	accessService := core.NewAccessService(
		"",
		slack,
		stores.PollyStorage,
		stores.AccessStorage,
		stores.DataProductsStorage,
		stores.BigQueryStorage,
		stores.JoinableViewsStorage,
		bqapi,
	)

	{
		h := handlers.NewAccessHandler(accessService, mbService, MetabaseProject)
		e := routes.NewAccessEndpoints(zlog, h)
		f := routes.NewAccessRoutes(e, integration.InjectUser(integration.UserOne))

		f(r)
	}

	mbDashboardService := core.NewMetabaseDashboardsService(stores.MetabaseDashboardStorage, mbapi, mbCfg.PublicHost)
	{
		h := handlers.NewMetabaseDashboardsHandler(mbDashboardService)
		e := routes.NewMetabaseDashboardEndpoints(zlog, h)
		f := routes.NewMetabaseDashboardRouter(e, integration.InjectUser(integration.UserOne))
		f(r)
	}

	notAllowedRouter := integration.TestRouter(zlog)
	{
		h := handlers.NewMetabaseDashboardsHandler(mbDashboardService)
		e := routes.NewMetabaseDashboardEndpoints(zlog, h)
		f := routes.NewMetabaseDashboardRouter(e, integration.InjectUser(integration.UserTwo))
		f(notAllowedRouter)
	}

	server := httptest.NewServer(r)
	defer server.Close()

	notAllowedServer := httptest.NewServer(notAllowedRouter)
	defer notAllowedServer.Close()

	integration.StorageCreateProductAreasAndTeams(t, stores.ProductAreaStorage)
	dataproduct, err := dataproductService.CreateDataproduct(ctx, integration.UserOne, integration.NewDataProductBiofuelProduction(integration.GroupEmailNada, integration.TeamSeagrassID))
	assert.NoError(t, err)

	openDataset, err := dataproductService.CreateDataset(ctx, integration.UserOne, service.NewDataset{
		DataproductID: dataproduct.ID,
		Name:          "Open dataset",
		BigQuery:      gcpHelper.BigQueryTable,
		Pii:           service.PiiLevelNone,
	})
	assert.NoError(t, err)

	newPublicMetabaseDashboard := integration.NewPublicMetabaseDashboardInput(integration.GroupEmailNada, integration.TeamNadaID)

	var createdDashboardID uuid.UUID
	publicDashboard := service.PublicMetabaseDashboardOutput{}

	t.Run("Adding an open bigquery dataset to metabase", func(t *testing.T) {
		integration.NewTester(t, server).
			Post(ctx, service.GrantAccessData{
				DatasetID:   openDataset.ID,
				Expires:     nil,
				Subject:     strToStrPtr(integration.GroupEmailAllUsers),
				SubjectType: strToStrPtr("group"),
			}, "/api/accesses/grant").
			HasStatusCode(httpapi.StatusNoContent)

		integration.NewTester(t, server).
			Post(ctx, nil, fmt.Sprintf("/api/datasets/%s/bigquery_open", openDataset.ID)).
			HasStatusCode(httpapi.StatusOK)

		status := &service.MetabaseBigQueryDatasetStatus{}
		for {
			integration.NewTester(t, server).
				Get(ctx, fmt.Sprintf("/api/datasets/%s/bigquery_open", openDataset.ID)).
				HasStatusCode(httpapi.StatusOK).
				Value(status)

			if err := status.Error(); err != nil {
				fmt.Println("Error: ", err)
			}

			if status.HasFailed {
				t.Fatalf("Failed to add open dataset to Metabase: %s", status.Error())
			}

			if !status.IsCompleted {
				time.Sleep(5 * time.Second)
				continue
			}

			break
		}

		meta, err := stores.OpenMetabaseStorage.GetMetadata(ctx, openDataset.ID)
		require.NoError(t, err)
		require.NotNil(t, meta.DatabaseID)
		require.NotNil(t, meta.SyncCompleted)

		permissionGraphForGroup, err := mbapi.GetPermissionGraphForGroup(ctx, service.MetabaseAllUsersGroupID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, permissionGraphForGroup.Groups, strconv.Itoa(service.MetabaseAllUsersGroupID))

		// When adding an open dataset to metabase the all users group should be granted access
		// while not losing access to the default open "sample dataset" database
		assert.Equal(t, numberOfDatabasesWithAccessForPermissionGroup(permissionGraphForGroup.Groups[strconv.Itoa(service.MetabaseAllUsersGroupID)]), 2)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, openDataset.Datasource.ProjectID, openDataset.Datasource.Dataset, openDataset.Datasource.Table)
		assert.NoError(t, err)
		assert.True(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+integration.MetabaseAllUsersServiceAccount))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, openDataset.Datasource.Dataset)
		assert.NoError(t, err)
		assert.True(t, integration.ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, integration.MetabaseAllUsersServiceAccount))
	})

	t.Run("Permanent delete of open metabase database", func(t *testing.T) {
		meta, err := stores.OpenMetabaseStorage.GetMetadata(ctx, openDataset.ID)
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		integration.NewTester(t, server).
			Delete(ctx, fmt.Sprintf("/api/datasets/%s/bigquery_open", openDataset.ID)).
			HasStatusCode(httpapi.StatusNoContent)

		time.Sleep(500 * time.Millisecond)

		_, err = stores.RestrictedMetaBaseStorage.GetMetadata(ctx, openDataset.ID)
		require.Error(t, err)

		_, err = mbapi.Database(ctx, *meta.DatabaseID)
		require.Error(t, err)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, openDataset.Datasource.ProjectID, openDataset.Datasource.Dataset, openDataset.Datasource.Table)
		assert.NoError(t, err)
		assert.False(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+integration.MetabaseAllUsersServiceAccount))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, openDataset.Datasource.Dataset)
		assert.NoError(t, err)
		// Dataset Metadata Viewer is intentionally not removed when access for table is revoked so should be true
		assert.True(t, integration.ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, integration.MetabaseAllUsersServiceAccount))
	})

	t.Run("Soft delete open metabase database", func(t *testing.T) {
		datasetAccessEntries, err := stores.AccessStorage.ListActiveAccessToDataset(ctx, openDataset.ID)
		require.NoError(t, err)
		require.Len(t, datasetAccessEntries, 1)

		integration.NewTester(t, server).
			Post(ctx, nil, fmt.Sprintf("/api/accesses/revoke?accessId=%s", datasetAccessEntries[0].ID)).
			HasStatusCode(httpapi.StatusNoContent)

		datasetAccessEntries, err = stores.AccessStorage.ListActiveAccessToDataset(ctx, openDataset.ID)
		require.NoError(t, err)
		require.Len(t, datasetAccessEntries, 0)

		meta, err := stores.OpenMetabaseStorage.GetMetadata(ctx, openDataset.ID)
		require.NoError(t, err)
		require.NotNil(t, meta.DatabaseID)

		permissionGraphForGroup, err := mbapi.GetPermissionGraphForGroup(ctx, service.MetabaseAllUsersGroupID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Contains(t, permissionGraphForGroup.Groups, strconv.Itoa(service.MetabaseAllUsersGroupID))

		time.Sleep(1 * time.Second)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, openDataset.Datasource.ProjectID, openDataset.Datasource.Dataset, openDataset.Datasource.Table)
		assert.NoError(t, err)
		assert.False(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+integration.MetabaseAllUsersServiceAccount))
	})

	t.Run("Restore soft deleted open metabase database", func(t *testing.T) {
		integration.NewTester(t, server).
			Post(ctx, service.GrantAccessData{
				DatasetID:   openDataset.ID,
				Expires:     nil,
				Subject:     strToStrPtr(integration.GroupEmailAllUsers),
				SubjectType: strToStrPtr("group"),
			}, "/api/accesses/grant").
			HasStatusCode(httpapi.StatusNoContent)

		time.Sleep(1 * time.Second)

		meta, err := stores.OpenMetabaseStorage.GetMetadata(ctx, openDataset.ID)
		require.NoError(t, err)
		require.NotNil(t, meta.DatabaseID)

		permissionGraphForGroup, err := mbapi.GetPermissionGraphForGroup(ctx, service.MetabaseAllUsersGroupID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Contains(t, permissionGraphForGroup.Groups, strconv.Itoa(service.MetabaseAllUsersGroupID))

		time.Sleep(10 * time.Second)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, openDataset.Datasource.ProjectID, openDataset.Datasource.Dataset, openDataset.Datasource.Table)
		assert.NoError(t, err)
		assert.True(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+integration.MetabaseAllUsersServiceAccount))
	})

	t.Run("Create a public dashboard", func(t *testing.T) {
		expected := service.PublicMetabaseDashboardOutput{
			Description: newPublicMetabaseDashboard.Description,
			TeamID:      newPublicMetabaseDashboard.TeamID,
			Group:       newPublicMetabaseDashboard.Group,
			Keywords:    []string{},
			CreatedBy:   integration.UserOneEmail,
		}

		integration.NewTester(t, server).
			Post(ctx, integration.NewPublicMetabaseDashboardInput(integration.GroupEmailNada, integration.TeamNadaID), "/api/metabaseDashboards/new").
			HasStatusCode(httpapi.StatusOK).Value(&publicDashboard)

		diff := cmp.Diff(expected, publicDashboard, cmpopts.IgnoreFields(service.PublicMetabaseDashboardOutput{}, "Name", "ID", "Link", "Created", "LastModified"))
		assert.Empty(t, diff)

		if !strings.HasPrefix(publicDashboard.Link, mbCfg.PublicHost+"/public/dashboard") {
			t.Errorf("Public dashboard link does not have expected prefix. got: %s, want prefix: %s", publicDashboard.Link, mbCfg.PublicHost+"/public/dashboard")
		}

		publicDashboards, err := mbapi.GetPublicMetabaseDashboards(ctx)
		if err != nil {
			t.Errorf("fetching public dashboards from metabase: %s", err)
		}

		if len(publicDashboards) != 1 {
			t.Errorf("expected %d public metabase dashboards, got %d", 1, len(publicDashboards))
		}

		linkParts := strings.Split(publicDashboard.Link, "/")
		assert.Equal(t, linkParts[len(linkParts)-1], publicDashboards[0].PublicUUID)

		createdDashboardID = publicDashboard.ID
	})

	t.Run("Get public dashboard", func(t *testing.T) {
		expected := service.PublicMetabaseDashboardOutput{
			ID:          createdDashboardID,
			Name:        publicDashboard.Name,
			Description: newPublicMetabaseDashboard.Description,
			TeamID:      newPublicMetabaseDashboard.TeamID,
			Group:       newPublicMetabaseDashboard.Group,
			Keywords:    []string{},
			CreatedBy:   integration.UserOneEmail,
		}

		got := service.PublicMetabaseDashboardOutput{}
		integration.NewTester(t, server).
			Get(ctx, fmt.Sprintf("/api/metabaseDashboards/%s", publicDashboard.ID)).
			HasStatusCode(httpapi.StatusOK).Value(&got)

		diff := cmp.Diff(expected, got, cmpopts.IgnoreFields(service.PublicMetabaseDashboardOutput{}, "Link", "Created", "LastModified"))
		assert.Empty(t, diff)
	})

	t.Run("Edit dashboard metadata", func(t *testing.T) {
		newDesc := "Updated description"

		updateInput := service.PublicMetabaseDashboardEditInput{
			Description: &newDesc,
			Keywords:    []string{"keyword1", "keyword2"},
			TeamID:      newPublicMetabaseDashboard.TeamID,
		}

		expected := service.PublicMetabaseDashboardOutput{
			ID:          createdDashboardID,
			Name:        publicDashboard.Name,
			Description: updateInput.Description,
			TeamID:      newPublicMetabaseDashboard.TeamID,
			Group:       newPublicMetabaseDashboard.Group,
			Keywords:    updateInput.Keywords,
			CreatedBy:   integration.UserOneEmail,
		}

		updatedDashboard := service.PublicMetabaseDashboardOutput{}
		integration.NewTester(t, server).
			Put(ctx, updateInput, fmt.Sprintf("/api/metabaseDashboards/%s", publicDashboard.ID)).
			HasStatusCode(httpapi.StatusOK).Value(&updatedDashboard)

		diff := cmp.Diff(expected, updatedDashboard, cmpopts.IgnoreFields(service.PublicMetabaseDashboardOutput{}, "Link", "Created", "LastModified"))
		assert.Empty(t, diff)
	})

	t.Run("Delete a public dashboard not allowed for another team", func(t *testing.T) {
		integration.NewTester(t, notAllowedServer).
			Delete(ctx, fmt.Sprintf("/api/metabaseDashboards/%s", publicDashboard.ID)).
			HasStatusCode(httpapi.StatusForbidden)
	})

	t.Run("Delete a public dashboard", func(t *testing.T) {
		integration.NewTester(t, server).
			Delete(ctx, fmt.Sprintf("/api/metabaseDashboards/%s", publicDashboard.ID)).
			HasStatusCode(httpapi.StatusNoContent)

		publicDashboards, err := mbapi.GetPublicMetabaseDashboards(ctx)
		if err != nil {
			t.Errorf("fetching public dashboards from metabase: %s", err)
		}

		if len(publicDashboards) != 0 {
			t.Errorf("expected %d public metabase dashboards, got %d", 0, len(publicDashboards))
		}
	})

	t.Run("Create public dashboard in a collection without access is not allowed", func(t *testing.T) {
		adminCollection, err := mbapi.CreateCollectionWithAccess(ctx, metabaseAdminGroupID, &service.CreateCollectionRequest{
			Name:        "Admin collection",
			Description: "All users does not have access to this collection",
		}, true)
		if err != nil {
			t.Errorf("creating collection: %s", err)
		}

		if err := mbapi.UpdateCollection(ctx, &service.MetabaseCollection{
			ID:          metabaseDashboardCollectionID,
			Name:        "Dashboard collection",
			Description: "This collection open for all users, but is located inside the admin collection",
			ParentID:    adminCollection,
		}); err != nil {
			t.Errorf("moving collection: %s", err)
		}

		integration.NewTester(t, server).
			Post(ctx, integration.NewPublicMetabaseDashboardInput(integration.GroupEmailNada, integration.TeamNadaID), "/api/metabaseDashboards/new").
			HasStatusCode(httpapi.StatusForbidden)
	})
}

// // nolint: tparallel,maintidx
//
//	func TestMetabasePublicDashboards(t *testing.T) {
//		t.Parallel()
//
//		ctx := context.Background()
//
//		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(20*time.Minute))
//		defer cancel()
//
//		log := zerolog.New(zerolog.NewConsoleWriter())
//		log.Level(zerolog.DebugLevel)
//
//		gcpHelper := NewGCPHelper(t, log)
//		cleanupFn := gcpHelper.Start(ctx)
//		defer cleanupFn(ctx)
//
//		c := integration.NewContainers(t, log)
//		defer c.Cleanup()
//
//		pgCfg := c.RunPostgres(integration.NewPostgresConfig())
//
//		repo, err := database.New(
//			pgCfg.ConnectionURL(),
//			10,
//			10,
//		)
//		assert.NoError(t, err)
//
//		mbCfg := c.RunMetabase(integration.NewMetabaseConfig(), "../../../.metabase_version")
//
//		stores := storage.NewStores(nil, repo, config.Config{}, log)
//
//		zlog := zerolog.New(os.Stdout)
//		r := integration.TestRouter(zlog)
//
//		mbapi := http.NewMetabaseHTTP(
//			mbCfg.ConnectionURL()+"/api",
//			mbCfg.Email,
//			mbCfg.Password,
//			"",
//			false,
//			false,
//			log,
//		)
//
//		credBytes, err := os.ReadFile("../../../tests-metabase-all-users-sa-creds.json")
//		assert.NoError(t, err)
//
//		_, err = google.CredentialsFromJSON(ctx, credBytes)
//		if err != nil {
//			t.Fatalf("Failed to parse Metabase service account credentials: %v", err)
//		}
//
//		server := httptest.NewServer(r)
//		defer server.Close()
//	}
func TestMetabaseOpenRestrictedDataset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(20*time.Minute))
	defer cancel()

	log := zerolog.New(zerolog.NewConsoleWriter())
	log.Level(zerolog.DebugLevel)

	gcpHelper := NewGCPHelper(t, log)
	cleanupFn := gcpHelper.Start(ctx)
	defer cleanupFn(ctx)

	c := integration.NewContainers(t, log)
	defer c.Cleanup()

	pgCfg := c.RunPostgres(integration.NewPostgresConfig())

	repo, err := database.New(
		pgCfg.ConnectionURL(),
		10,
		10,
	)
	assert.NoError(t, err)

	mbCfg := c.RunMetabase(integration.NewMetabaseConfig(), "../../../.metabase_version")

	bqClient := bq.NewClient("", true, log)
	saClient := dmpSA.NewClient("", false)
	crmClient := crm.NewClient("", false, nil)

	kmsEmulator := emulator.New(log)
	kmsEmulator.AddSymmetricKey(integration.MetabaseProject, integration.Location, integration.Keyring, integration.MetabaseKeyName, []byte("7b483b28d6e67cfd3b9b5813a286c763"))
	kmsURL := kmsEmulator.Run()

	kmsClient := kms.NewClient(kmsURL, true)

	stores := storage.NewStores(nil, repo, config.Config{}, log)

	zlog := zerolog.New(os.Stdout)
	r := integration.TestRouter(zlog)

	crmapi := gcp.NewCloudResourceManagerAPI(crmClient)
	saapi := gcp.NewServiceAccountAPI(saClient)
	bqapi := gcp.NewBigQueryAPI(integration.MetabaseProject, integration.Location, integration.PseudoDataSet, bqClient)
	kmsapi := gcp.NewKMSAPI(kmsClient)

	mbapi := http.NewMetabaseHTTP(
		mbCfg.ConnectionURL()+"/api",
		mbCfg.Email,
		mbCfg.Password,
		"",
		false,
		false,
		log,
	)

	credBytes, err := os.ReadFile("../../../tests-metabase-all-users-sa-creds.json")
	assert.NoError(t, err)

	_, err = google.CredentialsFromJSON(ctx, credBytes)
	if err != nil {
		t.Fatalf("Failed to parse Metabase service account credentials: %v", err)
	}

	// Subscribe to all River events and log them for debugging

	workers := river.NewWorkers()
	riverConfig := worker.RiverConfig(&zlog, workers)
	riverConfig.PeriodicJobs = []*river.PeriodicJob{}
	mbqueue := river2.NewMetabaseQueue(repo, riverConfig)

	mbService := core.NewMetabaseService(
		integration.MetabaseProject,
		integration.Location,
		integration.Keyring,
		integration.MetabaseKeyName,
		string(credBytes),
		integration.MetabaseAllUsersServiceAccount,
		"group:"+integration.GroupEmailAllUsers,
		integration.GroupEmailAllUsers,
		mbqueue,
		kmsapi,
		mbapi,
		bqapi,
		saapi,
		crmapi,
		stores.OpenMetabaseStorage,
		stores.RestrictedMetaBaseStorage,
		stores.BigQueryStorage,
		stores.DataProductsStorage,
		stores.AccessStorage,
		zlog,
	)

	err = worker.MetabaseAddWorkers(riverConfig, mbService, repo)
	require.NoError(t, err)

	riverClient, err := worker.RiverClient(riverConfig, repo)
	require.NoError(t, err)

	err = riverClient.Start(ctx)
	require.NoError(t, err)

	defer riverClient.Stop(ctx)

	// subscribeChan, subscribeCancel := riverClient.Subscribe(
	// 	river.EventKindJobCompleted,
	// 	river.EventKindJobFailed,
	// 	river.EventKindJobCancelled,
	// 	river.EventKindJobSnoozed,
	// 	river.EventKindQueuePaused,
	// 	river.EventKindQueueResumed,
	// )
	// defer subscribeCancel()

	// go func() {
	// 	log.Info().Msg("Starting to listen for River events")
	//
	// 	for {
	// 		select {
	// 		case event := <-subscribeChan:
	// 			if event != nil {
	// 				log.Info().Msgf("Received event: %s", spew.Sdump(event))
	// 			}
	// 		case <-time.After(5 * time.Second):
	// 			log.Info().Msg("No events received in the last 5 seconds")
	// 		}
	// 	}
	// }()

	err = stores.NaisConsoleStorage.UpdateAllTeamProjects(ctx, []*service.NaisTeamMapping{
		{
			Slug:       integration.NaisTeamNada,
			GroupEmail: integration.GroupEmailNada,
			ProjectID:  MetabaseProject,
		},
	})
	assert.NoError(t, err)

	dataproductService := core.NewDataProductsService(
		stores.DataProductsStorage,
		stores.BigQueryStorage,
		bqapi,
		stores.NaisConsoleStorage,
		integration.GroupEmailAllUsers,
	)

	{
		h := handlers.NewMetabaseHandler(mbService)
		e := routes.NewMetabaseEndpoints(zlog, h)
		f := routes.NewMetabaseRoutes(e, integration.InjectUser(integration.UserOne))

		f(r)
	}

	slack := static.NewSlackAPI(log)
	accessService := core.NewAccessService(
		"",
		slack,
		stores.PollyStorage,
		stores.AccessStorage,
		stores.DataProductsStorage,
		stores.BigQueryStorage,
		stores.JoinableViewsStorage,
		bqapi,
	)

	{
		h := handlers.NewAccessHandler(accessService, mbService, MetabaseProject)
		e := routes.NewAccessEndpoints(zlog, h)
		f := routes.NewAccessRoutes(e, integration.InjectUser(integration.UserOne))

		f(r)
	}

	server := httptest.NewServer(r)
	defer server.Close()

	integration.StorageCreateProductAreasAndTeams(t, stores.ProductAreaStorage)
	dataproduct, err := dataproductService.CreateDataproduct(ctx, integration.UserOne, integration.NewDataProductBiofuelProduction(integration.GroupEmailNada, integration.TeamSeagrassID))
	require.NoError(t, err)

	restrictedDataset2, err := dataproductService.CreateDataset(ctx, integration.UserOne, service.NewDataset{
		DataproductID: dataproduct.ID,
		Name:          "Another restricted dataset",
		BigQuery:      gcpHelper.BigQueryTable,
		Pii:           service.PiiLevelNone,
	})
	require.NoError(t, err)

	var restrictedMeta *service.RestrictedMetabaseMetadata
	t.Run("Add a restricted metabase database", func(t *testing.T) {
		integration.NewTester(t, server).
			Post(ctx, nil, fmt.Sprintf("/api/datasets/%s/bigquery_restricted", restrictedDataset2.ID)).
			HasStatusCode(httpapi.StatusOK)

		status := &service.MetabaseBigQueryDatasetStatus{}
		for {
			integration.NewTester(t, server).
				Get(ctx, fmt.Sprintf("/api/datasets/%s/bigquery_restricted", restrictedDataset2.ID)).
				HasStatusCode(httpapi.StatusOK).
				Value(status)

			if status.HasFailed {
				t.Fatalf("Failed to add open dataset to Metabase: %s", status.Error())
			}

			if err := status.Error(); err != nil {
				fmt.Println("Error: ", err)
			}

			if !status.IsCompleted {
				time.Sleep(5 * time.Second)
				continue
			}

			break
		}

		restrictedMeta, err = stores.RestrictedMetaBaseStorage.GetMetadata(ctx, restrictedDataset2.ID)
		require.NoError(t, err)
		require.NotNil(t, restrictedMeta.SyncCompleted)

		collections, err := mbapi.GetCollections(ctx)
		require.NoError(t, err)
		assert.True(t, integration.ContainsCollectionWithName(collections, "Another restricted dataset 🔐"))

		permissionGroups, err := mbapi.GetPermissionGroups(ctx)
		require.NoError(t, err)
		assert.True(t, integration.ContainsPermissionGroupWithNamePrefix(permissionGroups, "another-restricted-dataset"))

		sa, err := saClient.GetServiceAccount(ctx, fmt.Sprintf("projects/%s/serviceAccounts/%s", MetabaseProject, restrictedMeta.SAEmail))
		require.NoError(t, err)
		assert.True(t, sa.Email == restrictedMeta.SAEmail)

		keys, err := saClient.ListServiceAccountKeys(ctx, fmt.Sprintf("projects/%s/serviceAccounts/%s", MetabaseProject, restrictedMeta.SAEmail))
		require.NoError(t, err)
		assert.Len(t, keys, 2) // will return 1 system managed key (always) in addition to the user managed key we created

		bindings, err := crmClient.ListProjectIAMPolicyBindings(ctx, MetabaseProject, "serviceAccount:"+restrictedMeta.SAEmail)
		require.NoError(t, err)
		assert.True(t, integration.ContainsProjectIAMPolicyBindingForSubject(bindings, NadaMetabaseRole, "serviceAccount:"+restrictedMeta.SAEmail))

		require.NotNil(t, restrictedMeta.PermissionGroupID)
		permissionGraphForGroup, err := mbapi.GetPermissionGraphForGroup(ctx, *restrictedMeta.PermissionGroupID)
		require.NoError(t, err)

		assert.Contains(t, permissionGraphForGroup.Groups, strconv.Itoa(*restrictedMeta.PermissionGroupID))
		assert.Equal(t, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset2.ID), restrictedMeta.SAEmail)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, restrictedDataset2.Datasource.ProjectID, restrictedDataset2.Datasource.Dataset, restrictedDataset2.Datasource.Table)
		assert.NoError(t, err)
		assert.True(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset2.ID)))
		assert.False(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+integration.MetabaseAllUsersServiceAccount))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, restrictedDataset2.Datasource.Dataset)
		assert.NoError(t, err)
		assert.True(t, integration.ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset2.ID)))
	})

	t.Run("Opening a previously restricted metabase database", func(t *testing.T) {
		integration.NewTester(t, server).
			Post(ctx, nil, fmt.Sprintf("/api/datasets/%s/bigquery_open_restricted", restrictedDataset2.ID)).
			HasStatusCode(httpapi.StatusNoContent)

		_, err := stores.RestrictedMetaBaseStorage.GetMetadata(ctx, restrictedDataset2.ID)
		require.Error(t, err)

		_, err = stores.OpenMetabaseStorage.GetMetadata(ctx, restrictedDataset2.ID)
		require.NoError(t, err)

		collections, err := mbapi.GetCollections(ctx)
		require.NoError(t, err)
		assert.True(t, integration.ContainsCollectionWithName(collections, "Another restricted dataset"))

		permissionGroups, err := mbapi.GetPermissionGroups(ctx)
		require.NoError(t, err)
		assert.True(t, !integration.ContainsPermissionGroupWithNamePrefix(permissionGroups, "another-restricted-dataset"))

		var collectionID string
		for _, c := range collections {
			if c.Name == "Another restricted dataset" {
				collectionID = strconv.Itoa(c.ID)
				break
			}
		}

		collectionPermissions, err := mbapi.GetCollectionPermissions(ctx)
		require.NoError(t, err)

		allUsersCollectionAccess := collectionPermissions.Groups[strconv.Itoa(service.MetabaseAllUsersGroupID)]
		assert.True(t, allUsersCollectionAccess[collectionID] == "write")

		tablePolicy, err := bqClient.GetTablePolicy(ctx, restrictedDataset2.Datasource.ProjectID, restrictedDataset2.Datasource.Dataset, restrictedDataset2.Datasource.Table)
		assert.NoError(t, err)
		assert.False(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset2.ID)))
		assert.True(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+integration.MetabaseAllUsersServiceAccount))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, restrictedDataset2.Datasource.Dataset)
		assert.NoError(t, err)
		assert.True(t, integration.ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, integration.MetabaseAllUsersServiceAccount))
	})
}

// nolint: tparallel,maintidx
func TestMetabaseRestrictedDataset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(20*time.Minute))
	defer cancel()

	log := zerolog.New(zerolog.NewConsoleWriter())
	log.Level(zerolog.DebugLevel)

	gcpHelper := NewGCPHelper(t, log)
	cleanupFn := gcpHelper.Start(ctx)
	defer cleanupFn(ctx)

	c := integration.NewContainers(t, log)
	defer c.Cleanup()

	pgCfg := c.RunPostgres(integration.NewPostgresConfig())

	repo, err := database.New(
		pgCfg.ConnectionURL(),
		10,
		10,
	)
	assert.NoError(t, err)

	mbCfg := c.RunMetabase(integration.NewMetabaseConfig(), "../../../.metabase_version")

	bqClient := bq.NewClient("", true, log)
	saClient := dmpSA.NewClient("", false)
	crmClient := crm.NewClient("", false, nil)

	kmsEmulator := emulator.New(log)
	kmsEmulator.AddSymmetricKey(integration.MetabaseProject, integration.Location, integration.Keyring, integration.MetabaseKeyName, []byte("7b483b28d6e67cfd3b9b5813a286c763"))
	kmsURL := kmsEmulator.Run()

	kmsClient := kms.NewClient(kmsURL, true)

	stores := storage.NewStores(nil, repo, config.Config{}, log)

	zlog := zerolog.New(os.Stdout)
	r := integration.TestRouter(zlog)

	crmapi := gcp.NewCloudResourceManagerAPI(crmClient)
	saapi := gcp.NewServiceAccountAPI(saClient)
	bqapi := gcp.NewBigQueryAPI(integration.MetabaseProject, integration.Location, integration.PseudoDataSet, bqClient)
	kmsapi := gcp.NewKMSAPI(kmsClient)

	mbapi := http.NewMetabaseHTTP(
		mbCfg.ConnectionURL()+"/api",
		mbCfg.Email,
		mbCfg.Password,
		"",
		false,
		false,
		log,
	)

	credBytes, err := os.ReadFile("../../../tests-metabase-all-users-sa-creds.json")
	assert.NoError(t, err)

	_, err = google.CredentialsFromJSON(ctx, credBytes)
	if err != nil {
		t.Fatalf("Failed to parse Metabase service account credentials: %v", err)
	}

	// Subscribe to all River events and log them for debugging

	workers := river.NewWorkers()
	riverConfig := worker.RiverConfig(&zlog, workers)
	riverConfig.PeriodicJobs = []*river.PeriodicJob{}
	mbqueue := river2.NewMetabaseQueue(repo, riverConfig)

	mbService := core.NewMetabaseService(
		integration.MetabaseProject,
		integration.Location,
		integration.Keyring,
		integration.MetabaseKeyName,
		string(credBytes),
		integration.MetabaseAllUsersServiceAccount,
		"group:"+integration.GroupEmailAllUsers,
		integration.GroupEmailAllUsers,
		mbqueue,
		kmsapi,
		mbapi,
		bqapi,
		saapi,
		crmapi,
		stores.OpenMetabaseStorage,
		stores.RestrictedMetaBaseStorage,
		stores.BigQueryStorage,
		stores.DataProductsStorage,
		stores.AccessStorage,
		zlog,
	)

	err = worker.MetabaseAddWorkers(riverConfig, mbService, repo)
	require.NoError(t, err)

	riverClient, err := worker.RiverClient(riverConfig, repo)
	require.NoError(t, err)

	err = riverClient.Start(ctx)
	require.NoError(t, err)

	defer riverClient.Stop(ctx)

	// subscribeChan, subscribeCancel := riverClient.Subscribe(
	// 	river.EventKindJobCompleted,
	// 	river.EventKindJobFailed,
	// 	river.EventKindJobCancelled,
	// 	river.EventKindJobSnoozed,
	// 	river.EventKindQueuePaused,
	// 	river.EventKindQueueResumed,
	// )
	// defer subscribeCancel()

	// go func() {
	// 	log.Info().Msg("Starting to listen for River events")
	//
	// 	for {
	// 		select {
	// 		case event := <-subscribeChan:
	// 			if event != nil {
	// 				log.Info().Msgf("Received event: %s", spew.Sdump(event))
	// 			}
	// 		case <-time.After(5 * time.Second):
	// 			log.Info().Msg("No events received in the last 5 seconds")
	// 		}
	// 	}
	// }()

	err = stores.NaisConsoleStorage.UpdateAllTeamProjects(ctx, []*service.NaisTeamMapping{
		{
			Slug:       integration.NaisTeamNada,
			GroupEmail: integration.GroupEmailNada,
			ProjectID:  MetabaseProject,
		},
	})
	assert.NoError(t, err)

	dataproductService := core.NewDataProductsService(
		stores.DataProductsStorage,
		stores.BigQueryStorage,
		bqapi,
		stores.NaisConsoleStorage,
		integration.GroupEmailAllUsers,
	)

	{
		h := handlers.NewMetabaseHandler(mbService)
		e := routes.NewMetabaseEndpoints(zlog, h)
		f := routes.NewMetabaseRoutes(e, integration.InjectUser(integration.UserOne))

		f(r)
	}

	slack := static.NewSlackAPI(log)
	accessService := core.NewAccessService(
		"",
		slack,
		stores.PollyStorage,
		stores.AccessStorage,
		stores.DataProductsStorage,
		stores.BigQueryStorage,
		stores.JoinableViewsStorage,
		bqapi,
	)

	{
		h := handlers.NewAccessHandler(accessService, mbService, MetabaseProject)
		e := routes.NewAccessEndpoints(zlog, h)
		f := routes.NewAccessRoutes(e, integration.InjectUser(integration.UserOne))

		f(r)
	}

	server := httptest.NewServer(r)
	defer server.Close()

	integration.StorageCreateProductAreasAndTeams(t, stores.ProductAreaStorage)
	dataproduct, err := dataproductService.CreateDataproduct(ctx, integration.UserOne, integration.NewDataProductBiofuelProduction(integration.GroupEmailNada, integration.TeamSeagrassID))
	require.NoError(t, err)

	restrictedDataset, err := dataproductService.CreateDataset(ctx, integration.UserOne, service.NewDataset{
		DataproductID: dataproduct.ID,
		Name:          "Restricted dataset",
		BigQuery:      gcpHelper.BigQueryTable,
		Pii:           service.PiiLevelNone,
	})
	require.NoError(t, err)

	t.Run("Add restricted dataset to metabase", func(t *testing.T) {
		integration.NewTester(t, server).
			Post(ctx, nil, fmt.Sprintf("/api/datasets/%s/bigquery_restricted", restrictedDataset.ID)).
			HasStatusCode(httpapi.StatusOK)

		status := &service.MetabaseBigQueryDatasetStatus{}
		for {
			integration.NewTester(t, server).
				Get(ctx, fmt.Sprintf("/api/datasets/%s/bigquery_restricted", restrictedDataset.ID)).
				HasStatusCode(httpapi.StatusOK).
				Value(status)

			if status.HasFailed {
				t.Fatalf("Failed to add open dataset to Metabase: %s", status.Error())
			}

			if err := status.Error(); err != nil {
				fmt.Println("Error: ", err)
			}

			if !status.IsCompleted {
				time.Sleep(5 * time.Second)
				continue
			}

			break
		}

		meta, err := stores.RestrictedMetaBaseStorage.GetMetadata(ctx, restrictedDataset.ID)
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		collections, err := mbapi.GetCollections(ctx)
		require.NoError(t, err)
		assert.True(t, integration.ContainsCollectionWithName(collections, "Restricted dataset 🔐"))

		permissionGroups, err := mbapi.GetPermissionGroups(ctx)
		require.NoError(t, err)
		assert.True(t, integration.ContainsPermissionGroupWithNamePrefix(permissionGroups, "restricted-dataset"))

		sa, err := saClient.GetServiceAccount(ctx, fmt.Sprintf("projects/%s/serviceAccounts/%s", MetabaseProject, meta.SAEmail))
		require.NoError(t, err)
		assert.True(t, sa.Email == meta.SAEmail)

		keys, err := saClient.ListServiceAccountKeys(ctx, fmt.Sprintf("projects/%s/serviceAccounts/%s", MetabaseProject, meta.SAEmail))
		require.NoError(t, err)
		assert.Len(t, keys, 2) // will return 1 system managed key (always) in addition to the user managed key we created

		bindings, err := crmClient.ListProjectIAMPolicyBindings(ctx, MetabaseProject, "serviceAccount:"+meta.SAEmail)
		require.NoError(t, err)
		assert.True(t, integration.ContainsProjectIAMPolicyBindingForSubject(bindings, NadaMetabaseRole, "serviceAccount:"+meta.SAEmail))

		require.NotNil(t, meta.PermissionGroupID)
		permissionGraphForGroup, err := mbapi.GetPermissionGraphForGroup(ctx, *meta.PermissionGroupID)
		if err != nil {
			t.Fatal(err)
		}
		// When adding a restricted database to metabase the corresponding permission group
		// should be granted access to only this database
		assert.Equal(t, numberOfDatabasesWithAccessForPermissionGroup(permissionGraphForGroup.Groups[strconv.Itoa(*meta.PermissionGroupID)]), 1)

		assert.Contains(t, permissionGraphForGroup.Groups, strconv.Itoa(*meta.PermissionGroupID))
		assert.Equal(t, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset.ID), meta.SAEmail)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, restrictedDataset.Datasource.ProjectID, restrictedDataset.Datasource.Dataset, restrictedDataset.Datasource.Table)
		assert.NoError(t, err)
		assert.True(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset.ID)))
		assert.False(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+integration.MetabaseAllUsersServiceAccount))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, restrictedDataset.Datasource.Dataset)
		assert.NoError(t, err)
		assert.True(t, integration.ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset.ID)))
	})

	t.Run("Clear metabase bigquery jobs", func(t *testing.T) {
		integration.NewTester(t, server).
			Delete(ctx, fmt.Sprintf("/api/datasets/%s/bigquery_jobs", restrictedDataset.ID)).
			HasStatusCode(httpapi.StatusNoContent)
	})

	t.Run("Removing 🔐 is added back", func(t *testing.T) {
		meta, err := stores.RestrictedMetaBaseStorage.GetMetadata(ctx, restrictedDataset.ID)
		require.NoError(t, err)

		err = mbapi.UpdateCollection(ctx, &service.MetabaseCollection{
			ID:          *meta.CollectionID,
			Name:        "My new collection name",
			Description: "My new collection description",
		})
		require.NoError(t, err)

		err = metabase_collections.New(mbapi, stores.RestrictedMetaBaseStorage, stores.DataProductsStorage).RunOnce(ctx, log)
		require.NoError(t, err)

		collections, err := mbapi.GetCollections(ctx)
		require.NoError(t, err)
		assert.True(t, integration.ContainsCollectionWithName(collections, "Restricted dataset 🔐"))
	})

	t.Run("Delete restricted metabase database", func(t *testing.T) {
		integration.NewTester(t, server).
			Delete(ctx, fmt.Sprintf("/api/datasets/%s/bigquery_restricted", restrictedDataset.ID)).
			HasStatusCode(httpapi.StatusNoContent)

		_, err = stores.RestrictedMetaBaseStorage.GetMetadata(ctx, restrictedDataset.ID)
		require.Error(t, err)

		collections, err := mbapi.GetCollections(ctx)
		require.NoError(t, err)
		assert.False(t, integration.ContainsCollectionWithName(collections, "Restricted dataset 2 🔐"))

		permissionGroups, err := mbapi.GetPermissionGroups(ctx)
		require.NoError(t, err)
		assert.False(t, integration.ContainsPermissionGroupWithNamePrefix(permissionGroups, "restricted-dataset-2"))

		// Need to ensure that the service account actually is deleted
		for i := 0; i < 60; i++ {
			_, err = saClient.GetServiceAccount(ctx, fmt.Sprintf("projects/%s/serviceAccounts/%s", MetabaseProject, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset.ID)))
			if err != nil {
				break
			}
			time.Sleep(time.Second)
		}
		require.Error(t, err)

		bindings, err := crmClient.ListProjectIAMPolicyBindings(ctx, MetabaseProject, "serviceAccount:"+mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset.ID))
		require.NoError(t, err)
		assert.False(t, integration.ContainsProjectIAMPolicyBindingForSubject(bindings, NadaMetabaseRole, "serviceAccount:"+mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset.ID)))

		tablePolicy, err := bqClient.GetTablePolicy(ctx, restrictedDataset.Datasource.ProjectID, restrictedDataset.Datasource.Dataset, restrictedDataset.Datasource.Table)
		assert.NoError(t, err)
		assert.False(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset.ID)))
	})
}
