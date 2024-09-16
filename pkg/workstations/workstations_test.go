package workstations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkstationConfigOpts_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		opts      WorkstationConfigOpts
		expectErr bool
		expect    string
	}{
		{
			name: "should work",
			opts: WorkstationConfigOpts{
				Slug:                "test",
				DisplayName:         "Test",
				MachineType:         MachineTypeN2DStandard16,
				ServiceAccountEmail: "nada@nav.no",
				CreatedBy:           "datamarkedsplassen",
				SubjectEmail:        "nada@nav.no",
				ContainerImage:      ContainerImageVSCode,
			},
			expectErr: false,
		},
		{
			name: "should also work",
			opts: WorkstationConfigOpts{
				Slug:                "tes",
				DisplayName:         "Test",
				MachineType:         MachineTypeN2DStandard16,
				ServiceAccountEmail: "nada@nav.no",
				CreatedBy:           "datamarkedsplassen",
				SubjectEmail:        "nada@nav.no",
				ContainerImage:      ContainerImageVSCode,
			},
			expectErr: false,
		},
		{
			name: "starts with -",
			opts: WorkstationConfigOpts{
				Slug:                "-tt",
				DisplayName:         "Test",
				MachineType:         MachineTypeN2DStandard16,
				ServiceAccountEmail: "nada@nav.no",
				CreatedBy:           "datamarkedsplassen",
				SubjectEmail:        "nada@nav.no",
				ContainerImage:      ContainerImageVSCode,
			},
			expectErr: true,
			expect:    "Slug: must be in a valid format.",
		},
		{
			name: "starts with 0",
			opts: WorkstationConfigOpts{
				Slug:                "0tt",
				DisplayName:         "Test",
				MachineType:         MachineTypeN2DStandard16,
				ServiceAccountEmail: "nada@nav.no",
				CreatedBy:           "datamarkedsplassen",
				SubjectEmail:        "nada@nav.no",
				ContainerImage:      ContainerImageVSCode,
			},
			expectErr: true,
			expect:    "Slug: must be in a valid format.",
		},
		{
			name: "will also work",
			opts: WorkstationConfigOpts{
				Slug:                "t-1",
				DisplayName:         "Test",
				MachineType:         MachineTypeN2DStandard16,
				ServiceAccountEmail: "nada@nav.no",
				CreatedBy:           "datamarkedsplassen",
				SubjectEmail:        "nada@nav.no",
				ContainerImage:      ContainerImageVSCode,
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
