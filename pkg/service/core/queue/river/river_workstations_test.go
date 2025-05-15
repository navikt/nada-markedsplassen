package river_test

import (
	"context"
	riverstore "github.com/navikt/nada-backend/pkg/service/core/queue/river"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"

	"github.com/google/go-cmp/cmp"

	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/worker"
	"github.com/navikt/nada-backend/pkg/worker/worker_args"
	"github.com/navikt/nada-backend/test/integration"
	"github.com/neilotoole/slogt"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riverqueue/river/rivertest"
)

type workstationServiceMock struct {
	service.WorkstationsService
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
	config := worker.RiverConfig(&log, workers)
	config.TestOnly = true

	err = worker.WorkstationAddWorkers(config, &workstationServiceMock{}, repo)
	require.NoError(t, err)

	store := riverstore.NewWorkstationsQueue(config, repo)
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

func TestWorkstationsQueue(t *testing.T) {
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
	config := worker.RiverConfig(&log, workers)
	config.TestOnly = true

	err = worker.WorkstationAddWorkers(config, &workstationServiceMock{}, repo)
	require.NoError(t, err)

	store := riverstore.NewWorkstationsQueue(config, repo)
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

func TestWorkstationsQueue_CreateWorkstationConnectivityWorkflow(t *testing.T) {
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
	config := worker.RiverConfig(&log, workers)
	config.TestOnly = true
	config.Logger = slogt.New(t, slogt.JSON())

	err = worker.WorkstationAddWorkers(config, &workstationServiceMock{}, repo)
	require.NoError(t, err)

	store := riverstore.NewWorkstationsQueue(config, repo)

	err = store.CreateWorkstationConnectivityWorkflow(ctx, "x123456", "abc123", []string{"host1", "host2"})
	require.NoError(t, err)

	_ = rivertest.RequireManyInserted(ctx, t, riverdatabasesql.New(repo.GetDB()), []rivertest.ExpectedJob{
		{
			Args: &worker_args.WorkstationConnectJob{
				Ident: "x123456",
				Host:  "host1",
			},
		},
		{
			Args: &worker_args.WorkstationConnectJob{
				Ident: "x123456",
				Host:  "host2",
			},
		},
	})

	_ = rivertest.RequireInserted(ctx, t, riverdatabasesql.New(repo.GetDB()), &worker_args.WorkstationDisconnectJob{
		Ident: "x123456",
		Hosts: []string{"host1", "host2"},
	}, nil)

	_ = rivertest.RequireInserted(ctx, t, riverdatabasesql.New(repo.GetDB()), &worker_args.WorkstationNotifyJob{
		Ident: "x123456",
		Hosts: []string{"host1", "host2"},
	}, nil)

	err = store.CreateWorkstationConnectivityWorkflow(ctx, "x123456", "abc123", []string{"host1", "host2"})
	require.NoError(t, err)
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
				JobHeader: service.JobHeader{
					ID: 1,
				},
				MachineType:    "a",
				ContainerImage: "b",
			},
			b: &service.WorkstationJob{
				JobHeader: service.JobHeader{
					ID: 1,
				},
				MachineType:    "a",
				ContainerImage: "b",
			},
			expect: map[string]*service.Diff{},
		},
		{
			name: "lots of differences",
			a: &service.WorkstationJob{
				JobHeader: service.JobHeader{
					ID: 1,
				},
				MachineType:    "a",
				ContainerImage: "b",
			},
			b: &service.WorkstationJob{
				JobHeader: service.JobHeader{
					ID: 1,
				},
				MachineType:    "c",
				ContainerImage: "b",
			},
			expect: map[string]*service.Diff{
				service.WorkstationDiffMachineType: {
					Added:   []string{"c"},
					Removed: []string{"a"},
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
