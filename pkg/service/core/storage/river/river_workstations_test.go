package river_test

import (
	"context"
	"testing"

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
