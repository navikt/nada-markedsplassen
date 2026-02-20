package metabase_access_remover_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/syncers/metabase_access_remover"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// -- mocks --

type MockMetabaseAPI struct {
	mock.Mock
	service.MetabaseAPI
}

func (m *MockMetabaseAPI) GetPermissionGroup(ctx context.Context, groupID int) ([]service.MetabasePermissionGroupMember, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).([]service.MetabasePermissionGroupMember), args.Error(1)
}

func (m *MockMetabaseAPI) RemovePermissionGroupMember(ctx context.Context, memberID int) error {
	args := m.Called(ctx, memberID)
	return args.Error(0)
}

type MockMetabaseStorage struct {
	mock.Mock
	service.RestrictedMetabaseStorage
}

func (m *MockMetabaseStorage) GetAllMetadata(ctx context.Context) ([]*service.RestrictedMetabaseMetadata, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*service.RestrictedMetabaseMetadata), args.Error(1)
}

type MockAccessStorage struct {
	mock.Mock
	service.AccessStorage
}

func (m *MockAccessStorage) ListActiveAccessToDataset(ctx context.Context, datasetID uuid.UUID) ([]*service.Access, error) {
	args := m.Called(ctx, datasetID)
	return args.Get(0).([]*service.Access), args.Error(1)
}

// -- helpers --

func intPtr(i int) *int    { return &i }
func nowPtr() *time.Time   { t := time.Now(); return &t }

var (
	ds1 = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	ds2 = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

// -- tests --

func TestRunOnce_StorageError(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	api := new(MockMetabaseAPI)
	mbStorage := new(MockMetabaseStorage)
	accessStorage := new(MockAccessStorage)

	mbStorage.On("GetAllMetadata", ctx).Return([]*service.RestrictedMetabaseMetadata{}, errors.New("db error"))

	runner := metabase_access_remover.New(api, mbStorage, accessStorage)
	err := runner.RunOnce(ctx, log)

	assert.Error(t, err)
	assert.Equal(t, "getting all restricted metabase metadata: db error", err.Error())
	api.AssertExpectations(t)
	mbStorage.AssertExpectations(t)
	accessStorage.AssertExpectations(t)
}

func TestRunOnce_SkipsDatasetWithNilPermissionGroupID(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	api := new(MockMetabaseAPI)
	mbStorage := new(MockMetabaseStorage)
	accessStorage := new(MockAccessStorage)

	mbStorage.On("GetAllMetadata", ctx).Return([]*service.RestrictedMetabaseMetadata{
		{DatasetID: ds1, PermissionGroupID: nil, SyncCompleted: nowPtr()},
	}, nil)

	runner := metabase_access_remover.New(api, mbStorage, accessStorage)
	err := runner.RunOnce(ctx, log)

	assert.NoError(t, err)
	api.AssertNotCalled(t, "GetPermissionGroup")
	accessStorage.AssertNotCalled(t, "ListActiveAccessToDataset")
}

func TestRunOnce_SkipsDatasetWithNilSyncCompleted(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	api := new(MockMetabaseAPI)
	mbStorage := new(MockMetabaseStorage)
	accessStorage := new(MockAccessStorage)

	mbStorage.On("GetAllMetadata", ctx).Return([]*service.RestrictedMetabaseMetadata{
		{DatasetID: ds1, PermissionGroupID: intPtr(10), SyncCompleted: nil},
	}, nil)

	runner := metabase_access_remover.New(api, mbStorage, accessStorage)
	err := runner.RunOnce(ctx, log)

	assert.NoError(t, err)
	api.AssertNotCalled(t, "GetPermissionGroup")
	accessStorage.AssertNotCalled(t, "ListActiveAccessToDataset")
}

func TestRunOnce_NoStaleMembers(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	api := new(MockMetabaseAPI)
	mbStorage := new(MockMetabaseStorage)
	accessStorage := new(MockAccessStorage)

	mbStorage.On("GetAllMetadata", ctx).Return([]*service.RestrictedMetabaseMetadata{
		{DatasetID: ds1, PermissionGroupID: intPtr(10), SyncCompleted: nowPtr()},
	}, nil)
	api.On("GetPermissionGroup", ctx, 10).Return([]service.MetabasePermissionGroupMember{
		{ID: 1, Email: "alice@example.com"},
	}, nil)
	accessStorage.On("ListActiveAccessToDataset", ctx, ds1).Return([]*service.Access{
		{Subject: "user:alice@example.com", Platform: service.AccessPlatformMetabase},
	}, nil)

	runner := metabase_access_remover.New(api, mbStorage, accessStorage)
	err := runner.RunOnce(ctx, log)

	assert.NoError(t, err)
	api.AssertNotCalled(t, "RemovePermissionGroupMember")
	api.AssertExpectations(t)
	mbStorage.AssertExpectations(t)
	accessStorage.AssertExpectations(t)
}

func TestRunOnce_RemovesStaleMember(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	api := new(MockMetabaseAPI)
	mbStorage := new(MockMetabaseStorage)
	accessStorage := new(MockAccessStorage)

	mbStorage.On("GetAllMetadata", ctx).Return([]*service.RestrictedMetabaseMetadata{
		{DatasetID: ds1, PermissionGroupID: intPtr(10), SyncCompleted: nowPtr()},
	}, nil)
	api.On("GetPermissionGroup", ctx, 10).Return([]service.MetabasePermissionGroupMember{
		{ID: 99, Email: "stale@example.com"},
	}, nil)
	accessStorage.On("ListActiveAccessToDataset", ctx, ds1).Return([]*service.Access{}, nil)
	api.On("RemovePermissionGroupMember", ctx, 99).Return(nil)

	runner := metabase_access_remover.New(api, mbStorage, accessStorage)
	err := runner.RunOnce(ctx, log)

	assert.NoError(t, err)
	api.AssertExpectations(t)
	mbStorage.AssertExpectations(t)
	accessStorage.AssertExpectations(t)
}

func TestRunOnce_IgnoresBigQueryPlatformAccess(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	api := new(MockMetabaseAPI)
	mbStorage := new(MockMetabaseStorage)
	accessStorage := new(MockAccessStorage)

	// alice has BigQuery access only — she should NOT be considered an expected member
	mbStorage.On("GetAllMetadata", ctx).Return([]*service.RestrictedMetabaseMetadata{
		{DatasetID: ds1, PermissionGroupID: intPtr(10), SyncCompleted: nowPtr()},
	}, nil)
	api.On("GetPermissionGroup", ctx, 10).Return([]service.MetabasePermissionGroupMember{
		{ID: 99, Email: "alice@example.com"},
	}, nil)
	accessStorage.On("ListActiveAccessToDataset", ctx, ds1).Return([]*service.Access{
		{Subject: "user:alice@example.com", Platform: service.AccessPlatformBigQuery},
	}, nil)
	api.On("RemovePermissionGroupMember", ctx, 99).Return(nil)

	runner := metabase_access_remover.New(api, mbStorage, accessStorage)
	err := runner.RunOnce(ctx, log)

	assert.NoError(t, err)
	api.AssertExpectations(t)
}

func TestRunOnce_MultipleStaleMembers(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	api := new(MockMetabaseAPI)
	mbStorage := new(MockMetabaseStorage)
	accessStorage := new(MockAccessStorage)

	mbStorage.On("GetAllMetadata", ctx).Return([]*service.RestrictedMetabaseMetadata{
		{DatasetID: ds1, PermissionGroupID: intPtr(10), SyncCompleted: nowPtr()},
	}, nil)
	api.On("GetPermissionGroup", ctx, 10).Return([]service.MetabasePermissionGroupMember{
		{ID: 1, Email: "alice@example.com"},
		{ID: 2, Email: "bob@example.com"},
		{ID: 3, Email: "carol@example.com"},
	}, nil)
	accessStorage.On("ListActiveAccessToDataset", ctx, ds1).Return([]*service.Access{
		{Subject: "user:alice@example.com", Platform: service.AccessPlatformMetabase},
	}, nil)
	api.On("RemovePermissionGroupMember", ctx, 2).Return(nil)
	api.On("RemovePermissionGroupMember", ctx, 3).Return(nil)

	runner := metabase_access_remover.New(api, mbStorage, accessStorage)
	err := runner.RunOnce(ctx, log)

	assert.NoError(t, err)
	api.AssertExpectations(t)
}

func TestRunOnce_MultipleDatasets(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	api := new(MockMetabaseAPI)
	mbStorage := new(MockMetabaseStorage)
	accessStorage := new(MockAccessStorage)

	mbStorage.On("GetAllMetadata", ctx).Return([]*service.RestrictedMetabaseMetadata{
		{DatasetID: ds1, PermissionGroupID: intPtr(10), SyncCompleted: nowPtr()},
		{DatasetID: ds2, PermissionGroupID: intPtr(20), SyncCompleted: nowPtr()},
	}, nil)

	api.On("GetPermissionGroup", ctx, 10).Return([]service.MetabasePermissionGroupMember{
		{ID: 1, Email: "stale@example.com"},
	}, nil)
	accessStorage.On("ListActiveAccessToDataset", ctx, ds1).Return([]*service.Access{}, nil)
	api.On("RemovePermissionGroupMember", ctx, 1).Return(nil)

	api.On("GetPermissionGroup", ctx, 20).Return([]service.MetabasePermissionGroupMember{
		{ID: 2, Email: "active@example.com"},
	}, nil)
	accessStorage.On("ListActiveAccessToDataset", ctx, ds2).Return([]*service.Access{
		{Subject: "user:active@example.com", Platform: service.AccessPlatformMetabase},
	}, nil)

	runner := metabase_access_remover.New(api, mbStorage, accessStorage)
	err := runner.RunOnce(ctx, log)

	assert.NoError(t, err)
	api.AssertExpectations(t)
	accessStorage.AssertExpectations(t)
}

func TestRunOnce_ContinuesWhenGetPermissionGroupFails(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	api := new(MockMetabaseAPI)
	mbStorage := new(MockMetabaseStorage)
	accessStorage := new(MockAccessStorage)

	mbStorage.On("GetAllMetadata", ctx).Return([]*service.RestrictedMetabaseMetadata{
		{DatasetID: ds1, PermissionGroupID: intPtr(10), SyncCompleted: nowPtr()},
		{DatasetID: ds2, PermissionGroupID: intPtr(20), SyncCompleted: nowPtr()},
	}, nil)

	api.On("GetPermissionGroup", ctx, 10).Return([]service.MetabasePermissionGroupMember{}, errors.New("api error"))
	// ds2 should still be processed
	api.On("GetPermissionGroup", ctx, 20).Return([]service.MetabasePermissionGroupMember{
		{ID: 2, Email: "stale@example.com"},
	}, nil)
	accessStorage.On("ListActiveAccessToDataset", ctx, ds2).Return([]*service.Access{}, nil)
	api.On("RemovePermissionGroupMember", ctx, 2).Return(nil)

	runner := metabase_access_remover.New(api, mbStorage, accessStorage)
	err := runner.RunOnce(ctx, log)

	assert.NoError(t, err)
	api.AssertExpectations(t)
	accessStorage.AssertExpectations(t)
}

func TestRunOnce_ContinuesWhenListActiveAccessFails(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	api := new(MockMetabaseAPI)
	mbStorage := new(MockMetabaseStorage)
	accessStorage := new(MockAccessStorage)

	mbStorage.On("GetAllMetadata", ctx).Return([]*service.RestrictedMetabaseMetadata{
		{DatasetID: ds1, PermissionGroupID: intPtr(10), SyncCompleted: nowPtr()},
		{DatasetID: ds2, PermissionGroupID: intPtr(20), SyncCompleted: nowPtr()},
	}, nil)

	api.On("GetPermissionGroup", ctx, 10).Return([]service.MetabasePermissionGroupMember{
		{ID: 1, Email: "stale@example.com"},
	}, nil)
	accessStorage.On("ListActiveAccessToDataset", ctx, ds1).Return([]*service.Access{}, errors.New("db error"))

	// ds2 should still be processed
	api.On("GetPermissionGroup", ctx, 20).Return([]service.MetabasePermissionGroupMember{
		{ID: 2, Email: "stale@example.com"},
	}, nil)
	accessStorage.On("ListActiveAccessToDataset", ctx, ds2).Return([]*service.Access{}, nil)
	api.On("RemovePermissionGroupMember", ctx, 2).Return(nil)

	runner := metabase_access_remover.New(api, mbStorage, accessStorage)
	err := runner.RunOnce(ctx, log)

	assert.NoError(t, err)
	api.AssertExpectations(t)
	accessStorage.AssertExpectations(t)
}

func TestRunOnce_ContinuesWhenRemoveMemberFails(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	api := new(MockMetabaseAPI)
	mbStorage := new(MockMetabaseStorage)
	accessStorage := new(MockAccessStorage)

	mbStorage.On("GetAllMetadata", ctx).Return([]*service.RestrictedMetabaseMetadata{
		{DatasetID: ds1, PermissionGroupID: intPtr(10), SyncCompleted: nowPtr()},
	}, nil)
	api.On("GetPermissionGroup", ctx, 10).Return([]service.MetabasePermissionGroupMember{
		{ID: 1, Email: "stale1@example.com"},
		{ID: 2, Email: "stale2@example.com"},
	}, nil)
	accessStorage.On("ListActiveAccessToDataset", ctx, ds1).Return([]*service.Access{}, nil)
	api.On("RemovePermissionGroupMember", ctx, 1).Return(errors.New("api error"))
	api.On("RemovePermissionGroupMember", ctx, 2).Return(nil)

	runner := metabase_access_remover.New(api, mbStorage, accessStorage)
	err := runner.RunOnce(ctx, log)

	// RunOnce itself returns no error — per-member failures are only logged
	assert.NoError(t, err)
	api.AssertExpectations(t)
}
