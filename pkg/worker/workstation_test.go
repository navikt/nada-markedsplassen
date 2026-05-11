package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/worker/worker_args"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type workstationURLListServiceMock struct {
	service.WorkstationsService
	getActiveURLList func(ctx context.Context, user *service.User) (*service.WorkstationActiveURLListForIdent, error)
	ensureURLList    func(ctx context.Context, input *service.WorkstationActiveURLListForIdent) error
}

func (m *workstationURLListServiceMock) GetWorkstationActiveURLListForIdent(ctx context.Context, user *service.User) (*service.WorkstationActiveURLListForIdent, error) {
	return m.getActiveURLList(ctx, user)
}

func (m *workstationURLListServiceMock) EnsureWorkstationURLList(ctx context.Context, input *service.WorkstationActiveURLListForIdent) error {
	return m.ensureURLList(ctx, input)
}

func TestWorkstationEnsureURLListForIdent_Work_HappyPath(t *testing.T) {
	ctx := context.Background()
	ident := "nav-ident-1"

	expectedURLList := &service.WorkstationActiveURLListForIdent{
		Slug:                 ident,
		URLList:              []string{"https://example.com", "https://other.com"},
		DisableGlobalURLList: true,
	}

	var capturedInput *service.WorkstationActiveURLListForIdent

	mock := &workstationURLListServiceMock{
		getActiveURLList: func(ctx context.Context, user *service.User) (*service.WorkstationActiveURLListForIdent, error) {
			assert.Equal(t, ident, user.Ident)
			return expectedURLList, nil
		},
		ensureURLList: func(ctx context.Context, input *service.WorkstationActiveURLListForIdent) error {
			capturedInput = input
			return nil
		},
	}

	w := &WorkstationEnsureURLListForIdent{
		service: mock,
	}

	job := &river.Job[worker_args.WorkstationEnsureURLListForIdent]{
		Args: worker_args.WorkstationEnsureURLListForIdent{Ident: ident},
	}

	err := w.Work(ctx, job)
	require.NoError(t, err)

	require.NotNil(t, capturedInput)
	assert.Equal(t, ident, capturedInput.Slug)
	assert.Equal(t, expectedURLList.URLList, capturedInput.URLList)
	assert.Equal(t, expectedURLList.DisableGlobalURLList, capturedInput.DisableGlobalURLList)
}

func TestWorkstationEnsureURLListForIdent_Work_GetURLListError(t *testing.T) {
	ctx := context.Background()
	ident := "nav-ident-2"
	expectedErr := errors.New("service unavailable")

	mock := &workstationURLListServiceMock{
		getActiveURLList: func(ctx context.Context, user *service.User) (*service.WorkstationActiveURLListForIdent, error) {
			return nil, expectedErr
		},
		ensureURLList: func(ctx context.Context, input *service.WorkstationActiveURLListForIdent) error {
			t.Fatal("EnsureWorkstationURLList should not be called when GetWorkstationActiveURLListForIdent fails")
			return nil
		},
	}

	w := &WorkstationEnsureURLListForIdent{
		service: mock,
	}

	job := &river.Job[worker_args.WorkstationEnsureURLListForIdent]{
		Args: worker_args.WorkstationEnsureURLListForIdent{Ident: ident},
	}

	err := w.Work(ctx, job)
	require.Error(t, err)
	assert.ErrorContains(t, err, ident)
	assert.ErrorContains(t, err, "getting workstation active urllist")
}
