package workstations_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/navikt/nada-backend/pkg/workstations"
	"github.com/navikt/nada-backend/pkg/workstations/emulator"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkstationConfigOpts_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		opts      workstations.WorkstationConfigOpts
		expectErr bool
		expect    string
	}{
		{
			name: "should work",
			opts: workstations.WorkstationConfigOpts{
				Slug:                "test",
				DisplayName:         "Test",
				MachineType:         workstations.MachineTypeN2DStandard16,
				ServiceAccountEmail: "nada@nav.no",
				CreatedBy:           "datamarkedsplassen",
				SubjectEmail:        "nada@nav.no",
				ContainerImage:      workstations.ContainerImageVSCode,
			},
			expectErr: false,
		},
		{
			name: "should also work",
			opts: workstations.WorkstationConfigOpts{
				Slug:                "tes",
				DisplayName:         "Test",
				MachineType:         workstations.MachineTypeN2DStandard16,
				ServiceAccountEmail: "nada@nav.no",
				CreatedBy:           "datamarkedsplassen",
				SubjectEmail:        "nada@nav.no",
				ContainerImage:      workstations.ContainerImageVSCode,
			},
			expectErr: false,
		},
		{
			name: "starts with -",
			opts: workstations.WorkstationConfigOpts{
				Slug:                "-tt",
				DisplayName:         "Test",
				MachineType:         workstations.MachineTypeN2DStandard16,
				ServiceAccountEmail: "nada@nav.no",
				CreatedBy:           "datamarkedsplassen",
				SubjectEmail:        "nada@nav.no",
				ContainerImage:      workstations.ContainerImageVSCode,
			},
			expectErr: true,
			expect:    "Slug: must be in a valid format.",
		},
		{
			name: "starts with 0",
			opts: workstations.WorkstationConfigOpts{
				Slug:                "0tt",
				DisplayName:         "Test",
				MachineType:         workstations.MachineTypeN2DStandard16,
				ServiceAccountEmail: "nada@nav.no",
				CreatedBy:           "datamarkedsplassen",
				SubjectEmail:        "nada@nav.no",
				ContainerImage:      workstations.ContainerImageVSCode,
			},
			expectErr: true,
			expect:    "Slug: must be in a valid format.",
		},
		{
			name: "will also work",
			opts: workstations.WorkstationConfigOpts{
				Slug:                "t-1",
				DisplayName:         "Test",
				MachineType:         workstations.MachineTypeN2DStandard16,
				ServiceAccountEmail: "nada@nav.no",
				CreatedBy:           "datamarkedsplassen",
				SubjectEmail:        "nada@nav.no",
				ContainerImage:      workstations.ContainerImageVSCode,
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.opts.Validate()
			if tc.expectErr {
				require.Error(t, err)
				require.Equal(t, tc.expect, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWorkstationOperations(t *testing.T) {
	t.Parallel()

	log := zerolog.New(zerolog.NewConsoleWriter())
	ctx := context.Background()
	e := emulator.New(log)
	apiURL := e.Run()
	project := "test"
	configSlug := "nada"
	configDisplayName := "Team nada"
	location := "europe-north1"
	clusterID := "clusterID"

	client := workstations.New(project, location, clusterID, apiURL, true)

	t.Run("Create workstation config", func(t *testing.T) {
		expected := &workstations.WorkstationConfig{
			Name:        configSlug,
			DisplayName: configDisplayName,
		}

		got, err := client.CreateWorkstationConfig(ctx, &workstations.WorkstationConfigOpts{
			Slug:                configSlug,
			DisplayName:         configDisplayName,
			MachineType:         workstations.MachineTypeN2DStandard2,
			ServiceAccountEmail: fmt.Sprintf("%s@%s.iam.gserviceaccount.com", configSlug, project),
			CreatedBy:           "markedsplassen",
			SubjectEmail:        "nada@nav.no",
			ContainerImage:      workstations.ContainerImageVSCode,
		})
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("Create workstation", func(t *testing.T) {
		workstationSlug := "nada-workstation"
		workstationDisplayName := "Team nada workstation"
		expected := &workstations.Workstation{
			Name: workstationSlug,
		}

		got, err := client.CreateWorkstation(ctx, &workstations.WorkstationOpts{
			Slug:        workstationSlug,
			DisplayName: workstationDisplayName,
			Labels:      map[string]string{},
			ConfigName:  configSlug,
		})

		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("Update workstation config", func(t *testing.T) {
		expected := &workstations.WorkstationConfig{
			Name: client.WorkstationParent(configSlug),
		}

		got, err := client.UpdateWorkstationConfig(ctx, &workstations.WorkstationConfigUpdateOpts{
			Slug:           configSlug,
			MachineType:    workstations.MachineTypeN2DStandard2,
			ContainerImage: workstations.ContainerImageVSCode,
		})

		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("Delete workstation config", func(t *testing.T) {
		err := client.DeleteWorkstationConfig(ctx, &workstations.WorkstationConfigDeleteOpts{
			Slug: configSlug,
		})

		require.NoError(t, err)
	})
}
