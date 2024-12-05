package integration

import (
	"context"
	"errors"
	"fmt"
	http2 "net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/iam"
	crm "github.com/navikt/nada-backend/pkg/cloudresourcemanager"

	"github.com/navikt/nada-backend/pkg/config/v2"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core/api/static"
	"github.com/navikt/nada-backend/pkg/syncers/metabase_collections"
	"github.com/navikt/nada-backend/pkg/syncers/metabase_mapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/navikt/nada-backend/pkg/bq"
	"github.com/navikt/nada-backend/pkg/database"
	dmpSA "github.com/navikt/nada-backend/pkg/sa"
	"github.com/navikt/nada-backend/pkg/service/core"
	"github.com/navikt/nada-backend/pkg/service/core/api/gcp"
	"github.com/navikt/nada-backend/pkg/service/core/api/http"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/routes"
	"github.com/navikt/nada-backend/pkg/service/core/storage"
	"github.com/rs/zerolog"
)

// nolint: tparallel,maintidx
func TestMetabase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(20*time.Minute))
	defer cancel()

	log := zerolog.New(zerolog.NewConsoleWriter())
	log.Level(zerolog.DebugLevel)

	c := NewContainers(t, log)
	defer c.Cleanup()

	pgCfg := c.RunPostgres(NewPostgresConfig())

	repo, err := database.New(
		pgCfg.ConnectionURL(),
		10,
		10,
	)
	assert.NoError(t, err)

	mbCfg := c.RunMetabase(NewMetabaseConfig())

	bqClient := bq.NewClient("", true, log)
	saClient := dmpSA.NewClient("", false)
	crmClient := crm.NewClient("", false, nil)

	stores := storage.NewStores(nil, repo, config.Config{}, log)

	zlog := zerolog.New(os.Stdout)
	r := TestRouter(zlog)

	crmapi := gcp.NewCloudResourceManagerAPI(crmClient)
	saapi := gcp.NewServiceAccountAPI(saClient)
	bqapi := gcp.NewBigQueryAPI(MetabaseProject, Location, PseudoDataSet, bqClient)

	mbapi := http.NewMetabaseHTTP(
		mbCfg.ConnectionURL()+"/api",
		mbCfg.Email,
		mbCfg.Password,
		"",
		false,
		false,
		log,
	)

	credBytes, err := os.ReadFile("../../tests-metabase-all-users-sa-creds.json")
	assert.NoError(t, err)

	mbService := core.NewMetabaseService(
		MetabaseProject,
		string(credBytes),
		MetabaseAllUsersServiceAccount,
		"group:"+GroupEmailAllUsers,
		mbapi,
		bqapi,
		saapi,
		crmapi,
		stores.ThirdPartyMappingStorage,
		stores.MetaBaseStorage,
		stores.BigQueryStorage,
		stores.DataProductsStorage,
		stores.AccessStorage,
		zlog,
	)

	queue := make(chan metabase_mapper.Work, 10)
	mapper := metabase_mapper.New(mbService, stores.ThirdPartyMappingStorage, 120, queue, log)

	err = stores.NaisConsoleStorage.UpdateAllTeamProjects(ctx, []*service.NaisTeamMapping{
		{
			Slug:       NaisTeamNada,
			GroupEmail: GroupEmailNada,
			ProjectID:  MetabaseProject,
		},
	})
	assert.NoError(t, err)

	dataproductService := core.NewDataProductsService(
		stores.DataProductsStorage,
		stores.BigQueryStorage,
		bqapi,
		stores.NaisConsoleStorage,
		GroupEmailAllUsers,
	)

	StorageCreateProductAreasAndTeams(t, stores.ProductAreaStorage)
	fuel, err := dataproductService.CreateDataproduct(ctx, UserOne, NewDataProductBiofuelProduction(GroupEmailNada, TeamSeagrassID))
	assert.NoError(t, err)

	openDataset, err := dataproductService.CreateDataset(ctx, UserOne, service.NewDataset{
		DataproductID: fuel.ID,
		Name:          "Open dataset",
		BigQuery: service.NewBigQuery{
			ProjectID: MetabaseProject,
			Dataset:   MetabaseDataset,
			Table:     MetabaseTableA,
		},
		Pii: service.PiiLevelNone,
	})
	assert.NoError(t, err)

	restrictedData, err := dataproductService.CreateDataset(ctx, UserOne, service.NewDataset{
		DataproductID: fuel.ID,
		Name:          "Restricted dataset",
		BigQuery: service.NewBigQuery{
			ProjectID: MetabaseProject,
			Dataset:   MetabaseDataset,
			Table:     MetabaseTableB,
		},
		Pii: service.PiiLevelNone,
	})
	assert.NoError(t, err)

	{
		h := handlers.NewMetabaseHandler(mbService, queue)
		e := routes.NewMetabaseEndpoints(zlog, h)
		f := routes.NewMetabaseRoutes(e, injectUser(UserOne))

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
		f := routes.NewAccessRoutes(e, injectUser(UserOne))

		f(r)
	}

	server := httptest.NewServer(r)
	defer server.Close()

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		cleanupCtx, cancel := context.WithDeadline(cleanupCtx, time.Now().Add(2*time.Minute))
		defer cancel()

		err := cleanupTestProject(cleanupCtx)
		if err != nil {
			t.Error(err)
		}
	})

	err = cleanupTestProject(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Adding an open dataset to metabase", func(t *testing.T) {
		NewTester(t, server).
			Post(ctx, service.GrantAccessData{
				DatasetID:   openDataset.ID,
				Expires:     nil,
				Subject:     strToStrPtr(GroupEmailAllUsers),
				SubjectType: strToStrPtr("group"),
			}, "/api/accesses/grant").
			HasStatusCode(http2.StatusNoContent)

		NewTester(t, server).
			Post(ctx, service.DatasetMap{Services: []string{service.MappingServiceMetabase}}, fmt.Sprintf("/api/datasets/%s/map", openDataset.ID)).
			HasStatusCode(http2.StatusAccepted)

		time.Sleep(200 * time.Millisecond)
		mapper.ProcessOne(ctx)

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

		tablePolicy, err := bqClient.GetTablePolicy(ctx, openDataset.Datasource.ProjectID, openDataset.Datasource.Dataset, openDataset.Datasource.Table)
		assert.NoError(t, err)
		assert.True(t, ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, MetabaseDataset)
		assert.NoError(t, err)
		assert.True(t, ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, MetabaseAllUsersServiceAccount))
	})

	t.Run("Soft delete open metabase database", func(t *testing.T) {
		datasetAccessEntries, err := stores.AccessStorage.ListActiveAccessToDataset(ctx, openDataset.ID)
		require.NoError(t, err)
		require.Len(t, datasetAccessEntries, 1)

		NewTester(t, server).
			Post(ctx, nil, fmt.Sprintf("/api/accesses/revoke?accessId=%s", datasetAccessEntries[0].ID)).
			HasStatusCode(http2.StatusNoContent)

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
		assert.False(t, ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))
	})

	t.Run("Restore soft deleted open metabase database", func(t *testing.T) {
		NewTester(t, server).
			Post(ctx, service.GrantAccessData{
				DatasetID:   openDataset.ID,
				Expires:     nil,
				Subject:     strToStrPtr(GroupEmailAllUsers),
				SubjectType: strToStrPtr("group"),
			}, "/api/accesses/grant").
			HasStatusCode(http2.StatusNoContent)

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

		tablePolicy, err := bqClient.GetTablePolicy(ctx, openDataset.Datasource.ProjectID, openDataset.Datasource.Dataset, openDataset.Datasource.Table)
		assert.NoError(t, err)
		assert.True(t, ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))
	})

	t.Run("Permanent delete of open metabase database", func(t *testing.T) {
		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, openDataset.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		NewTester(t, server).
			Post(ctx, service.DatasetMap{Services: []string{}}, fmt.Sprintf("/api/datasets/%s/map", openDataset.ID)).
			HasStatusCode(http2.StatusAccepted)

		time.Sleep(500 * time.Millisecond)
		mapper.ProcessOne(ctx)

		_, err = stores.MetaBaseStorage.GetMetadata(ctx, openDataset.ID, false)
		require.Error(t, err)

		_, err = mbapi.Database(ctx, *meta.DatabaseID)
		require.Error(t, err)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, openDataset.Datasource.ProjectID, openDataset.Datasource.Dataset, openDataset.Datasource.Table)
		assert.NoError(t, err)
		assert.False(t, ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, MetabaseDataset)
		assert.NoError(t, err)
		// Dataset Metadata Viewer is intentionally not removed when access for table is revoked so should be true
		assert.True(t, ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, meta.SAEmail))
	})

	t.Run("Adding a restricted dataset to metabase", func(t *testing.T) {
		NewTester(t, server).
			Post(ctx, service.DatasetMap{Services: []string{service.MappingServiceMetabase}}, fmt.Sprintf("/api/datasets/%s/map", restrictedData.ID)).
			HasStatusCode(http2.StatusAccepted)

		time.Sleep(500 * time.Millisecond)
		mapper.ProcessOne(ctx)

		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, restrictedData.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		collections, err := mbapi.GetCollections(ctx)
		require.NoError(t, err)
		assert.True(t, ContainsCollectionWithName(collections, "Restricted dataset 🔐"))

		permissionGroups, err := mbapi.GetPermissionGroups(ctx)
		require.NoError(t, err)
		assert.True(t, ContainsPermissionGroupWithNamePrefix(permissionGroups, "restricted-dataset"))

		sa, err := saClient.GetServiceAccount(ctx, fmt.Sprintf("projects/%s/serviceAccounts/%s", MetabaseProject, meta.SAEmail))
		require.NoError(t, err)
		assert.True(t, sa.Email == meta.SAEmail)

		keys, err := saClient.ListServiceAccountKeys(ctx, fmt.Sprintf("projects/%s/serviceAccounts/%s", MetabaseProject, meta.SAEmail))
		require.NoError(t, err)
		assert.Len(t, keys, 2) // will return 1 system managed key (always) in addition to the user managed key we created

		bindings, err := crmClient.ListProjectIAMPolicyBindings(ctx, MetabaseProject, "serviceAccount:"+meta.SAEmail)
		require.NoError(t, err)
		assert.True(t, ContainsProjectIAMPolicyBindingForSubject(bindings, NadaMetabaseRole, "serviceAccount:"+meta.SAEmail))

		require.NotNil(t, meta.PermissionGroupID)
		permissionGraphForGroup, err := mbapi.GetPermissionGraphForGroup(ctx, *meta.PermissionGroupID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, permissionGraphForGroup.Groups, strconv.Itoa(*meta.PermissionGroupID))
		assert.Equal(t, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedData.ID), meta.SAEmail)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, restrictedData.Datasource.ProjectID, restrictedData.Datasource.Dataset, restrictedData.Datasource.Table)
		assert.NoError(t, err)
		assert.True(t, ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+mbService.ConstantServiceAccountEmailFromDatasetID(restrictedData.ID)))
		assert.False(t, ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, MetabaseDataset)
		assert.NoError(t, err)
		assert.True(t, ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedData.ID)))
	})

	t.Run("Delete restricted metabase database", func(t *testing.T) {
		NewTester(t, server).
			Post(ctx, service.DatasetMap{Services: []string{}}, fmt.Sprintf("/api/datasets/%s/map", restrictedData.ID)).
			HasStatusCode(http2.StatusAccepted)

		time.Sleep(500 * time.Millisecond)
		mapper.ProcessOne(ctx)

		_, err = stores.MetaBaseStorage.GetMetadata(ctx, restrictedData.ID, true)
		require.Error(t, err)

		collections, err := mbapi.GetCollections(ctx)
		require.NoError(t, err)
		assert.False(t, ContainsCollectionWithName(collections, "Restricted dataset 🔐"))

		permissionGroups, err := mbapi.GetPermissionGroups(ctx)
		require.NoError(t, err)
		assert.False(t, ContainsPermissionGroupWithNamePrefix(permissionGroups, "restricted-dataset"))

		_, err = saClient.GetServiceAccount(ctx, fmt.Sprintf("projects/%s/serviceAccounts/%s", MetabaseProject, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedData.ID)))
		require.Error(t, err)

		bindings, err := crmClient.ListProjectIAMPolicyBindings(ctx, MetabaseProject, "serviceAccount:"+mbService.ConstantServiceAccountEmailFromDatasetID(restrictedData.ID))
		require.NoError(t, err)
		assert.False(t, ContainsProjectIAMPolicyBindingForSubject(bindings, NadaMetabaseRole, "serviceAccount:"+mbService.ConstantServiceAccountEmailFromDatasetID(restrictedData.ID)))

		tablePolicy, err := bqClient.GetTablePolicy(ctx, restrictedData.Datasource.ProjectID, restrictedData.Datasource.Dataset, restrictedData.Datasource.Table)
		assert.NoError(t, err)
		assert.False(t, ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+mbService.ConstantServiceAccountEmailFromDatasetID(restrictedData.ID)))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, MetabaseDataset)
		assert.NoError(t, err)
		assert.False(t, ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedData.ID)))
	})

	t.Run("Add new restricted metabase database", func(t *testing.T) {
		NewTester(t, server).
			Post(ctx, service.DatasetMap{Services: []string{service.MappingServiceMetabase}}, fmt.Sprintf("/api/datasets/%s/map", restrictedData.ID)).
			HasStatusCode(http2.StatusAccepted)

		time.Sleep(500 * time.Millisecond)
		mapper.ProcessOne(ctx)

		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, restrictedData.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		collections, err := mbapi.GetCollections(ctx)
		require.NoError(t, err)
		assert.True(t, ContainsCollectionWithName(collections, "Restricted dataset 🔐"))

		permissionGroups, err := mbapi.GetPermissionGroups(ctx)
		require.NoError(t, err)
		assert.True(t, ContainsPermissionGroupWithNamePrefix(permissionGroups, "restricted-dataset"))

		sa, err := saClient.GetServiceAccount(ctx, fmt.Sprintf("projects/%s/serviceAccounts/%s", MetabaseProject, meta.SAEmail))
		require.NoError(t, err)
		assert.True(t, sa.Email == meta.SAEmail)

		keys, err := saClient.ListServiceAccountKeys(ctx, fmt.Sprintf("projects/%s/serviceAccounts/%s", MetabaseProject, meta.SAEmail))
		require.NoError(t, err)
		assert.Len(t, keys, 2) // will return 1 system managed key (always) in addition to the user managed key we created

		bindings, err := crmClient.ListProjectIAMPolicyBindings(ctx, MetabaseProject, "serviceAccount:"+meta.SAEmail)
		require.NoError(t, err)
		assert.True(t, ContainsProjectIAMPolicyBindingForSubject(bindings, NadaMetabaseRole, "serviceAccount:"+meta.SAEmail))

		require.NotNil(t, meta.PermissionGroupID)
		permissionGraphForGroup, err := mbapi.GetPermissionGraphForGroup(ctx, *meta.PermissionGroupID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, permissionGraphForGroup.Groups, strconv.Itoa(*meta.PermissionGroupID))
		assert.Equal(t, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedData.ID), meta.SAEmail)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, restrictedData.Datasource.ProjectID, restrictedData.Datasource.Dataset, restrictedData.Datasource.Table)
		assert.NoError(t, err)
		assert.True(t, ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+mbService.ConstantServiceAccountEmailFromDatasetID(restrictedData.ID)))
		assert.False(t, ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, MetabaseDataset)
		assert.NoError(t, err)
		assert.True(t, ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedData.ID)))
	})

	t.Run("Removing 🔐 is added back", func(t *testing.T) {
		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, restrictedData.ID, false)
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
		assert.True(t, ContainsCollectionWithName(collections, "My new collection name 🔐"))
	})

	t.Run("Opening a previously restricted metabase dataset", func(t *testing.T) {
		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, restrictedData.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		permissionGraphForGroup, err := mbapi.GetPermissionGraphForGroup(ctx, service.MetabaseAllUsersGroupID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Contains(t, permissionGraphForGroup.Groups, strconv.Itoa(service.MetabaseAllUsersGroupID))

		NewTester(t, server).
			Post(ctx, service.GrantAccessData{
				DatasetID:   restrictedData.ID,
				Expires:     nil,
				Subject:     strToStrPtr(GroupEmailAllUsers),
				SubjectType: strToStrPtr("group"),
			}, "/api/accesses/grant").
			HasStatusCode(http2.StatusNoContent)

		time.Sleep(1000 * time.Millisecond)

		meta, err = stores.MetaBaseStorage.GetMetadata(ctx, restrictedData.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		permissionGroups, err := mbapi.GetPermissionGroups(ctx)
		require.NoError(t, err)
		assert.False(t, ContainsPermissionGroupWithNamePrefix(permissionGroups, "restricted-dataset"))
		assert.Equal(t, MetabaseAllUsersServiceAccount, meta.SAEmail)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, restrictedData.Datasource.ProjectID, restrictedData.Datasource.Dataset, restrictedData.Datasource.Table)
		assert.NoError(t, err)
		assert.True(t, ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))
		assert.False(t, ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+mbService.ConstantServiceAccountEmailFromDatasetID(restrictedData.ID)))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, MetabaseDataset)
		assert.NoError(t, err)
		assert.False(t, ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, mbService.ConstantServiceAccountEmailFromDatasetID(restrictedData.ID)))
	})

	t.Run("Delete previously restricted metabase database", func(t *testing.T) {
		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, restrictedData.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		NewTester(t, server).
			Post(ctx, service.DatasetMap{Services: []string{}}, fmt.Sprintf("/api/datasets/%s/map", restrictedData.ID)).
			HasStatusCode(http2.StatusAccepted)

		time.Sleep(500 * time.Millisecond)
		mapper.ProcessOne(ctx)

		_, err = stores.MetaBaseStorage.GetMetadata(ctx, restrictedData.ID, false)
		require.Error(t, err)

		_, err = mbapi.Database(ctx, *meta.DatabaseID)
		require.Error(t, err)

		tablePolicy, err := bqClient.GetTablePolicy(ctx, restrictedData.Datasource.ProjectID, restrictedData.Datasource.Dataset, restrictedData.Datasource.Table)
		assert.NoError(t, err)
		assert.False(t, ContainsTablePolicyBindingForSubject(tablePolicy, BigQueryDataViewerRole, "serviceAccount:"+MetabaseAllUsersServiceAccount))

		bqDataset, err := bqClient.GetDataset(ctx, MetabaseProject, MetabaseDataset)
		assert.NoError(t, err)
		// Dataset Metadata Viewer is intentionally not removed when access for table is revoked so should be true
		assert.True(t, ContainsDatasetAccessForSubject(bqDataset.Access, BigQueryMetadataViewerRole, meta.SAEmail))
	})
}

// Ensure that the test project has no lingering state from previous test runs
// Perhaps a better idea to create a brand new dataset and tables for each test?
func cleanupTestProject(ctx context.Context) error {
	saClient := dmpSA.NewClient("", false)
	bqClient := bq.NewClient("", true, zerolog.New(os.Stdout))
	crmClient := crm.NewClient("", false, nil)

	// Cleaning up BigQuery dataset Metadata Viewer grants
	err := bqClient.UpdateDatasetAccess(ctx, MetabaseProject, MetabaseDataset, bq.RemoveDatasetAccessMembersWithRole(BigQueryMetadataViewerRole, zerolog.Nop()))
	if err != nil {
		return fmt.Errorf("error updating dataset access: %v", err)
	}

	// Cleaning up BigQuery Data Viewer grants for test tables
	tables, err := bqClient.GetTables(ctx, MetabaseProject, MetabaseDataset)
	if err != nil {
		return fmt.Errorf("error getting tables: %v", err)
	}
	for _, t := range tables {
		policy, err := bqClient.GetTablePolicy(ctx, MetabaseProject, MetabaseDataset, t.TableID)
		if err != nil {
			return fmt.Errorf("error getting table policy: %v", err)
		}

		for _, e := range policy.Members(iam.RoleName(BigQueryDataViewerRole)) {
			fmt.Printf("Removing %s on table %s for user %s\n", BigQueryDataViewerRole, t.TableID, e)
			err := bqClient.RemoveAndSetTablePolicy(ctx, MetabaseProject, MetabaseDataset, t.TableID, BigQueryDataViewerRole, e)
			if err != nil {
				return fmt.Errorf("error removing table policy: %v", err)
			}
		}
	}

	// Deleting restricted database service accounts
	serviceAccounts, err := saClient.ListServiceAccounts(ctx, MetabaseProject)
	if err != nil {
		return fmt.Errorf("error getting service accounts: %v", err)
	}
	for _, sa := range serviceAccounts {
		if !strings.HasPrefix(sa.Email, "nada-") {
			continue
		}

		fmt.Printf("Deleting restricted database service account %s\n", sa.Email)
		err := saClient.DeleteServiceAccount(ctx, fmt.Sprintf("projects/%s/serviceAccounts/%s", MetabaseProject, sa.Email))
		if err != nil {
			if errors.Is(err, dmpSA.ErrNotFound) {
				continue
			}
			return fmt.Errorf("error deleting service account: %v", err)
		}
	}

	// Cleaning up NADA metabase project iam role grants for deleted service accounts
	err = crmClient.UpdateProjectIAMPolicyBindingsMembers(ctx, MetabaseProject, crm.RemoveDeletedMembersWithRole([]string{NadaMetabaseRole}, zerolog.Nop()))
	if err != nil {
		return fmt.Errorf("error updating project iam policy bindings: %v", err)
	}

	return nil
}
