package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/goccy/bigquery-emulator/types"
	bigQueryEmulator "github.com/navikt/nada-backend/pkg/bq/emulator"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/service"
)

// nolint: gochecknoglobals
var (
	ProductAreaOceanicID   = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	ProductAreaOceanicName = "Oceanic"

	ProductAreaCostalID   = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	ProductAreaCostalName = "Costal"

	TeamSeagrassID   = uuid.MustParse("00000000-0000-0000-0000-000000000003")
	TeamSeagrassName = "Seagrass"

	TeamReefID   = uuid.MustParse("00000000-0000-0000-0000-000000000004")
	TeamReefName = "Reef"

	GroupNameReef  = "reef"
	GroupEmailReef = "reef@nav.no"

	UserOneName        = "User Userson"
	UserOneEmail       = "user.userson@email.com"
	UserOneIdent       = "v101010"
	UserTwoName        = "Another Userson"
	UserTwoEmail       = "another.userson@email.com"
	UserTwoIdent       = "v202020"
	GroupNameNada      = "nada"
	GroupEmailNada     = "nada@nav.no"
	NaisTeamNada       = "nada"
	GroupNameAllUsers  = "all-users"
	GroupEmailAllUsers = "all-users@nav.no"

	Project       = "test-project"
	Location      = "europe-north1"
	PseudoDataSet = "pseudo-test-dataset"

	MetabaseProject                = "metabase-integration-tests"
	MetabaseDatasetPrefix          = "integration_tests"
	MetabaseTableA                 = "table_a"
	MetabaseTableB                 = "table_b"
	MetabaseAllUsersServiceAccount = "all-metabase-users@metabase-integration-tests.iam.gserviceaccount.com"
	BigQueryDataViewerRole         = "roles/bigquery.dataViewer"
	BigQueryMetadataViewerRole     = "roles/bigquery.metadataViewer"
	NadaMetabaseRole               = fmt.Sprintf("projects/%s/roles/nada.metabase", MetabaseProject)

	UserOne = &service.User{
		Name:  UserOneName,
		Email: UserOneEmail,
		Ident: UserOneIdent,
		GoogleGroups: []service.Group{
			{
				Name:  GroupNameNada,
				Email: GroupEmailNada,
			},
			{
				Name:  GroupNameAllUsers,
				Email: GroupEmailAllUsers,
			},
		},
	}

	UserTwo = &service.User{
		Name:  UserTwoName,
		Email: UserTwoEmail,
		Ident: UserTwoIdent,
		GoogleGroups: []service.Group{
			{
				Name:  GroupNameAllUsers,
				Email: GroupEmailAllUsers,
			},
		},
	}
)

// nolint: gosec
const fakeMetabaseSA = `{
	"type": "service_account",
	"project_id": "test",
	"private_key_id": "very-long-key-id",
	"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCQ+2XPvBgrF0rC\nSNsZxYgyrTxtr47z32yKO38Hi4ViRbDZVMr2isz2gnEW3MgwzCWEDuZqFkpigAUS\n9JPTHFXbNgHv28mAcoNEUBlabmzGJG+MUL3ZbiwjOVRhAyZXexFb9XTNk/0a30tF\nPw2HHtJYKda/C0ELitOBjfz74yhVXN8U16gEU23eRAmUb82yb2v2gy9pe8CjB5AZ\n8iOEtCACte5uwpoSSpb9wUZNcI4uwlX9jUDrLP7eFeAhaYz7xSMuHHGYM9P4rdcN\n77OGB8oA3k1A+sI6vnZaU/++RvbhHkK7xVQdlrEWHd9Mkw1SBTN4TIkxAcdn1fPv\ngujC4pbNAgMBAAECggEAEeWWqdAUQ616YwVPVp0VtPvBi+CyCtS3t1Ck26+TZZAo\nxos7CrgTmRJ9F09lvvBUaTrVop7hy8r11WMvpE3JI2hFUPCWrS51mccxegJLlyIE\nSxPke6Sn+ikni1oyL3ZXrDxekMoF4n1R81hXOSjK2ms/wRGIk/4tIb7/TbC6196R\ncWtT9unK1tWpaIzBaC0MH0/hnKKC9Vq9g+ezdnFo8eesdOiBoO8R5aAOofcA00Fm\n47RW0BoQNwIeyv6YcD2jg3cg8/FzODEktOW3/WP/Svr6FFPeQKKp3wjRN/ovU4kS\n/tDsHrYa5cSrpuI1dkzqHbu3AbQBV/NNd9pzJwAfwQKBgQDDn/8D1BhllPPTqOSD\n/2Z51MpZ+1OLkp11zFLUCnxsIop1wQHI58CTdn6+6Z4z3hEpJVkcxumkH0LMFHnH\n0ZZt8PaaqMQuIhxJZC4t8a/PuAGDK2G7kvusoSq92AVZjwlrIJ9Ij/RFE2Qoyo1o\n+9tLGUI9TMBAqcsdFReUxIAZOQKBgQC9ui999G5d5S0Dyg8+qGNf/QPM/yVfSA3d\nViCw77c/UiHG4mDa/leMBMLunKp00x8aE4dpbZYP138Jwzm1YS1i/ifqNEfdu5it\nLeNKHe6lEWnjW7j31Hn9oQbSG5BNtpHyII/1qs2YwmANWmh8o42SuNnrxoq3qlbd\n13KC2B5ONQKBgGvChdKREgNbAtlUTtTbapKwAeuBQ2s+D1jlfbbqM9HJUSY+dII8\nD1vryTPXMtt1d1SIC0eL1wYeZkhO+yp0LH5RXzagwrh698QB2GJcoTE2NjcQPZz7\nAYH9obLD/WZxIYoOhU+OZMtsPB8wPKdZHVqIBnIIBltYbNePV9cOS1YZAoGBALV3\nDhOfpaxDFbIJIlmgvyPBIVCCPWGLzk8EINJ7BT8YNFxAi7kKCfxPVY7Z46NHhvju\n8tZgzWWrjMNuqZSVNM75Hn5AsPgghOAnAr0SMf5J0Ih4Y0sPO/rdeGOfn37k/2Sh\nxm+HhYv1Zd9/uG52FGPgT/bV+DnBP8KBXfJN+XZ9AoGBAIPl1Ue+ZF3aUGPEwaZk\nMPeLEvGpVfxgBs8iJszJ+PXerNlZW1Bf+Xz3M0nlQF5vV7IdvcpMdGLBEhqxbijM\n7i0RnotQpqGObSoezJivZTmy8Yn6apH5HBAwlmrSYsqYJeWfWNdMhaYiCSOsq1eT\nlz7CEBohLrUvBJZe2b0MRFIJ\n-----END PRIVATE KEY-----\n",
	"client_email": "nada-metabase@test.iam.gserviceaccount.com",
	"client_id": "very-long-client-id",
	"auth_uri": "https://accounts.google.com/o/oauth2/auth",
	"token_uri": "https://oauth2.googleapis.com/token",
	"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/nada-metabase%40test.iam.gserviceaccount.com"
  }
  `

func StorageCreateNaisConsoleTeamsAndProjects(t *testing.T, storage service.NaisConsoleStorage, teams []*service.NaisTeamMapping) {
	t.Helper()

	err := storage.UpdateAllTeamProjects(context.Background(), teams)
	if err != nil {
		t.Fatalf("creating teams and projects: %v", err)
	}
}

func StorageCreateProductAreasAndTeams(t *testing.T, storage service.ProductAreaStorage) {
	t.Helper()

	pas := []*service.UpsertProductAreaRequest{
		{
			ID:   ProductAreaOceanicID,
			Name: ProductAreaOceanicName,
		},
		{
			ID:   ProductAreaCostalID,
			Name: ProductAreaCostalName,
		},
	}

	teams := []*service.UpsertTeamRequest{
		{
			ID:            TeamSeagrassID,
			ProductAreaID: ProductAreaOceanicID,
			Name:          TeamSeagrassName,
		},
		{
			ID:            TeamReefID,
			ProductAreaID: ProductAreaCostalID,
			Name:          TeamReefName,
		},
	}

	err := storage.UpsertProductAreaAndTeam(context.Background(), pas, teams)
	if err != nil {
		t.Fatalf("creating product areas and teams: %v", err)
	}
}

func NewDataProductBiofuelProduction(group string, teamID uuid.UUID) service.NewDataproduct {
	return service.NewDataproduct{
		Name:          "Biofuel Production",
		Description:   strToStrPtr("Using seagrass as a feedstock to create renewable biofuels"),
		Group:         group,
		ProductAreaID: &ProductAreaOceanicID,
		TeamID:        &teamID,
	}
}

func NewDatasetBiofuelConsumptionRatesSchema() []*bigQueryEmulator.Dataset {
	return []*bigQueryEmulator.Dataset{
		{
			DatasetID: "biofuel",
			TableID:   "consumption_rates",
			Columns: []*types.Column{
				bigQueryEmulator.ColumnRequired("id"),
				bigQueryEmulator.ColumnNullable("fuel_type"),
				bigQueryEmulator.ColumnNullable("consumption_rate"),
				bigQueryEmulator.ColumnNullable("unit"),
			},
		},
		{
			DatasetID: PseudoDataSet,
		},
	}
}

func NewDatasetBiofuelConsumptionRates(dataProductID uuid.UUID) service.NewDataset {
	dataset := NewDatasetBiofuelConsumptionRatesSchema()[0]

	return service.NewDataset{
		DataproductID: dataProductID,
		Name:          "Biofuel Consumption Rates",
		Description:   strToStrPtr("Consumption rates of biofuels in the transportation sector"),
		Keywords:      []string{"biofuel", "consumption", "rates"},
		BigQuery: service.NewBigQuery{
			ProjectID: Project,
			Dataset:   dataset.DatasetID,
			Table:     dataset.TableID,
		},
		Pii: service.PiiLevelNone,
	}
}

func NewDataProductAquacultureFeed(group string, teamID uuid.UUID) service.NewDataproduct {
	return service.NewDataproduct{
		Name:          "Aquaculture Feed",
		Description:   strToStrPtr("Producing high-nutrient feed for aquaculture industries from processed seagrass"),
		Group:         group,
		ProductAreaID: &ProductAreaOceanicID,
		TeamID:        &teamID,
	}
}

func NewDataProductReefMonitoring(group string, teamID uuid.UUID) service.NewDataproduct {
	return service.NewDataproduct{
		Name:          "Reef Monitoring Equipment",
		Description:   strToStrPtr("Advanced sensors and monitoring devices for continuous assessment"),
		Group:         group,
		ProductAreaID: &ProductAreaCostalID,
		TeamID:        &teamID,
	}
}

func NewDataProductProtectiveBarriers(group string, teamID uuid.UUID) service.NewDataproduct {
	return service.NewDataproduct{
		Name:          "Protective Barriers",
		Description:   strToStrPtr("Physical barriers to protect coral reefs from human activity"),
		Group:         group,
		ProductAreaID: &ProductAreaCostalID,
		TeamID:        &teamID,
	}
}

func StorageCreateDataproduct(t *testing.T, storage service.DataProductsStorage, ndp service.NewDataproduct) *service.DataproductMinimal {
	t.Helper()

	dp, err := storage.CreateDataproduct(context.Background(), ndp)
	if err != nil {
		t.Fatalf("creating dataproduct: %v", err)
	}

	return dp
}

func StorageCreateStory(t *testing.T, storage service.StoryStorage, creator string, ns service.NewStory) *service.Story {
	t.Helper()

	story, err := storage.CreateStory(context.Background(), creator, &ns)
	if err != nil {
		t.Fatalf("creating story: %v", err)
	}

	return story
}

func NewStoryBiofuelProduction(group string, teamId *uuid.UUID) service.NewStory {
	return service.NewStory{
		Name:        "Biofuel Production",
		Description: strToStrPtr("Using seagrass as a feedstock to create renewable biofuels"),
		Group:       group,
		Keywords:    []string{"biofuel", "production", "seagrass"},
		TeamID:      teamId,
	}
}

func NewStoryReefMonitoring(group string, teamId *uuid.UUID) service.NewStory {
	return service.NewStory{
		Name:        "Reef Monitoring Equipment",
		Description: strToStrPtr("Advanced sensors and monitoring devices for continuous assessment"),
		Group:       group,
		Keywords:    []string{"reef", "monitoring", "equipment"},
		TeamID:      teamId,
	}
}

func NewStoryProtectiveBarriers(group string, teamId *uuid.UUID) service.NewStory {
	return service.NewStory{
		Name:        "Protective Barriers",
		Description: strToStrPtr("Physical barriers to protect coral reefs from human activity"),
		Group:       group,
		Keywords:    []string{"protective", "barriers", "coral", "reefs"},
		TeamID:      teamId,
	}
}

func NewStoryAquacultureFeed(group string, teamId *uuid.UUID) service.NewStory {
	return service.NewStory{
		Name:        "Aquaculture Feed",
		Description: strToStrPtr("Producing high-nutrient feed for aquaculture industries from processed seagrass"),
		Group:       group,
		Keywords:    []string{"aquaculture", "feed", "seagrass"},
		TeamID:      teamId,
	}
}

func StorageCreateInsightProduct(t *testing.T, userEmail string, storage service.InsightProductStorage, nip service.NewInsightProduct) *service.InsightProduct {
	t.Helper()

	ip, err := storage.CreateInsightProduct(context.Background(), userEmail, nip)
	if err != nil {
		t.Fatalf("creating insight product: %v", err)
	}

	return ip
}

func NewInsightProductBiofuelProduction(group string, teamID uuid.UUID) service.NewInsightProduct {
	return service.NewInsightProduct{
		Name:          "Biofuel Production",
		Description:   strToStrPtr("Using seagrass as a feedstock to create renewable biofuels"),
		Group:         group,
		ProductAreaID: &ProductAreaOceanicID,
		TeamID:        &teamID,
	}
}

func NewInsightProductReefMonitoring(group string, teamID uuid.UUID) service.NewInsightProduct {
	return service.NewInsightProduct{
		Name:          "Reef Monitoring Equipment",
		Description:   strToStrPtr("Advanced sensors and monitoring devices for continuous assessment"),
		Group:         group,
		ProductAreaID: &ProductAreaCostalID,
		TeamID:        &teamID,
	}
}

func NewInsightProductProtectiveBarriers(group string, teamID uuid.UUID) service.NewInsightProduct {
	return service.NewInsightProduct{
		Name:          "Protective Barriers",
		Description:   strToStrPtr("Physical barriers to protect coral reefs from human activity"),
		Group:         group,
		ProductAreaID: &ProductAreaCostalID,
		TeamID:        &teamID,
	}
}

func NewInsightProductAquacultureFeed(group string, teamID uuid.UUID) service.NewInsightProduct {
	return service.NewInsightProduct{
		Name:          "Aquaculture Feed",
		Description:   strToStrPtr("Producing high-nutrient feed for aquaculture industries from processed seagrass"),
		Group:         group,
		ProductAreaID: &ProductAreaOceanicID,
		TeamID:        &teamID,
	}
}

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
