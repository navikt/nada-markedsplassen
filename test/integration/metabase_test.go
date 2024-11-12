package integration

import (
	"context"
	"fmt"
	http2 "net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	crm "github.com/navikt/nada-backend/pkg/cloudresourcemanager"

	crmEmulator "github.com/navikt/nada-backend/pkg/cloudresourcemanager/emulator"
	"github.com/navikt/nada-backend/pkg/config/v2"
	"github.com/navikt/nada-backend/pkg/sa"
	serviceAccountEmulator "github.com/navikt/nada-backend/pkg/sa/emulator"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core/api/static"
	"github.com/navikt/nada-backend/pkg/syncers/metabase_collections"
	"github.com/navikt/nada-backend/pkg/syncers/metabase_mapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/cloudresourcemanager/v1"

	"github.com/navikt/nada-backend/pkg/bq"
	bigQueryEmulator "github.com/navikt/nada-backend/pkg/bq/emulator"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service/core"
	"github.com/navikt/nada-backend/pkg/service/core/api/gcp"
	"github.com/navikt/nada-backend/pkg/service/core/api/http"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/routes"
	"github.com/navikt/nada-backend/pkg/service/core/storage"
	"github.com/rs/zerolog"
)

// I don't know man, is this really better?

type metabaseAPIMock struct {
	api service.MetabaseAPI
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

	fuelBqSchema := NewDatasetBiofuelConsumptionRatesSchema()

	bqe := bigQueryEmulator.New(log)
	bqe.WithProject(Project, fuelBqSchema...)
	bqe.EnableMock(false, log, bigQueryEmulator.NewPolicyMock(log).Mocks()...)

	bqHTTPPort := strconv.Itoa(GetFreePort(t))
	bqHTTPAddr := fmt.Sprintf("127.0.0.1:%s", bqHTTPPort)
	if len(os.Getenv("CI")) > 0 {
		bqHTTPAddr = fmt.Sprintf("0.0.0.0:%s", bqHTTPPort)
	}
	bqGRPCAddr := fmt.Sprintf("127.0.0.1:%s", strconv.Itoa(GetFreePort(t)))

	go func() {
		_ = bqe.Serve(ctx, bqHTTPAddr, bqGRPCAddr)
	}()

	bqClient := bq.NewClient("http://"+bqHTTPAddr, false, log)

	saEmulator := serviceAccountEmulator.New(log)
	saURL := saEmulator.Run()
	saClient := sa.NewClient(saURL, true)

	crmEmulator := crmEmulator.New(log)
	crmEmulator.SetPolicy(Project, &cloudresourcemanager.Policy{
		Bindings: []*cloudresourcemanager.Binding{
			{
				Role:    "roles/owner",
				Members: []string{fmt.Sprintf("user:%s", GroupEmailNada)},
			},
		},
	})
	crmURL := crmEmulator.Run()
	crmClient := crm.NewClient(crmURL, true, nil)

	stores := storage.NewStores(nil, repo, config.Config{}, log)

	zlog := zerolog.New(os.Stdout)
	r := TestRouter(zlog)

	bigQueryContainerHostPort := "http://host.docker.internal:" + bqHTTPPort
	if len(os.Getenv("CI")) > 0 {
		bigQueryContainerHostPort = "http://172.17.0.1:" + bqHTTPPort
	}

	crmapi := gcp.NewCloudResourceManagerAPI(crmClient)
	saapi := gcp.NewServiceAccountAPI(saClient)
	bqapi := gcp.NewBigQueryAPI(Project, Location, PseudoDataSet, bqClient)
	// FIXME: should we just add /api to the connectionurl returned
	mbapi := &metabaseAPIMock{
		api: http.NewMetabaseHTTP(
			mbCfg.ConnectionURL()+"/api",
			mbCfg.Email,
			mbCfg.Password,
			// We want metabase to connect with the big query emulator
			// running on the host
			bigQueryContainerHostPort,
			true,
			false,
			log,
		),
	}

	mbService := core.NewMetabaseService(
		Project,
		fakeMetabaseSA,
		"nada-metabase@test.iam.gserviceaccount.com",
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
	mapper := metabase_mapper.New(mbService, stores.ThirdPartyMappingStorage, 60, queue, log)

	err = stores.NaisConsoleStorage.UpdateAllTeamProjects(ctx, []*service.NaisTeamMapping{
		{
			Slug:       NaisTeamNada,
			GroupEmail: GroupEmailNada,
			ProjectID:  Project,
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

	fuelData, err := dataproductService.CreateDataset(ctx, UserOne, NewDatasetBiofuelConsumptionRates(fuel.ID))
	assert.NoError(t, err)

	openDataset, err := dataproductService.CreateDataset(ctx, UserOne, NewDatasetBiofuelConsumptionRates(fuel.ID))
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
		h := handlers.NewAccessHandler(accessService, mbService, Project)
		e := routes.NewAccessEndpoints(zlog, h)
		f := routes.NewAccessRoutes(e, injectUser(UserOne))

		f(r)
	}

	server := httptest.NewServer(r)
	defer server.Close()

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
	})

	t.Run("Adding a restricted dataset to metabase", func(t *testing.T) {
		NewTester(t, server).
			Post(ctx, service.GrantAccessData{
				DatasetID:   fuelData.ID,
				Expires:     nil,
				Subject:     strToStrPtr(UserOne.Email),
				SubjectType: strToStrPtr(service.SubjectTypeUser),
			}, "/api/accesses/grant").
			HasStatusCode(http2.StatusNoContent)

		NewTester(t, server).
			Post(ctx, service.DatasetMap{Services: []string{service.MappingServiceMetabase}}, fmt.Sprintf("/api/datasets/%s/map", fuelData.ID)).
			HasStatusCode(http2.StatusAccepted)

		time.Sleep(200 * time.Millisecond)
		mapper.ProcessOne(ctx)

		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, fuelData.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		collections, err := mbapi.GetCollections(ctx)
		require.NoError(t, err)
		assert.True(t, ContainsCollectionWithName(collections, "Biofuel Consumption Rates 🔐"))

		permissionGroups, err := mbapi.GetPermissionGroups(ctx)
		require.NoError(t, err)
		assert.True(t, ContainsPermissionGroupWithNamePrefix(permissionGroups, "biofuel-consumption-rates"))

		serviceAccount := saEmulator.GetServiceAccounts()
		assert.Len(t, serviceAccount, 1)
		assert.True(t, ContainsServiceAccount(serviceAccount, "nada-", "@test-project.iam.gserviceaccount.com"))

		serviceAccountKeys := saEmulator.GetServiceAccountKeys()
		assert.Len(t, serviceAccountKeys, 1)

		projectPolicy := crmEmulator.GetPolicy(Project)
		assert.Len(t, projectPolicy.Bindings, 2)
		assert.Equal(t, projectPolicy.Bindings[1].Role, "projects/test-project/roles/nada.metabase")

		require.NotNil(t, meta.PermissionGroupID)
		permissionGraphForGroup, err := mbapi.GetPermissionGraphForGroup(ctx, *meta.PermissionGroupID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, permissionGraphForGroup.Groups, strconv.Itoa(*meta.PermissionGroupID))
	})

	t.Run("Removing 🔐 is added back", func(t *testing.T) {
		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, fuelData.ID, false)
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
		meta, err := stores.MetaBaseStorage.GetMetadata(ctx, fuelData.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		permissionGraphForGroup, err := mbapi.GetPermissionGraphForGroup(ctx, service.MetabaseAllUsersGroupID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, permissionGraphForGroup.Groups, strconv.Itoa(service.MetabaseAllUsersGroupID))

		NewTester(t, server).
			Post(ctx, service.GrantAccessData{
				DatasetID:   fuelData.ID,
				Expires:     nil,
				Subject:     strToStrPtr(GroupEmailAllUsers),
				SubjectType: strToStrPtr("group"),
			}, "/api/accesses/grant").
			HasStatusCode(http2.StatusNoContent)

		time.Sleep(time.Second)

		meta, err = stores.MetaBaseStorage.GetMetadata(ctx, fuelData.ID, false)
		require.NoError(t, err)
		require.NotNil(t, meta.SyncCompleted)

		permissionGroups, err := mbapi.GetPermissionGroups(ctx)
		require.NoError(t, err)
		assert.False(t, ContainsPermissionGroupWithNamePrefix(permissionGroups, "biofuel-consumption-rates"))
	})
}
