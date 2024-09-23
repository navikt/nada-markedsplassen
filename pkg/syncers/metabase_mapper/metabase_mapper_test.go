package metabase_mapper_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/navikt/nada-backend/pkg/service"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/syncers/metabase_mapper"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMetabaseService struct {
	service.MetabaseService
	mock.Mock
}

func (m *MockMetabaseService) MapDataset(ctx context.Context, datasetID uuid.UUID, services []string) error {
	args := m.Called(ctx, datasetID, services)
	return args.Error(0)
}

type MockThirdPartyMappingStorage struct {
	service.ThirdPartyMappingStorage
	mock.Mock
}

func (m *MockThirdPartyMappingStorage) GetAddMetabaseDatasetMappings(ctx context.Context) ([]uuid.UUID, error) {
	args := m.Called(ctx)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *MockThirdPartyMappingStorage) GetRemoveMetabaseDatasetMappings(ctx context.Context) ([]uuid.UUID, error) {
	args := m.Called(ctx)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func TestMapperProcessesDatasetsFromQueue(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	mockService := new(MockMetabaseService)
	mockStorage := new(MockThirdPartyMappingStorage)
	queue := make(chan metabase_mapper.Work, 10)
	mapper := metabase_mapper.New(mockService, mockStorage, 20, queue, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	datasetID := uuid.New()
	mockService.On("MapDataset", mock.Anything, datasetID, mock.Anything).Return(nil)

	go mapper.ProcessQueue(ctx)
	queue <- metabase_mapper.Work{
		DatasetID: datasetID,
	}

	time.Sleep(2 * time.Second)

	mockService.AssertCalled(t, "MapDataset", mock.Anything, datasetID, mock.Anything)
}

func TestMapperProcessesDatasetsFromStorage(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	mockService := new(MockMetabaseService)
	mockStorage := new(MockThirdPartyMappingStorage)
	queue := make(chan metabase_mapper.Work, 10)
	mapper := metabase_mapper.New(mockService, mockStorage, 10, queue, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	datasetID := uuid.New()
	mockStorage.On("GetAddMetabaseDatasetMappings", mock.Anything).Return([]uuid.UUID{datasetID}, nil)
	mockStorage.On("GetRemoveMetabaseDatasetMappings", mock.Anything).Return([]uuid.UUID{}, nil)
	mockService.On("MapDataset", mock.Anything, datasetID, mock.Anything).Return(nil)

	go mapper.ProcessQueue(ctx)
	err := mapper.RunOnce(ctx, logger)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	mockService.AssertCalled(t, "MapDataset", mock.Anything, datasetID, mock.Anything)
}

func TestMapperHandlesMappingError(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	mockService := new(MockMetabaseService)
	mockStorage := new(MockThirdPartyMappingStorage)
	queue := make(chan metabase_mapper.Work, 10)
	mapper := metabase_mapper.New(mockService, mockStorage, 10, queue, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	datasetID := uuid.New()
	mockService.On("MapDataset", mock.Anything, datasetID, mock.Anything).Return(assert.AnError)

	go mapper.ProcessQueue(ctx)
	queue <- metabase_mapper.Work{
		DatasetID: datasetID,
	}

	time.Sleep(2 * time.Second)

	mockService.AssertCalled(t, "MapDataset", mock.Anything, datasetID, mock.Anything)
}

func TestMapperShutsDownGracefully(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	mockService := new(MockMetabaseService)
	mockStorage := new(MockThirdPartyMappingStorage)
	queue := make(chan metabase_mapper.Work, 10)
	mapper := metabase_mapper.New(mockService, mockStorage, 10, queue, logger)

	ctx, cancel := context.WithCancel(context.Background())

	go mapper.ProcessQueue(ctx)
	cancel()

	time.Sleep(1 * time.Second)

	assert.NotPanics(t, func() { cancel() })
}
