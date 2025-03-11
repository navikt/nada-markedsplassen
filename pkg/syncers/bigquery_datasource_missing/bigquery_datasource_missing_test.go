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

func TestRunOnce_AllDatasourcesExist(t *testing.T) {
	ctx := context.Background()
	log := zerolog.New(zerolog.NewConsoleWriter())

	bigQueryOps := new(MockBigQueryOps)
	bigQueryStorage := new(MockBigQueryStorage)
	metabaseService := new(MockMetabaseService)
	metabaseStorage := new(MockMetabaseStorage)
	dataProductStorage := new(MockDataProductsStorage)

	runner := bigquery_datasource_missing.New(
		bigQueryOps,
		bigQueryStorage,
		metabaseService,
		metabaseStorage,
		dataProductStorage,
	)

	datasources := []*service.BigQuery{
		{ProjectID: "project1", Dataset: "dataset1", Table: "table1"},
	}

	bigQueryStorage.On("GetBigqueryDatasources", ctx).Return(datasources, nil)

	err := runner.RunOnce(ctx, log)
	assert.NoError(t, err)

	bigQueryStorage.AssertExpectations(t)
	bigQueryOps.AssertExpectations(t)
}

func TestRunOnce_MissingOnlyALittleWhile(t *testing.T) {
	ctx := context.Background()
	log := zerolog.New(zerolog.NewConsoleWriter())

	bigQueryOps := new(MockBigQueryOps)
	bigQueryStorage := new(MockBigQueryStorage)
	metabaseService := new(MockMetabaseService)
	metabaseStorage := new(MockMetabaseStorage)
	dataProductStorage := new(MockDataProductsStorage)

	runner := bigquery_datasource_missing.New(
		bigQueryOps,
		bigQueryStorage,
		metabaseService,
		metabaseStorage,
		dataProductStorage,
	)

	missingSince := time.Now().Add(-time.Hour * 24 * 6)
	dsid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	datasources := []*service.BigQuery{
		{ProjectID: "project1", Dataset: "dataset1", Table: "table1", DatasetID: dsid, MissingSince: &missingSince},
	}

	bigQueryStorage.On("GetBigqueryDatasources", ctx).Return(datasources, nil).Once()
	bigQueryOps.On("GetTable", ctx, "project1", "dataset1", "table1").Return(bq.ErrNotExist).Once()

	err := runner.RunOnce(ctx, log)
	assert.NoError(t, err)

	bigQueryStorage.AssertExpectations(t)
	bigQueryOps.AssertExpectations(t)
	metabaseStorage.AssertExpectations(t)
	dataProductStorage.AssertExpectations(t)
}

func TestRunOnce_MissingDatasources(t *testing.T) {
	ctx := context.Background()
	log := zerolog.New(zerolog.NewConsoleWriter())

	bigQueryOps := new(MockBigQueryOps)
	bigQueryStorage := new(MockBigQueryStorage)
	metabaseService := new(MockMetabaseService)
	metabaseStorage := new(MockMetabaseStorage)
	dataProductStorage := new(MockDataProductsStorage)

	runner := bigquery_datasource_missing.New(
		bigQueryOps,
		bigQueryStorage,
		metabaseService,
		metabaseStorage,
		dataProductStorage,
	)

	missingSince := time.Now().Add(-time.Hour * 24 * 8)
	dsid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	datasources := []*service.BigQuery{
		{ProjectID: "project1", Dataset: "dataset1", Table: "table1", DatasetID: dsid, MissingSince: &missingSince},
	}

	bigQueryStorage.On("GetBigqueryDatasources", ctx).Return(datasources, nil).Once()
	bigQueryOps.On("GetTable", ctx, "project1", "dataset1", "table1").Return(bq.ErrNotExist).Once()
	metabaseStorage.On("GetMetadata", ctx, dsid, true).Return(errs.E(errs.NotExist, errors.New("not found"))).Once()
	dataProductStorage.On("DeleteDataset", ctx, dsid).Return(nil).Once()

	err := runner.RunOnce(ctx, log)
	assert.NoError(t, err)

	bigQueryStorage.AssertExpectations(t)
	bigQueryOps.AssertExpectations(t)
	metabaseStorage.AssertExpectations(t)
	dataProductStorage.AssertExpectations(t)
}

func TestRunOnce_ErrorGettingDatasources(t *testing.T) {
	ctx := context.Background()
	log := zerolog.New(zerolog.NewConsoleWriter())

	bigQueryOps := new(MockBigQueryOps)
	bigQueryStorage := new(MockBigQueryStorage)
	metabaseService := new(MockMetabaseService)
	metabaseStorage := new(MockMetabaseStorage)
	dataProductStorage := new(MockDataProductsStorage)

	runner := bigquery_datasource_missing.New(
		bigQueryOps,
		bigQueryStorage,
		metabaseService,
		metabaseStorage,
		dataProductStorage,
	)

	bigQueryStorage.On("GetBigqueryDatasources", ctx).Return([]*service.BigQuery{}, errors.New("error"))

	err := runner.RunOnce(ctx, log)
	assert.Error(t, err)

	bigQueryStorage.AssertExpectations(t)
}

func TestRunOnce_ErrorDeletingDataset(t *testing.T) {
	ctx := context.Background()
	log := zerolog.New(zerolog.NewConsoleWriter())

	bigQueryOps := new(MockBigQueryOps)
	bigQueryStorage := new(MockBigQueryStorage)
	metabaseService := new(MockMetabaseService)
	metabaseStorage := new(MockMetabaseStorage)
	dataProductStorage := new(MockDataProductsStorage)

	runner := bigquery_datasource_missing.New(
		bigQueryOps,
		bigQueryStorage,
		metabaseService,
		metabaseStorage,
		dataProductStorage,
	)

	missingSince := time.Now().Add(-time.Hour * 24 * 8)
	dsid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	datasources := []*service.BigQuery{
		{ProjectID: "project1", Dataset: "dataset1", Table: "table1", DatasetID: dsid, MissingSince: &missingSince},
	}

	bigQueryStorage.On("GetBigqueryDatasources", ctx).Return(datasources, nil)
	bigQueryOps.On("GetTable", ctx, "project1", "dataset1", "table1").Return(bq.ErrNotExist)
	metabaseService.On("DeleteDatabase", ctx, dsid).Return(nil)
	metabaseStorage.On("GetMetadata", ctx, dsid, true).Return(nil)
	dataProductStorage.On("DeleteDataset", ctx, dsid).Return(errors.New("delete error"))

	err := runner.RunOnce(ctx, log)
	assert.Error(t, err)

	bigQueryStorage.AssertExpectations(t)
	bigQueryOps.AssertExpectations(t)
	metabaseStorage.AssertExpectations(t)
	dataProductStorage.AssertExpectations(t)
}
