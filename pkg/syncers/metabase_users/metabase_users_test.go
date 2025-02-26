package metabase_users_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/syncers/metabase_users"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMetabaseAPI struct {
	mock.Mock
	service.MetabaseAPI
}

func (m *MockMetabaseAPI) GetUsers(ctx context.Context) ([]service.MetabaseUser, error) {
	args := m.Called(ctx)
	return args.Get(0).([]service.MetabaseUser), args.Error(1)
}

func (m *MockMetabaseAPI) DeleteUser(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestMetabaseDeactivateUsersWithNoActivity(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("returns error when GetUsers fails", func(t *testing.T) {
		api := new(MockMetabaseAPI)
		api.On("GetUsers", ctx).Return([]service.MetabaseUser{}, errors.New("failed to get users"))

		runner := metabase_users.New(api)
		err := runner.RunOnce(ctx, log)

		assert.Error(t, err)
		assert.Equal(t, "getting users: failed to get users", err.Error())
		api.AssertExpectations(t)
	})

	t.Run("logs users with no activity", func(t *testing.T) {
		api := new(MockMetabaseAPI)
		users := []service.MetabaseUser{
			{Email: "user1@example.com", LastLogin: nil, ID: 1},
		}
		api.On("GetUsers", ctx).Return(users, nil)

		runner := metabase_users.New(api)
		err := runner.RunOnce(ctx, log)

		assert.NoError(t, err)
		api.AssertExpectations(t)
	})

	t.Run("logs users with no activity for over a year", func(t *testing.T) {
		api := new(MockMetabaseAPI)
		lastYear := time.Now().AddDate(-1, -1, 0)
		users := []service.MetabaseUser{
			{Email: "user2@example.com", LastLogin: &lastYear},
		}
		api.On("GetUsers", ctx).Return(users, nil)
		api.On("DeleteUser", ctx, mock.Anything).Return(nil)

		runner := metabase_users.New(api)
		err := runner.RunOnce(ctx, log)

		assert.NoError(t, err)
		api.AssertExpectations(t)
	})

	t.Run("does not log users with recent activity", func(t *testing.T) {
		api := new(MockMetabaseAPI)
		recent := time.Now().AddDate(0, -1, 0)
		users := []service.MetabaseUser{
			{Email: "user3@example.com", LastLogin: &recent},
		}
		api.On("GetUsers", ctx).Return(users, nil)

		runner := metabase_users.New(api)
		err := runner.RunOnce(ctx, log)

		assert.NoError(t, err)
		api.AssertExpectations(t)
	})
}
