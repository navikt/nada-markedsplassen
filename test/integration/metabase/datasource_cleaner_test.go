package integration_metabase

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/navikt/nada-backend/pkg/kms"
	"github.com/navikt/nada-backend/pkg/kms/emulator"
	"github.com/navikt/nada-backend/pkg/service/core/api/http"
	river2 "github.com/navikt/nada-backend/pkg/service/core/queue/river"
	"github.com/navikt/nada-backend/pkg/worker"
	"github.com/navikt/nada-backend/test/integration"
	"github.com/riverqueue/river"
	"golang.org/x/oauth2/google"
	httpapi "net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	crm "github.com/navikt/nada-backend/pkg/cloudresourcemanager"

	"github.com/navikt/nada-backend/pkg/syncers/bigquery_datasource_missing"

	"github.com/navikt/nada-backend/pkg/config/v2"
	dmpSA "github.com/navikt/nada-backend/pkg/sa"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core/api/static"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/navikt/nada-backend/pkg/bq"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service/core"
	"github.com/navikt/nada-backend/pkg/service/core/api/gcp"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/routes"
	"github.com/navikt/nada-backend/pkg/service/core/storage"
	"github.com/rs/zerolog"
)

func TestBigQueryDatasourceCleaner(t *testing.T) {
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
	require.NoError(t, err)

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
	require.NoError(t, err)

	_, err = google.CredentialsFromJSON(ctx, credBytes)
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
	require.NoError(t, err)

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
	bqTable, err := createBigQueryTable(ctx, testRunBigQueryDataset, "bigquery_datasource_cleaner_test")
	require.NoError(t, err)

	integration.StorageCreateProductAreasAndTeams(t, stores.ProductAreaStorage)
	dataproduct, err := dataproductService.CreateDataproduct(ctx, integration.UserOne, integration.NewDataProductBiofuelProduction(integration.GroupEmailNada, integration.TeamSeagrassID))
	require.NoError(t, err)

	openDataset, err := dataproductService.CreateDataset(ctx, integration.UserOne, service.NewDataset{
		DataproductID: dataproduct.ID,
		Name:          "Open dataset",
		BigQuery:      bqTable,
		Pii:           service.PiiLevelNone,
	})
	require.NoError(t, err)

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

	t.Run("Removing datasource with metabase works", func(t *testing.T) {
		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, openDataset.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.DatabaseID)
		require.NotNil(t, meta.SyncCompleted)

		source, err := stores.BigQueryStorage.GetBigqueryDatasource(ctx, openDataset.ID, false)
		require.NoError(t, err)

		err = bqClient.DeleteTable(ctx, source.ProjectID, source.Dataset, source.Table)
		require.NoError(t, err)

		err = bigquery_datasource_missing.New(bqClient, stores.BigQueryStorage, mbService, stores.MetaBaseStorage, stores.DataProductsStorage).RunOnce(ctx, log)
		require.NoError(t, err)

		_, err = stores.BigQueryStorage.GetBigqueryDatasource(ctx, openDataset.ID, false)
		require.Error(t, err)

		_, err = stores.MetaBaseStorage.GetMetadata(ctx, openDataset.ID, true)
		require.Error(t, err)
	})

}
