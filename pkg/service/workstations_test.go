package service_test

import (
	"testing"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/stretchr/testify/require"
)

func TestWorkstationConfigOpts_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		opts      service.WorkstationConfigOpts
		expectErr bool
		expect    string
	}{
		{
			name: "should work",
			opts: service.WorkstationConfigOpts{
				Slug:                "test",
				DisplayName:         "Test",
				MachineType:         service.MachineTypeN2DStandard16,
				ServiceAccountEmail: "nada@nav.no",
				SubjectEmail:        "nada@nav.no",
				ContainerImage:      service.ContainerImageVSCode,
			},
			expectErr: false,
		},
		{
			name: "should also work",
			opts: service.WorkstationConfigOpts{
				Slug:                "tes",
				DisplayName:         "Test",
				MachineType:         service.MachineTypeN2DStandard16,
				ServiceAccountEmail: "nada@nav.no",
				SubjectEmail:        "nada@nav.no",
				ContainerImage:      service.ContainerImageVSCode,
			},
			expectErr: false,
		},
		{
			name: "starts with -",
			opts: service.WorkstationConfigOpts{
				Slug:                "-tt",
				DisplayName:         "Test",
				MachineType:         service.MachineTypeN2DStandard16,
				ServiceAccountEmail: "nada@nav.no",
				SubjectEmail:        "nada@nav.no",
				ContainerImage:      service.ContainerImageVSCode,
			},
			expectErr: true,
			expect:    "Slug: must be in a valid format.",
		},
		{
			name: "starts with 0",
			opts: service.WorkstationConfigOpts{
				Slug:                "0tt",
				DisplayName:         "Test",
				MachineType:         service.MachineTypeN2DStandard16,
				ServiceAccountEmail: "nada@nav.no",
				SubjectEmail:        "nada@nav.no",
				ContainerImage:      service.ContainerImageVSCode,
			},
			expectErr: true,
			expect:    "Slug: must be in a valid format.",
		},
		{
			name: "will also work",
			opts: service.WorkstationConfigOpts{
				Slug:                "t-1",
				DisplayName:         "Test",
				MachineType:         service.MachineTypeN2DStandard16,
				ServiceAccountEmail: "nada@nav.no",
				SubjectEmail:        "nada@nav.no",
				ContainerImage:      service.ContainerImageVSCode,
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
