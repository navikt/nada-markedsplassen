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

	crm "github.com/navikt/nada-backend/pkg/cloudresourcemanager"

	"github.com/navikt/nada-backend/pkg/config/v2"
	"github.com/navikt/nada-backend/pkg/sa"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core/api/static"
	"github.com/navikt/nada-backend/pkg/syncers/metabase_collections"
	"github.com/navikt/nada-backend/pkg/syncers/metabase_mapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/navikt/nada-backend/pkg/bq"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service/core"
	"github.com/navikt/nada-backend/pkg/service/core/api/gcp"
	"github.com/navikt/nada-backend/pkg/service/core/api/http"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/routes"
	"github.com/navikt/nada-backend/pkg/service/core/storage"
	"github.com/rs/zerolog"

	dmpSA "github.com/navikt/nada-backend/pkg/sa"
)

// I don't know man, is this really better?

type metabaseAPIMock struct {
	api service.MetabaseAPI
}

func (m *metabaseAPIMock) DeleteUser(ctx context.Context, id int) error {
	return m.api.DeleteUser(ctx, id)
}

func (m *metabaseAPIMock) AddPermissionGroupMember(ctx context.Context, groupID int, userID int) error {
	return m.api.AddPermissionGroupMember(ctx, groupID, userID)
}

func (m *metabaseAPIMock) ArchiveCollection(ctx context.Context, colID int) error {
	return m.api.ArchiveCollection(ctx, colID)
}

func (m *metabaseAPIMock) AutoMapSemanticTypes(ctx context.Context, dbID int) error {
	return nil
}

func (m *metabaseAPIMock) CreateCollection(ctx context.Context, name string) (int, error) {
	return m.api.CreateCollection(ctx, name)
}

func (m *metabaseAPIMock) CreateCollectionWithAccess(ctx context.Context, groupID int, name string, removeAllUsersAccess bool) (int, error) {
	return m.api.CreateCollectionWithAccess(ctx, groupID, name, removeAllUsersAccess)
}

func (m *metabaseAPIMock) CreateDatabase(ctx context.Context, team, name, saJSON, saEmail string, ds *service.BigQuery) (int, error) {
	return 10, nil
}

func (m *metabaseAPIMock) UpdateDatabase(ctx context.Context, dbID int, saJSON, saEmail string) error {
	return nil
}

func (m *metabaseAPIMock) GetPermissionGroups(ctx context.Context) ([]service.MetabasePermissionGroup, error) {
	return m.api.GetPermissionGroups(ctx)
}

func (m *metabaseAPIMock) GetOrCreatePermissionGroup(ctx context.Context, name string) (int, error) {
	return m.api.GetOrCreatePermissionGroup(ctx, name)
}

func (m *metabaseAPIMock) CreatePermissionGroup(ctx context.Context, name string) (int, error) {
	return m.api.CreatePermissionGroup(ctx, name)
}

func (m *metabaseAPIMock) Databases(ctx context.Context) ([]service.MetabaseDatabase, error) {
	return []service.MetabaseDatabase{}, nil
}

func (m *metabaseAPIMock) Database(ctx context.Context, dbID int) (*service.MetabaseDatabase, error) {
	return &service.MetabaseDatabase{}, nil
}

func (m *metabaseAPIMock) DeleteDatabase(ctx context.Context, id int) error {
	return nil
}

func (m *metabaseAPIMock) DeletePermissionGroup(ctx context.Context, groupID int) error {
	return m.api.DeletePermissionGroup(ctx, groupID)
}

func (m *metabaseAPIMock) GetPermissionGroup(ctx context.Context, groupID int) ([]service.MetabasePermissionGroupMember, error) {
	return m.api.GetPermissionGroup(ctx, groupID)
}

func (m *metabaseAPIMock) HideTables(ctx context.Context, ids []int) error {
	return nil
}

func (m *metabaseAPIMock) OpenAccessToDatabase(ctx context.Context, databaseID int) error {
	return nil
}

func (m *metabaseAPIMock) RemovePermissionGroupMember(ctx context.Context, memberID int) error {
	return m.api.RemovePermissionGroupMember(ctx, memberID)
}

func (m *metabaseAPIMock) GetPermissionGraphForGroup(ctx context.Context, groupID int) (*service.PermissionGraphGroups, error) {
	return m.api.GetPermissionGraphForGroup(ctx, groupID)
}

func (m *metabaseAPIMock) RestrictAccessToDatabase(ctx context.Context, groupID int, databaseID int) error {
	return nil
}

func (m *metabaseAPIMock) SetCollectionAccess(ctx context.Context, groupID int, collectionID int, removeAllUsersAccess bool) error {
	return m.api.SetCollectionAccess(ctx, groupID, collectionID, removeAllUsersAccess)
}

func (m *metabaseAPIMock) ShowTables(ctx context.Context, ids []int) error {
	return nil
}

func (m *metabaseAPIMock) Tables(ctx context.Context, dbID int, includeHidden bool) ([]service.MetabaseTable, error) {
	tables := map[int][]service.MetabaseTable{
		10: {
			{
				Name: "consumption_rates",
				Fields: []service.MetabaseField{
					{},
				},
			},
		},
	}

	return tables[dbID], nil
}

func (m *metabaseAPIMock) GetCollections(ctx context.Context) ([]*service.MetabaseCollection, error) {
	return m.api.GetCollections(ctx)
}

func (m *metabaseAPIMock) UpdateCollection(ctx context.Context, collection *service.MetabaseCollection) error {
	return m.api.UpdateCollection(ctx, collection)
}

func (m *metabaseAPIMock) FindUserByEmail(ctx context.Context, email string) (*service.MetabaseUser, error) {
	return m.api.FindUserByEmail(ctx, email)
}

func (m *metabaseAPIMock) GetUsers(ctx context.Context) ([]service.MetabaseUser, error) {
	return m.api.GetUsers(ctx)
}

func (m *metabaseAPIMock) CreateUser(ctx context.Context, email string) (*service.MetabaseUser, error) {
	return m.api.CreateUser(ctx, email)
}

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
	saClient := sa.NewClient("", false)
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
		cleanupTestProject(ctx)
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
		assert.True(t, ContainsProjectIAMPolicyBindingForSubject(bindings, fmt.Sprintf("projects/%s/roles/nada.metabase", MetabaseProject), "serviceAccount:"+meta.SAEmail))

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
	})
}

func cleanupTestProject(ctx context.Context) error {
	saClient := sa.NewClient("", false)
	bqClient := bq.NewClient("", true, zerolog.New(os.Stdout))
	// crmClient := crm.NewClient("", false, nil)

	serviceAccounts, err := saClient.ListServiceAccounts(ctx, MetabaseProject)
	if err != nil {
		return fmt.Errorf("error getting service accounts: %v", err)
	}

	tables, err := bqClient.GetTables(ctx, MetabaseProject, MetabaseDataset)
	if err != nil {
		return fmt.Errorf("error getting tables: %v", err)
	}

	for _, sa := range serviceAccounts {
		for _, t := range tables {
			fmt.Printf("Removing policy for table %s for user %s\n", t.TableID, sa.Email)
			err := bqClient.RemoveAndSetTablePolicy(ctx, MetabaseProject, MetabaseDataset, t.TableID, BigQueryDataViewerRole, "serviceAccount:"+sa.Email)
			if err != nil {
				return fmt.Errorf("error removing table policy: %v", err)
			}
		}
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

	return nil
}
