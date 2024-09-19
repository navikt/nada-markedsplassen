package workstations_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/navikt/nada-backend/pkg/workstations"
	"github.com/navikt/nada-backend/pkg/workstations/emulator"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkstationOperations(t *testing.T) {
	t.Parallel()

	log := zerolog.New(zerolog.NewConsoleWriter())
	ctx := context.Background()
	e := emulator.New(log)
	apiURL := e.Run()

	project := "test"
	location := "europe-north1"
	clusterID := "clusterID"

	configSlug := "nada"
	configDisplayName := "Team nada"
	saEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", configSlug, project)
	client := workstations.New(project, location, clusterID, apiURL, true)

	idleTimeout := workstations.DefaultIdleTimeoutInSec * time.Second
	runningTimeout := workstations.DefaultRunningTimeoutInSec * time.Second

	workstationConfig := &workstations.WorkstationConfig{
		Slug:               configSlug,
		FullyQualifiedName: fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", project, location, clusterID, configSlug),
		DisplayName:        configDisplayName,
		MachineType:        workstations.MachineTypeN2DStandard2,
		ServiceAccount:     saEmail,
		Image:              workstations.ContainerImageVSCode,
		IdleTimeout:        &idleTimeout,
		RunningTimeout:     &runningTimeout,
		Env: map[string]string{
			"WORKSTATION_NAME": "nada",
		},
	}

	workstationSlug := "nada-workstation"
	workstationDisplayName := "Team nada workstation"

	workstation := &workstations.Workstation{
		Slug:               workstationSlug,
		FullyQualifiedName: fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s/workstations/%s", project, location, clusterID, configSlug, workstationSlug),
		DisplayName:        workstationDisplayName,
		State:              workstations.Workstation_STATE_STARTING,
		Host:               "https://127.0.0.1",
	}

	t.Run("Get workstation config that does not exist", func(t *testing.T) {
		_, err := client.GetWorkstationConfig(ctx, &workstations.WorkstationConfigGetOpts{
			Slug: configSlug,
		})

		require.ErrorIs(t, err, workstations.ErrNotExist)
	})

	t.Run("Create workstation config", func(t *testing.T) {
		got, err := client.CreateWorkstationConfig(ctx, &workstations.WorkstationConfigOpts{
			Slug:                configSlug,
			DisplayName:         configDisplayName,
			MachineType:         workstations.MachineTypeN2DStandard2,
			ServiceAccountEmail: saEmail,
			SubjectEmail:        "nada@nav.no",
			ContainerImage:      workstations.ContainerImageVSCode,
		})

		require.NoError(t, err)
		diff := cmp.Diff(workstationConfig, got, cmpopts.IgnoreFields(workstations.WorkstationConfig{}, "CreateTime"))
		assert.Empty(t, diff)
		assert.NotNil(t, got.CreateTime)
	})

	t.Run("Get workstation config", func(t *testing.T) {
		got, err := client.GetWorkstationConfig(ctx, &workstations.WorkstationConfigGetOpts{
			Slug: configSlug,
		})

		require.NoError(t, err)
		diff := cmp.Diff(workstationConfig, got, cmpopts.IgnoreFields(workstations.WorkstationConfig{}, "CreateTime"))
		assert.Empty(t, diff)
	})

	t.Run("Update workstation config", func(t *testing.T) {
		workstationConfig.MachineType = workstations.MachineTypeN2DStandard32
		workstationConfig.Image = workstations.ContainerImagePosit

		got, err := client.UpdateWorkstationConfig(ctx, &workstations.WorkstationConfigUpdateOpts{
			Slug:           configSlug,
			MachineType:    workstations.MachineTypeN2DStandard32,
			ContainerImage: workstations.ContainerImagePosit,
		})

		require.NoError(t, err)
		diff := cmp.Diff(workstationConfig, got, cmpopts.IgnoreFields(workstations.WorkstationConfig{}, "CreateTime", "UpdateTime"))
		assert.Empty(t, diff)
		assert.NotNil(t, got.UpdateTime)
	})

	t.Run("Get updated workstation config", func(t *testing.T) {
		got, err := client.GetWorkstationConfig(ctx, &workstations.WorkstationConfigGetOpts{
			Slug: configSlug,
		})
		require.NoError(t, err)
		diff := cmp.Diff(workstationConfig, got, cmpopts.IgnoreFields(workstations.WorkstationConfig{}, "CreateTime", "UpdateTime"))
		assert.Empty(t, diff)
		assert.NotNil(t, got.UpdateTime)
	})

	t.Run("Update workstation config that does not exist", func(t *testing.T) {
		_, err := client.UpdateWorkstationConfig(ctx, &workstations.WorkstationConfigUpdateOpts{
			Slug:           "non-existent",
			MachineType:    workstations.MachineTypeN2DStandard32,
			ContainerImage: workstations.ContainerImagePosit,
		})

		require.ErrorIs(t, err, workstations.ErrNotExist)
	})

	t.Run("Create workstation", func(t *testing.T) {
		got, err := client.CreateWorkstation(ctx, &workstations.WorkstationOpts{
			Slug:                  workstationSlug,
			DisplayName:           workstationDisplayName,
			Labels:                map[string]string{},
			WorkstationConfigSlug: configSlug,
		})

		require.NoError(t, err)
		diff := cmp.Diff(workstation, got, cmpopts.IgnoreFields(workstations.Workstation{}, "CreateTime"))
		assert.Empty(t, diff)
		assert.NotNil(t, got.CreateTime)
	})

	t.Run("Delete workstation config", func(t *testing.T) {
		err := client.DeleteWorkstationConfig(ctx, &workstations.WorkstationConfigDeleteOpts{
			Slug: configSlug,
		})

		require.NoError(t, err)

		assert.Empty(t, e.GetWorkstationConfigs())
		assert.Empty(t, e.GetWorkstations())
	})
}
