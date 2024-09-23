package syncers_test

import (
	"context"
	"testing"
	"time"

	"github.com/navikt/nada-backend/pkg/syncers"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
)

type MockRunner struct {
	mock.Mock
}

func (m *MockRunner) Name() string {
	return "mockRunner"
}

func (m *MockRunner) RunOnce(ctx context.Context, _ zerolog.Logger) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestSyncerRunsInitialSync(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(zerolog.NewConsoleWriter())
	runner := new(MockRunner)
	runner.On("RunOnce", mock.Anything).Return(nil).Once()

	s := syncers.New(10, runner, logger, syncers.WithRunAtStart())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.Run(ctx)
	time.Sleep(100 * time.Millisecond)

	runner.AssertExpectations(t)
}

func TestSyncerRunsPeriodically(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(zerolog.NewConsoleWriter())
	runner := new(MockRunner)
	runner.On("RunOnce", mock.Anything).Return(nil).Twice()

	s := syncers.New(1, runner, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.Run(ctx)
	time.Sleep(2500 * time.Millisecond)

	runner.AssertExpectations(t)
}

func TestSyncerStopsOnContextDone(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(zerolog.NewConsoleWriter())
	runner := new(MockRunner)
	runner.On("RunOnce", mock.Anything).Return(nil).Once()

	s := syncers.New(1, runner, logger)

	ctx, cancel := context.WithCancel(context.Background())

	go s.Run(ctx)
	time.Sleep(1100 * time.Millisecond)
	cancel()
	time.Sleep(100 * time.Millisecond)

	runner.AssertExpectations(t)
}

func TestSyncerOnlyRunsIfLeader(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(zerolog.NewConsoleWriter())
	runner := new(MockRunner)
	runner.On("RunOnce", mock.Anything).Return(nil).Once()

	s := syncers.New(1, runner, logger, syncers.WithOnlyRunIfLeader())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.Run(ctx)
	time.Sleep(1100 * time.Millisecond)

	runner.AssertExpectations(t)
}
