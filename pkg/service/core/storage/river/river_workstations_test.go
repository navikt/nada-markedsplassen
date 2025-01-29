package river_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service"
	riverstore "github.com/navikt/nada-backend/pkg/service/core/storage/river"
	"github.com/navikt/nada-backend/pkg/worker"
	"github.com/navikt/nada-backend/pkg/worker/worker_args"
	"github.com/navikt/nada-backend/test/integration"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertest"
)

type workstationServiceMock struct {
	service.WorkstationsService
}

func TestWorkstationZonalTagBindingJob(t *testing.T) {
	log := zerolog.New(zerolog.NewConsoleWriter()).Level(zerolog.InfoLevel)
	ctx := context.Background()

	c := integration.NewContainers(t, log)
	defer c.Cleanup()

	pgCfg := c.RunPostgres(integration.NewPostgresConfig())

	repo, err := database.New(
		pgCfg.ConnectionURL(),
		10,
		10,
	)
	require.NoError(t, err)

	workers := river.NewWorkers()
	config := worker.WorkstationConfig(&log, workers)
	config.TestOnly = true

	_, err = worker.NewWorkstationWorker(config, &workstationServiceMock{}, repo)
	require.NoError(t, err)

	store := riverstore.NewWorkstationsStorage(config, repo)
	_, err = store.CreateWorkstationZonalTagBindingJob(ctx, &service.WorkstationZonalTagBindingJobOpts{
		Ident:             "test",
		Action:            service.WorkstationZonalTagBindingJobActionAdd,
		Zone:              "europe-north1-a",
		Parent:            "//my/vm",
		TagValue:          "",
		TagNamespacedName: "something/something/something",
	})
	require.NoError(t, err)

	_ = rivertest.RequireInserted(ctx, t, riverdatabasesql.New(repo.GetDB()), &worker_args.WorkstationZonalTagBindingJob{
		Ident:             "test",
		Action:            service.WorkstationZonalTagBindingJobActionAdd,
		Zone:              "europe-north1-a",
		Parent:            "//my/vm",
		TagValue:          "",
		TagNamespacedName: "something/something/something",
	}, nil)

	_, err = store.CreateWorkstationZonalTagBindingJob(ctx, &service.WorkstationZonalTagBindingJobOpts{
		Ident:             "test",
		Action:            service.WorkstationZonalTagBindingJobActionAdd,
		Zone:              "europe-north1-a",
		Parent:            "//my/vm",
		TagValue:          "",
		TagNamespacedName: "something/something/something",
	})
	require.NoError(t, err)
	require.NoError(t, err)

	jobs, err := store.GetWorkstationZonalTagBindingJobsForUser(ctx, "test")
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "test", jobs[0].Ident)
}

func TestWorkstationStartJob(t *testing.T) {
	log := zerolog.New(zerolog.NewConsoleWriter()).Level(zerolog.InfoLevel)
	ctx := context.Background()

	c := integration.NewContainers(t, log)
	defer c.Cleanup()

	pgCfg := c.RunPostgres(integration.NewPostgresConfig())

	repo, err := database.New(
		pgCfg.ConnectionURL(),
		10,
		10,
	)
	require.NoError(t, err)

	workers := river.NewWorkers()
	config := worker.WorkstationConfig(&log, workers)
	config.TestOnly = true

	_, err = worker.NewWorkstationWorker(config, &workstationServiceMock{}, repo)
	require.NoError(t, err)

	store := riverstore.NewWorkstationsStorage(config, repo)
	_, err = store.CreateWorkstationStartJob(ctx, "test")
	require.NoError(t, err)

	_ = rivertest.RequireInserted(ctx, t, riverdatabasesql.New(repo.GetDB()), &worker_args.WorkstationStart{Ident: "test"}, nil)

	_, err = store.CreateWorkstationStartJob(ctx, "test")
	require.NoError(t, err)

	jobs, err := store.GetWorkstationStartJobsForUser(ctx, "test")
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "test", jobs[0].Ident)
}

func TestWorkstationsStorage(t *testing.T) {
	log := zerolog.New(zerolog.NewConsoleWriter()).Level(zerolog.InfoLevel)
	ctx := context.Background()

	c := integration.NewContainers(t, log)
	defer c.Cleanup()

	pgCfg := c.RunPostgres(integration.NewPostgresConfig())

	repo, err := database.New(
		pgCfg.ConnectionURL(),
		10,
		10,
	)
	require.NoError(t, err)

	workers := river.NewWorkers()
	config := worker.WorkstationConfig(&log, workers)
	config.TestOnly = true

	_, err = worker.NewWorkstationWorker(config, &workstationServiceMock{}, repo)
	require.NoError(t, err)

	store := riverstore.NewWorkstationsStorage(config, repo)
	job, err := store.CreateWorkstationJob(ctx, &service.WorkstationJobOpts{
		User: &service.User{
			Ident: "test",
		},
		Input: &service.WorkstationInput{},
	})
	require.NoError(t, err)

	_ = rivertest.RequireInserted(ctx, t, riverdatabasesql.New(repo.GetDB()), &worker_args.WorkstationJob{Ident: "test"}, nil)

	_, err = store.CreateWorkstationJob(ctx, &service.WorkstationJobOpts{
		User: &service.User{
			Ident: "test-two",
		},
		Input: &service.WorkstationInput{},
	})
	require.NoError(t, err)

	jobs, err := store.GetWorkstationJobsForUser(ctx, "test")
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "test", jobs[0].Ident)

	job, err = store.GetWorkstationJob(ctx, job.ID)
	assert.NoError(t, err)
	assert.Equal(t, "test", job.Ident)
}

func TestJobDifference(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		a      *service.WorkstationJob
		b      *service.WorkstationJob
		expect map[string]*service.Diff
	}{
		{
			name: "no difference",
			a: &service.WorkstationJob{
				ID:                        1,
				MachineType:               "a",
				ContainerImage:            "b",
				URLAllowList:              []string{"c"},
				OnPremAllowList:           []string{"d"},
				DisableGlobalURLAllowList: false,
			},
			b: &service.WorkstationJob{
				ID:                        1,
				MachineType:               "a",
				ContainerImage:            "b",
				URLAllowList:              []string{"c"},
				OnPremAllowList:           []string{"d"},
				DisableGlobalURLAllowList: false,
			},
			expect: map[string]*service.Diff{},
		},
		{
			name: "lots of differences",
			a: &service.WorkstationJob{
				ID:                        1,
				MachineType:               "a",
				ContainerImage:            "b",
				URLAllowList:              []string{"c"},
				OnPremAllowList:           []string{"d"},
				DisableGlobalURLAllowList: false,
			},
			b: &service.WorkstationJob{
				ID:                        2,
				MachineType:               "c",
				ContainerImage:            "b",
				URLAllowList:              []string{"c", "b"},
				OnPremAllowList:           []string{"b"},
				DisableGlobalURLAllowList: true,
			},
			expect: map[string]*service.Diff{
				service.WorkstationDiffMachineType: {
					Added:   []string{"c"},
					Removed: []string{"a"},
				},
				service.WorkstationDiffURLAllowList: {
					Added: []string{"b"},
				},
				service.WorkstationDiffOnPremAllowList: {
					Added:   []string{"b"},
					Removed: []string{"d"},
				},
				service.WorkstationDiffDisableGlobalURLAllowList: {
					Added:   []string{"true"},
					Removed: []string{"false"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			diff := riverstore.JobDifference(tc.a, tc.b)
			changes := cmp.Diff(tc.expect, diff)
			assert.Empty(t, changes)
		})
	}
}
