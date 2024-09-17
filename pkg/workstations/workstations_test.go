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

func TestCreateWorkstationConfig(t *testing.T) {
	t.Parallel()

	log := zerolog.New(zerolog.NewConsoleWriter())

	t.Run("Create workstation", func(t *testing.T) {
		ctx := context.Background()
		e := emulator.New(log)
		apiURL := e.Run()

		project := "test"
		slug := "nada"
		displayName := "Team nada"
		client := workstations.New(project, "europe-north1", "clusterID", apiURL, true)

		expected := &workstations.WorkstationConfig{
			Name:        slug,
			DisplayName: displayName,
		}

		got, err := client.CreateWorkstationConfig(ctx, &workstations.WorkstationConfigOpts{
			Slug:                slug,
			DisplayName:         displayName,
			MachineType:         workstations.MachineTypeN2DStandard2,
			ServiceAccountEmail: fmt.Sprintf("%s@%s.iam.gserviceaccount.com", slug, project),
			CreatedBy:           "markedsplassen",
			SubjectEmail:        "nada@nav.no",
			ContainerImage:      workstations.ContainerImageVSCode,
		})
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})
}
