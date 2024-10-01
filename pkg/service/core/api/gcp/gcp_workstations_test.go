package gcp

import (
	"testing"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/workstations"
	"github.com/stretchr/testify/assert"
)

func TestAddUserToWorkstation(t *testing.T) {
	testCases := []struct {
		name   string
		member string
		input  []*workstations.Binding
		expect []*workstations.Binding
	}{
		{
			name:   "Add user to workstation with no bindings should add user",
			member: "user:nada@nav.no",
			input:  []*workstations.Binding{},
			expect: []*workstations.Binding{
				{
					Role:    service.WorkstationUserRole,
					Members: []string{"user:nada@nav.no"},
				},
			},
		},
		{
			name:   "Add user to workstation with existing user should ignore",
			member: "user:nada@nav.no",
			input: []*workstations.Binding{
				{
					Role:    service.WorkstationUserRole,
					Members: []string{"user:nada@nav.no"},
				},
			},
			expect: []*workstations.Binding{
				{
					Role:    service.WorkstationUserRole,
					Members: []string{"user:nada@nav.no"},
				},
			},
		},
		{
			name:   "Add user to workstation with other role should add user",
			member: "user:nada@nav.no",
			input: []*workstations.Binding{
				{
					Role: service.WorkstationUserRole,
				},
				{
					Role: "role/editor",
					Members: []string{
						"user:somebody@example.com",
					},
				},
			},
			expect: []*workstations.Binding{
				{
					Role:    service.WorkstationUserRole,
					Members: []string{"user:nada@nav.no"},
				},
				{
					Role: "role/editor",
					Members: []string{
						"user:somebody@example.com",
					},
				},
			},
		},
		{
			name:   "Add user to workstation with other user should add user",
			member: "user:nada@nav.no",
			input: []*workstations.Binding{
				{
					Role: service.WorkstationUserRole,
					Members: []string{
						"user:somebody@example.com",
					},
				},
				{
					Role: "role/editor",
					Members: []string{
						"user:somebody@example.com",
					},
				},
			},
			expect: []*workstations.Binding{
				{
					Role: service.WorkstationUserRole,
					Members: []string{
						"user:somebody@example.com",
						"user:nada@nav.no",
					},
				},
				{
					Role: "role/editor",
					Members: []string{
						"user:somebody@example.com",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := addUserToWorkstation(tc.member)(tc.input)
			assert.Equal(t, tc.expect, got)
		})
	}
}
