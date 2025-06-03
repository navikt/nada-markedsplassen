package integration_metabase

import (
	"context"
	"fmt"
	"github.com/navikt/nada-backend/pkg/kms"
	"github.com/navikt/nada-backend/pkg/kms/emulator"
	river2 "github.com/navikt/nada-backend/pkg/service/core/queue/river"
	"github.com/navikt/nada-backend/pkg/worker"
	"github.com/riverqueue/river"
	httpapi "net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
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
	"github.com/navikt/nada-backend/pkg/syncers/metabase_collections"
	"github.com/navikt/nada-backend/test/integration"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testRunBigQueryDataset string

func TestMain(m *testing.M) {
	var err error

	ctx := context.Background()
	log := zerolog.New(zerolog.NewConsoleWriter())
	log.Info().Msg("Running Metabase integration tests")

	testRunBigQueryDataset, err = prepareTestProject(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to prepare test project")
	}

	code := m.Run()

	err = cleanupAfterTestRun(ctx, testRunBigQueryDataset)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to clean up after test run")
	}

	os.Exit(code)
}

// nolint: tparallel,maintidx
func TestMetabaseOpenDataset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(20*time.Minute))
	defer cancel()

	log := zerolog.New(zerolog.NewConsoleWriter())
	log.Level(zerolog.DebugLevel)

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

	kmsClient := kms.NewClient("", kmsURL, true)

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

	workers := river.NewWorkers()
	riverConfig := worker.RiverConfig(&zlog, workers)
	mbqueue := river2.NewMetabaseQueue(repo, riverConfig)

	mbService := core.NewMetabaseService(
		integration.MetabaseProject,
		integration.Location,
		integration.Keyring,
		integration.MetabaseKeyName,
		string(credBytes),
		integration.MetabaseAllUsersServiceAccount,
		"group:"+integration.GroupEmailAllUsers,
		mbqueue,
		kmsapi,
		mbapi,
		bqapi,
		saapi,
		crmapi,
		stores.MetaBaseStorage,
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

	server := httptest.NewServer(r)
	defer server.Close()

	// Prepare BigQuery table for test run
	bqTable, err := createBigQueryTable(ctx, testRunBigQueryDataset, "open_table")
	assert.NoError(t, err)

	integration.StorageCreateProductAreasAndTeams(t, stores.ProductAreaStorage)
	dataproduct, err := dataproductService.CreateDataproduct(ctx, integration.UserOne, integration.NewDataProductBiofuelProduction(integration.GroupEmailNada, integration.TeamSeagrassID))
	assert.NoError(t, err)

	openDataset, err := dataproductService.CreateDataset(ctx, integration.UserOne, service.NewDataset{
		DataproductID: dataproduct.ID,
		Name:          "Open dataset",
		BigQuery:      bqTable,
		Pii:           service.PiiLevelNone,
	})
	assert.NoError(t, err)

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
				fmt.Println("Status: ", spew.Sdump(status))
				t.Fatalf("Failed to add open dataset to Metabase: %s", status.Error())
			}

			if !status.IsCompleted {
				time.Sleep(5 * time.Second)
				continue
			}

			break
		}

		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, openDataset.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.DatabaseID)
		require.NotNil(t, meta.SyncCompleted)

		permissionGraphForGroup, err := mbapi.GetPermissionGraphForGroup(ctx, service.MetabaseAllUsersGroupID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, permissionGraphForGroup.Groups, strconv.Itoa(service.MetabaseAllUsersGroupID))
		assert.Equal(t, MetabaseAllUsersServiceAccount, meta.SAEmail)

		spew.Dump(permissionGraphForGroup.Groups)

		// When adding an open dataset to metabase the all users group should be granted access
		// while not losing access to the default open "sample dataset" database
		assert.Equal(t, numberOfDatabasesWithAccessForPermissionGroup(permissionGraphForGroup.Groups[strconv.Itoa(service.MetabaseAllUsersGroupID)]), 2)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, openDataset.Datasource.ProjectID, openDataset.Datasource.Dataset, openDataset.Datasource.Table)
		assert.NoError(t, err)
		assert.True(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, openDataset.Datasource.Dataset)
		assert.NoError(t, err)
		assert.True(t, integration.ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, MetabaseAllUsersServiceAccount))
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

		time.Sleep(1 * time.Second)

		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, openDataset.ID, true)
		require.NoError(t, err)
		require.NotNil(t, meta.DatabaseID)

		permissionGraphForGroup, err := mbapi.GetPermissionGraphForGroup(ctx, service.MetabaseAllUsersGroupID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Contains(t, permissionGraphForGroup.Groups, strconv.Itoa(service.MetabaseAllUsersGroupID))
		assert.Equal(t, MetabaseAllUsersServiceAccount, meta.SAEmail)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, openDataset.Datasource.ProjectID, openDataset.Datasource.Dataset, openDataset.Datasource.Table)
		assert.NoError(t, err)
		assert.False(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))
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

		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, openDataset.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.DatabaseID)

		permissionGraphForGroup, err := mbapi.GetPermissionGraphForGroup(ctx, service.MetabaseAllUsersGroupID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Contains(t, permissionGraphForGroup.Groups, strconv.Itoa(service.MetabaseAllUsersGroupID))
		assert.Equal(t, MetabaseAllUsersServiceAccount, meta.SAEmail)

		time.Sleep(10 * time.Second)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, openDataset.Datasource.ProjectID, openDataset.Datasource.Dataset, openDataset.Datasource.Table)
		fmt.Println("Table policy: ", spew.Sdump(tablePolicy))
		assert.NoError(t, err)
		assert.True(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))
	})

	t.Run("Permanent delete of open metabase database", func(t *testing.T) {
		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, openDataset.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		integration.NewTester(t, server).
			Delete(ctx, fmt.Sprintf("/api/datasets/%s/bigquery_open", openDataset.ID)).
			HasStatusCode(httpapi.StatusNoContent)

		time.Sleep(500 * time.Millisecond)

		_, err = stores.MetaBaseStorage.GetMetadata(ctx, openDataset.ID, false)
		require.Error(t, err)

		_, err = mbapi.Database(ctx, *meta.DatabaseID)
		require.Error(t, err)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, openDataset.Datasource.ProjectID, openDataset.Datasource.Dataset, openDataset.Datasource.Table)
		assert.NoError(t, err)
		assert.False(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, openDataset.Datasource.Dataset)
		assert.NoError(t, err)
		// Dataset Metadata Viewer is intentionally not removed when access for table is revoked so should be true
		assert.True(t, integration.ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, meta.SAEmail))
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

	kmsClient := kms.NewClient("", kmsURL, true)

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

	// Subscribe to all River events and log them for debugging

	workers := river.NewWorkers()
	riverConfig := worker.RiverConfig(&zlog, workers)
	mbqueue := river2.NewMetabaseQueue(repo, riverConfig)

	mbService := core.NewMetabaseService(
		integration.MetabaseProject,
		integration.Location,
		integration.Keyring,
		integration.MetabaseKeyName,
		string(credBytes),
		integration.MetabaseAllUsersServiceAccount,
		"group:"+integration.GroupEmailAllUsers,
		mbqueue,
		kmsapi,
		mbapi,
		bqapi,
		saapi,
		crmapi,
		stores.MetaBaseStorage,
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

	// Prepare BigQuery table for test run
	bqTable, err := createBigQueryTable(ctx, testRunBigQueryDataset, "restricted_table")
	require.NoError(t, err)

	integration.StorageCreateProductAreasAndTeams(t, stores.ProductAreaStorage)
	dataproduct, err := dataproductService.CreateDataproduct(ctx, integration.UserOne, integration.NewDataProductBiofuelProduction(integration.GroupEmailNada, integration.TeamSeagrassID))
	require.NoError(t, err)

	restrictedDataset, err := dataproductService.CreateDataset(ctx, integration.UserOne, service.NewDataset{
		DataproductID: dataproduct.ID,
		Name:          "Restricted dataset",
		BigQuery:      bqTable,
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
				fmt.Println("Status: ", spew.Sdump(status))
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

		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, restrictedDataset.ID, false)
		fmt.Println("Meta: ", spew.Sdump(meta))
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		collections, err := mbapi.GetCollections(ctx)
		fmt.Println("Collections: ", spew.Sdump(collections))
		require.NoError(t, err)
		assert.True(t, integration.ContainsCollectionWithName(collections, "Restricted dataset 🔐"))

		permissionGroups, err := mbapi.GetPermissionGroups(ctx)
		fmt.Println("Permission groups: ", spew.Sdump(permissionGroups))
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
		assert.False(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, restrictedDataset.Datasource.Dataset)
		assert.NoError(t, err)
		assert.True(t, integration.ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset.ID)))
	})

	t.Run("Delete restricted metabase database", func(t *testing.T) {
		integration.NewTester(t, server).
			Delete(ctx, fmt.Sprintf("/api/datasets/%s/bigquery_restricted", restrictedDataset.ID)).
			HasStatusCode(httpapi.StatusNoContent)

		_, err = stores.MetaBaseStorage.GetMetadata(ctx, restrictedDataset.ID, true)
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

// nolint: tparallel,maintidx
func TestMetabaseOpeningRestrictedDataset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(20*time.Minute))
	defer cancel()

	log := zerolog.New(zerolog.NewConsoleWriter())
	log.Level(zerolog.DebugLevel)

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

	kmsClient := kms.NewClient("", kmsURL, true)

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

	workers := river.NewWorkers()
	riverConfig := worker.RiverConfig(&zlog, workers)
	mbqueue := river2.NewMetabaseQueue(repo, riverConfig)

	mbService := core.NewMetabaseService(
		integration.MetabaseProject,
		integration.Location,
		integration.Keyring,
		integration.MetabaseKeyName,
		string(credBytes),
		integration.MetabaseAllUsersServiceAccount,
		"group:"+integration.GroupEmailAllUsers,
		mbqueue,
		kmsapi,
		mbapi,
		bqapi,
		saapi,
		crmapi,
		stores.MetaBaseStorage,
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

	server := httptest.NewServer(r)
	defer server.Close()

	// Prepare BigQuery table for test run
	bqTable, err := createBigQueryTable(ctx, testRunBigQueryDataset, "restricted_to_open_table")
	assert.NoError(t, err)

	integration.StorageCreateProductAreasAndTeams(t, stores.ProductAreaStorage)
	dataproduct, err := dataproductService.CreateDataproduct(ctx, integration.UserOne, integration.NewDataProductBiofuelProduction(integration.GroupEmailNada, integration.TeamSeagrassID))
	assert.NoError(t, err)

	restrictedDataset, err := dataproductService.CreateDataset(ctx, integration.UserOne, service.NewDataset{
		DataproductID: dataproduct.ID,
		Name:          "Restricted dataset",
		BigQuery:      bqTable,
		Pii:           service.PiiLevelNone,
	})
	assert.NoError(t, err)

	t.Run("Add a restricted metabase database", func(t *testing.T) {
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
				fmt.Println("Status: ", spew.Sdump(status))
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

		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, restrictedDataset.ID, false)
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
		require.NoError(t, err)

		assert.Contains(t, permissionGraphForGroup.Groups, strconv.Itoa(*meta.PermissionGroupID))
		assert.Equal(t, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset.ID), meta.SAEmail)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, restrictedDataset.Datasource.ProjectID, restrictedDataset.Datasource.Dataset, restrictedDataset.Datasource.Table)
		assert.NoError(t, err)
		assert.True(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset.ID)))
		assert.False(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, restrictedDataset.Datasource.Dataset)
		assert.NoError(t, err)
		assert.True(t, integration.ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset.ID)))
	})

	t.Run("Removing 🔐 is added back", func(t *testing.T) {
		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, restrictedDataset.ID, false)
		require.NoError(t, err)

		err = mbapi.UpdateCollection(ctx, &service.MetabaseCollection{
			ID:          *meta.CollectionID,
			Name:        "My new collection name",
			Description: "My new collection description",
		})
		require.NoError(t, err)

		err = metabase_collections.New(mbapi, stores.MetaBaseStorage).RunOnce(ctx, log)
		require.NoError(t, err)

		collections, err := mbapi.GetCollections(ctx)
		require.NoError(t, err)
		assert.True(t, integration.ContainsCollectionWithName(collections, "My new collection name 🔐"))
	})

	t.Run("Opening a previously restricted metabase dataset", func(t *testing.T) {
		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, restrictedDataset.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		permissionGraphForGroup, err := mbapi.GetPermissionGraphForGroup(ctx, service.MetabaseAllUsersGroupID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Contains(t, permissionGraphForGroup.Groups, strconv.Itoa(service.MetabaseAllUsersGroupID))

		integration.NewTester(t, server).
			Post(ctx, service.GrantAccessData{
				DatasetID:   restrictedDataset.ID,
				Expires:     nil,
				Subject:     strToStrPtr(integration.GroupEmailAllUsers),
				SubjectType: strToStrPtr("group"),
			}, "/api/accesses/grant").
			HasStatusCode(httpapi.StatusNoContent)

		time.Sleep(1000 * time.Millisecond)

		meta, err = stores.MetaBaseStorage.GetMetadata(ctx, restrictedDataset.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		permissionGroups, err := mbapi.GetPermissionGroups(ctx)
		require.NoError(t, err)
		assert.False(t, integration.ContainsPermissionGroupWithNamePrefix(permissionGroups, "restricted-dataset"))
		assert.Equal(t, MetabaseAllUsersServiceAccount, meta.SAEmail)

		// Need to ensure that the service account actually is deleted
		for i := 0; i < 60; i++ {
			_, err = saClient.GetServiceAccount(ctx, fmt.Sprintf("projects/%s/serviceAccounts/%s", MetabaseProject, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset.ID)))
			if err != nil {
				break
			}
			time.Sleep(time.Second)
		}
		require.Error(t, err)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, restrictedDataset.Datasource.ProjectID, restrictedDataset.Datasource.Dataset, restrictedDataset.Datasource.Table)
		assert.NoError(t, err)
		assert.True(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))
		assert.False(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+mbService.ConstantServiceAccountEmailFromDatasetID(restrictedDataset.ID)))
	})

	t.Run("Delete previously restricted metabase database", func(t *testing.T) {
		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, restrictedDataset.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		integration.NewTester(t, server).
			Delete(ctx, fmt.Sprintf("/api/datasets/%s/bigquery_open", restrictedDataset.ID)).
			HasStatusCode(httpapi.StatusNoContent)

		time.Sleep(500 * time.Millisecond)

		_, err = stores.MetaBaseStorage.GetMetadata(ctx, restrictedDataset.ID, false)
		require.Error(t, err)

		_, err = mbapi.Database(ctx, *meta.DatabaseID)
		require.Error(t, err)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, restrictedDataset.Datasource.ProjectID, restrictedDataset.Datasource.Dataset, restrictedDataset.Datasource.Table)
		assert.NoError(t, err)
		assert.False(t, integration.ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, restrictedDataset.Datasource.Dataset)
		assert.NoError(t, err)
		// Dataset Metadata Viewer is intentionally not removed when access for table is revoked so should be true
		assert.True(t, integration.ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, meta.SAEmail))
	})
}
