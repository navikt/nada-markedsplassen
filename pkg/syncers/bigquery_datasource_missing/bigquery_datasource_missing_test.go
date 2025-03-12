package bigquery_datasource_missing_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/errs"

	"github.com/navikt/nada-backend/pkg/bq"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/syncers/bigquery_datasource_missing"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBigQueryOps struct {
	mock.Mock
	bq.Operations
}

func (m *MockBigQueryOps) GetTable(ctx context.Context, projectID, dataset, table string) (*bq.Table, error) {
	args := m.Called(ctx, projectID, dataset, table)
	return nil, args.Error(0)
}

type MockBigQueryStorage struct {
	mock.Mock
	service.BigQueryStorage
}

func (m *MockBigQueryStorage) GetBigqueryDatasources(ctx context.Context) ([]*service.BigQuery, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*service.BigQuery), args.Error(1)
}

func (m *MockBigQueryStorage) UpdateBigqueryDatasourceNotMissing(ctx context.Context, datasetID uuid.UUID) error {
	args := m.Called(ctx, datasetID)
	return args.Error(0)
}

func (m *MockBigQueryStorage) DeleteBigqueryDatasource(ctx context.Context, datasetID uuid.UUID) error {
	args := m.Called(ctx, datasetID)
	return args.Error(0)
}

type MockMetabaseService struct {
	mock.Mock
	service.MetabaseService
}

func (m *MockMetabaseService) DeleteDatabase(ctx context.Context, datasetID uuid.UUID) error {
	args := m.Called(ctx, datasetID)
	return args.Error(0)
}

type MockMetabaseStorage struct {
	mock.Mock
	service.MetabaseStorage
}

func (m *MockMetabaseStorage) GetMetadata(ctx context.Context, datasetID uuid.UUID, includeDeleted bool) (*service.MetabaseMetadata, error) {
	args := m.Called(ctx, datasetID, includeDeleted)
	return nil, args.Error(0)
}

type MockDataProductsStorage struct {
	mock.Mock
	service.DataProductsStorage
}

func (m *MockDataProductsStorage) DeleteDataset(ctx context.Context, datasetID uuid.UUID) error {
	args := m.Called(ctx, datasetID)
	return args.Error(0)
}

func TestBigQueryDatasourceMissing(t *testing.T) {
	dsid := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	testCases := []struct {
		name           string
		setupMocks     func(*MockBigQueryOps, *MockBigQueryStorage, *MockMetabaseService, *MockMetabaseStorage, *MockDataProductsStorage, context.Context)
		expectedError  bool
		errorSubstring string
	}{
		{
			name: "all datasources exist",
			setupMocks: func(ops *MockBigQueryOps, storage *MockBigQueryStorage, mbSvc *MockMetabaseService, mbStorage *MockMetabaseStorage, dpStorage *MockDataProductsStorage, ctx context.Context) {
				storage.On("GetBigqueryDatasources", ctx).Return([]*service.BigQuery{
					{ProjectID: "project1", Dataset: "dataset1", Table: "table1"},
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "datasource missing only for a little while",
			setupMocks: func(ops *MockBigQueryOps, storage *MockBigQueryStorage, mbSvc *MockMetabaseService, mbStorage *MockMetabaseStorage, dpStorage *MockDataProductsStorage, ctx context.Context) {
				missingSince := time.Now().Add(-time.Hour * 24 * 6)

				storage.On("GetBigqueryDatasources", ctx).Return([]*service.BigQuery{
					{ProjectID: "project1", Dataset: "dataset1", Table: "table1", DatasetID: dsid, MissingSince: &missingSince},
				}, nil)
				ops.On("GetTable", ctx, "project1", "dataset1", "table1").Return(bq.ErrNotExist)
			},
			expectedError: false,
		},
		{
			name: "missing datasources for more than a week",
			setupMocks: func(ops *MockBigQueryOps, storage *MockBigQueryStorage, mbSvc *MockMetabaseService, mbStorage *MockMetabaseStorage, dpStorage *MockDataProductsStorage, ctx context.Context) {
				missingSince := time.Now().Add(-time.Hour * 24 * 8)
				storage.On("GetBigqueryDatasources", ctx).Return([]*service.BigQuery{
					{ProjectID: "project1", Dataset: "dataset1", Table: "table1", DatasetID: dsid, MissingSince: &missingSince},
				}, nil)
				ops.On("GetTable", ctx, "project1", "dataset1", "table1").Return(bq.ErrNotExist)
				mbStorage.On("GetMetadata", ctx, dsid, true).Return(errs.E(errs.NotExist, errors.New("not found")))
				dpStorage.On("DeleteDataset", ctx, dsid).Return(nil)
				storage.On("DeleteBigqueryDatasource", ctx, dsid).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "error getting datasources",
			setupMocks: func(ops *MockBigQueryOps, storage *MockBigQueryStorage, mbSvc *MockMetabaseService, mbStorage *MockMetabaseStorage, dpStorage *MockDataProductsStorage, ctx context.Context) {
				storage.On("GetBigqueryDatasources", ctx).Return([]*service.BigQuery{}, errors.New("database error"))
			},
			expectedError:  true,
			errorSubstring: "database error",
		},
		{
			name: "error deleting dataset",
			setupMocks: func(ops *MockBigQueryOps, storage *MockBigQueryStorage, mbSvc *MockMetabaseService, mbStorage *MockMetabaseStorage, dpStorage *MockDataProductsStorage, ctx context.Context) {
				missingSince := time.Now().Add(-time.Hour * 24 * 8)
				storage.On("GetBigqueryDatasources", ctx).Return([]*service.BigQuery{
					{ProjectID: "project1", Dataset: "dataset1", Table: "table1", DatasetID: dsid, MissingSince: &missingSince},
				}, nil)
				ops.On("GetTable", ctx, "project1", "dataset1", "table1").Return(bq.ErrNotExist)
				mbStorage.On("GetMetadata", ctx, dsid, true).Return(nil)
				mbSvc.On("DeleteDatabase", ctx, dsid).Return(nil)
				dpStorage.On("DeleteDataset", ctx, dsid).Return(errors.New("delete error"))
			},
			expectedError:  true,
			errorSubstring: "delete error",
		},
		{
			name: "table exists - update datasource not missing",
			setupMocks: func(ops *MockBigQueryOps, storage *MockBigQueryStorage, mbSvc *MockMetabaseService, mbStorage *MockMetabaseStorage, dpStorage *MockDataProductsStorage, ctx context.Context) {
				missingSince := time.Now().Add(-time.Hour * 24 * 6)
				storage.On("GetBigqueryDatasources", ctx).Return([]*service.BigQuery{
					{ProjectID: "project1", Dataset: "dataset1", Table: "table1", DatasetID: dsid, MissingSince: &missingSince},
				}, nil)
				ops.On("GetTable", ctx, "project1", "dataset1", "table1").Return(nil, nil)
				storage.On("UpdateBigqueryDatasourceNotMissing", ctx, dsid).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "error getting table status",
			setupMocks: func(ops *MockBigQueryOps, storage *MockBigQueryStorage, mbSvc *MockMetabaseService, mbStorage *MockMetabaseStorage, dpStorage *MockDataProductsStorage, ctx context.Context) {
				missingSince := time.Now().Add(-time.Hour * 24 * 6)

				storage.On("GetBigqueryDatasources", ctx).Return([]*service.BigQuery{
					{ProjectID: "project1", Dataset: "dataset1", Table: "table1", DatasetID: dsid, MissingSince: &missingSince},
				}, nil)

				ops.On("GetTable", ctx, "project1", "dataset1", "table1").Return(errors.New("connection error"))
			},
			expectedError:  true,
			errorSubstring: "connection error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			log := zerolog.Nop()

			bigQueryOps := new(MockBigQueryOps)
			bigQueryStorage := new(MockBigQueryStorage)
			metabaseService := new(MockMetabaseService)
			metabaseStorage := new(MockMetabaseStorage)
			dataProductStorage := new(MockDataProductsStorage)

			tc.setupMocks(bigQueryOps, bigQueryStorage, metabaseService, metabaseStorage, dataProductStorage, ctx)

			runner := bigquery_datasource_missing.New(
				bigQueryOps,
				bigQueryStorage,
				metabaseService,
				metabaseStorage,
				dataProductStorage,
			)

			err := runner.RunOnce(ctx, log)

			if tc.expectedError {
				assert.Error(t, err)
				if tc.errorSubstring != "" {
					assert.Contains(t, err.Error(), tc.errorSubstring)
				}
			} else {
				assert.NoError(t, err)
			}

			bigQueryOps.AssertExpectations(t)
			bigQueryStorage.AssertExpectations(t)
			metabaseService.AssertExpectations(t)
			metabaseStorage.AssertExpectations(t)
			dataProductStorage.AssertExpectations(t)
		})
	}
}
