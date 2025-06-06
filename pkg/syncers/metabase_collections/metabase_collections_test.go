package metabase_collections_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/syncers/metabase_collections"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMetabaseAPI struct {
	mock.Mock
	service.MetabaseAPI
}

func (m *MockMetabaseAPI) GetCollections(ctx context.Context) ([]*service.MetabaseCollection, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*service.MetabaseCollection), args.Error(1)
}

func (m *MockMetabaseAPI) UpdateCollection(ctx context.Context, collection *service.MetabaseCollection) error {
	args := m.Called(ctx, collection)
	return args.Error(0)
}

type MockMetabaseStorage struct {
	mock.Mock
	service.MetabaseStorage
}

func (m *MockMetabaseStorage) GetAllMetadata(ctx context.Context) ([]*service.MetabaseMetadata, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*service.MetabaseMetadata), args.Error(1)
}

type MockMetabaseDataproductStorage struct {
	mock.Mock
	service.DataProductsStorage
}

func (m *MockMetabaseDataproductStorage) GetDataset(ctx context.Context, datasetID uuid.UUID) (*service.Dataset, error) {
	args := m.Called(ctx, datasetID)
	return args.Get(0).(*service.Dataset), args.Error(1)
}

func setupMockAPI() *MockMetabaseAPI {
	return new(MockMetabaseAPI)
}

func setupMockStorage() *MockMetabaseStorage {
	return new(MockMetabaseStorage)
}

func setupMockDataproductStorage() *MockMetabaseDataproductStorage {
	return new(MockMetabaseDataproductStorage)
}

func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func TestSyncer_MissingCollections(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name         string
		setupAPI     func(api *MockMetabaseAPI)
		setupStorage func(storage *MockMetabaseStorage)
		expect       *metabase_collections.CollectionsReport
		expectErr    error
	}{
		{
			name: "logs missing collections in metabase",
			setupAPI: func(api *MockMetabaseAPI) {
				api.On("GetCollections", ctx).Return([]*service.MetabaseCollection{}, nil)
			},
			setupStorage: func(storage *MockMetabaseStorage) {
				storage.On("GetAllMetadata", ctx).Return([]*service.MetabaseMetadata{
					{CollectionID: intPtr(1), DatasetID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), SyncCompleted: timePtr(time.Now()), DatabaseID: intPtr(0)},
				}, nil)
			},
			expect: &metabase_collections.CollectionsReport{
				Missing: []metabase_collections.Missing{
					{DatasetID: "00000000-0000-0000-0000-000000000001", CollectionID: 1, DatabaseID: 0},
				},
			},
		},
		{
			name: "logs missing collections in database",
			setupAPI: func(api *MockMetabaseAPI) {
				api.On("GetCollections", ctx).Return([]*service.MetabaseCollection{
					{ID: 1, Name: "collection1"},
				}, nil)
			},
			setupStorage: func(storage *MockMetabaseStorage) {
				storage.On("GetAllMetadata", ctx).Return([]*service.MetabaseMetadata{}, nil)
			},
			expect: &metabase_collections.CollectionsReport{
				Dangling: []metabase_collections.Dangling{
					{ID: 1, Name: "collection1"},
				},
			},
		},
		{
			name:     "handles storage error",
			setupAPI: func(api *MockMetabaseAPI) {},
			setupStorage: func(storage *MockMetabaseStorage) {
				storage.On("GetAllMetadata", ctx).Return([]*service.MetabaseMetadata{}, errors.New("storage error"))
			},
			expectErr: fmt.Errorf("storage error"),
		},
		{
			name: "handles API error",
			setupAPI: func(api *MockMetabaseAPI) {
				api.On("GetCollections", ctx).Return([]*service.MetabaseCollection{}, errors.New("api error"))
			},
			setupStorage: func(storage *MockMetabaseStorage) {
				storage.On("GetAllMetadata", ctx).Return([]*service.MetabaseMetadata{
					{CollectionID: intPtr(1), DatasetID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), SyncCompleted: timePtr(time.Now())},
				}, nil)
			},
			expectErr: fmt.Errorf("api error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			api := setupMockAPI()
			storage := setupMockStorage()
			dataproductStorage := setupMockDataproductStorage()
			syncer := metabase_collections.New(api, storage, dataproductStorage)

			tc.setupAPI(api)
			tc.setupStorage(storage)

			report, err := syncer.CollectionsReport(ctx)
			if tc.expectErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expect, report)
			}
		})
	}
}

func TestSyncer_AddRestrictedTagToCollections(t *testing.T) {
	logger := zerolog.New(zerolog.NewConsoleWriter())
	ctx := context.Background()

	testCases := []struct {
		name                    string
		setupAPI                func(api *MockMetabaseAPI)
		setupStorage            func(storage *MockMetabaseStorage)
		setupDataproductStorage func(storage *MockMetabaseDataproductStorage)
		expectErr               error
	}{
		{
			name: "updates collection name",
			setupAPI: func(api *MockMetabaseAPI) {
				api.On("GetCollections", ctx).Return([]*service.MetabaseCollection{
					{ID: 1, Name: "collection1"},
				}, nil)
				api.On("UpdateCollection", ctx, &service.MetabaseCollection{
					ID:   1,
					Name: "collection1 🔐",
				}).Return(nil)
			},
			setupStorage: func(storage *MockMetabaseStorage) {
				storage.On("GetAllMetadata", ctx).Return([]*service.MetabaseMetadata{
					{CollectionID: intPtr(1), DatasetID: uuid.MustParse("00000000-0000-0000-0000-000000000001")},
				}, nil)
			},
			setupDataproductStorage: func(storage *MockMetabaseDataproductStorage) {
				storage.On("GetDataset", ctx, uuid.MustParse("00000000-0000-0000-0000-000000000001")).Return(&service.Dataset{
					ID:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Name: "Dataset 1",
					Datasource: &service.BigQuery{
						ProjectID: "project1",
						Dataset:   "dataset1",
						Table:     "table1",
					},
				}, nil)
			},
		},
		{
			name: "handles update error",
			setupAPI: func(api *MockMetabaseAPI) {
				api.On("GetCollections", ctx).Return([]*service.MetabaseCollection{
					{ID: 1, Name: "collection1"},
				}, nil)
				api.On("UpdateCollection", ctx, &service.MetabaseCollection{
					ID:          1,
					Name:        "collection1 🔐",
					Description: "Dette er en tilgangsstyrt samling for BigQuery tabellen: project1.dataset1.table1. I markedsplassen er dette datasettet 00000000-0000-0000-0000-000000000001, i dataprodutet 00000000-0000-0000-0000-000000000000",
				}).Return(errors.New("update error"))
			},
			setupStorage: func(storage *MockMetabaseStorage) {
				storage.On("GetAllMetadata", ctx).Return([]*service.MetabaseMetadata{
					{CollectionID: intPtr(1), DatasetID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), SyncCompleted: timePtr(time.Now()), DatabaseID: intPtr(0)},
				}, nil)
			},
			setupDataproductStorage: func(storage *MockMetabaseDataproductStorage) {
				storage.On("GetDataset", ctx, uuid.MustParse("00000000-0000-0000-0000-000000000001")).Return(&service.Dataset{
					ID:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Name: "Dataset 1",
					Datasource: &service.BigQuery{
						ProjectID: "project1",
						Dataset:   "dataset1",
						Table:     "table1",
					},
				}, nil)
			},
			expectErr: fmt.Errorf("updating collection metadata 1: update error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			api := setupMockAPI()
			storage := setupMockStorage()
			dataproductStorage := setupMockDataproductStorage()
			syncer := metabase_collections.New(api, storage, dataproductStorage)

			tc.setupAPI(api)
			tc.setupStorage(storage)
			tc.setupDataproductStorage(dataproductStorage)

			err := syncer.UpdateCollectionMetadata(ctx, logger)
			if tc.expectErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSyncer_Run(t *testing.T) {
	logger := zerolog.New(zerolog.NewConsoleWriter())
	ctx := context.Background()

	testCases := []struct {
		name               string
		setupAPI           func(api *MockMetabaseAPI)
		setupStorage       func(storage *MockMetabaseStorage)
		dataproductStorage func(*MockMetabaseDataproductStorage)
		expectErr          bool
		expect             string
	}{
		{
			name: "runs syncer and logs missing collections",
			setupAPI: func(api *MockMetabaseAPI) {
				api.On("GetCollections", ctx).Return([]*service.MetabaseCollection{
					{ID: 1, Name: "collection1"},
				}, nil)
				api.On("UpdateCollection", ctx, mock.Anything).Return(nil)
			},
			setupStorage: func(storage *MockMetabaseStorage) {
				storage.On("GetAllMetadata", ctx).Return([]*service.MetabaseMetadata{
					{CollectionID: intPtr(1), DatasetID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), SyncCompleted: timePtr(time.Now()), DatabaseID: intPtr(10)},
				}, nil)
			},
			dataproductStorage: func(storage *MockMetabaseDataproductStorage) {
				storage.On("GetDataset", ctx, uuid.MustParse("00000000-0000-0000-0000-000000000001")).Return(&service.Dataset{
					ID:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Name: "Dataset 1",
					Datasource: &service.BigQuery{
						ProjectID: "project1",
						Dataset:   "dataset1",
						Table:     "table1",
					},
				}, nil)
			},
			expectErr: false,
		},
		{
			name: "handles UpdateCollectionMetadata error",
			setupAPI: func(api *MockMetabaseAPI) {
				api.On("GetCollections", ctx).Return([]*service.MetabaseCollection{
					{ID: 1, Name: "collection1"},
				}, nil)
				api.On("UpdateCollection", ctx, mock.Anything).Return(fmt.Errorf("update error"))
			},
			setupStorage: func(storage *MockMetabaseStorage) {
				storage.On("GetAllMetadata", ctx).Return([]*service.MetabaseMetadata{
					{CollectionID: intPtr(1), DatasetID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), SyncCompleted: timePtr(time.Now()), DatabaseID: intPtr(10)},
				}, nil)
			},
			dataproductStorage: func(storage *MockMetabaseDataproductStorage) {
				storage.On("GetDataset", ctx, uuid.MustParse("00000000-0000-0000-0000-000000000001")).Return(&service.Dataset{
					ID:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Name: "Dataset 1",
					Datasource: &service.BigQuery{
						ProjectID: "project1",
						Dataset:   "dataset1",
						Table:     "table1",
					},
				}, nil)
			},
			expectErr: true,
			expect:    "updating metadata for collection: updating collection metadata 1: update error",
		},
		{
			name: "handles CollectionsReport error",
			setupAPI: func(api *MockMetabaseAPI) {
				api.On("GetCollections", ctx).Return([]*service.MetabaseCollection{}, errors.New("api error"))
			},
			setupStorage: func(storage *MockMetabaseStorage) {
				storage.On("GetAllMetadata", ctx).Return([]*service.MetabaseMetadata{
					{CollectionID: intPtr(1), DatasetID: uuid.MustParse("00000000-0000-0000-0000-000000000001")},
				}, nil)
			},
			dataproductStorage: func(storage *MockMetabaseDataproductStorage) {
				storage.On("GetDataset", ctx, uuid.MustParse("00000000-0000-0000-0000-000000000001")).Return(&service.Dataset{
					ID:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Name: "Dataset 1",
					Datasource: &service.BigQuery{
						ProjectID: "project1",
						Dataset:   "dataset1",
						Table:     "table1",
					},
				}, nil)
			},
			expectErr: true,
			expect:    "updating metadata for collection: api error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			api := setupMockAPI()
			storage := setupMockStorage()
			dataproductStorage := setupMockDataproductStorage()
			syncer := metabase_collections.New(api, storage, dataproductStorage)

			tc.setupAPI(api)
			tc.setupStorage(storage)
			tc.dataproductStorage(dataproductStorage)

			err := syncer.RunOnce(ctx, logger)
			if tc.expectErr {
				require.Error(t, err)
				assert.Equal(t, tc.expect, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
