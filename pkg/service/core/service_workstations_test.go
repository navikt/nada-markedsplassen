package core_test

import (
	"errors"
	"testing"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core"
	"github.com/stretchr/testify/require"
)

func TestPortInRange(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		port   int
		ranges []service.PortRange
		expect bool
	}{
		{
			name: "port in range",
			port: 8080,
			ranges: []service.PortRange{
				{First: 8000, Last: 9000},
			},
			expect: true,
		},
		{
			name: "port out of range",
			port: 7000,
			ranges: []service.PortRange{
				{First: 8000, Last: 9000},
			},
			expect: false,
		},
		{
			name: "port on lower boundary",
			port: 8000,
			ranges: []service.PortRange{
				{First: 8000, Last: 9000},
			},
			expect: true,
		},
		{
			name: "port on upper boundary",
			port: 9000,
			ranges: []service.PortRange{
				{First: 8000, Last: 9000},
			},
			expect: true,
		},
		{
			name: "multiple ranges, port in second range",
			port: 8500,
			ranges: []service.PortRange{
				{First: 8000, Last: 8100},
				{First: 8400, Last: 8600},
			},
			expect: true,
		},
		{
			name: "multiple ranges, port out of all ranges",
			port: 8200,
			ranges: []service.PortRange{
				{First: 8000, Last: 8100},
				{First: 8400, Last: 8600},
			},
			expect: false,
		},
		{
			name: "empty ranges",
			port: 8080,
			ranges: []service.PortRange{
				// No ranges defined
			},
			expect: false,
		},
		{
			name: "exact match with single port range",
			port: 8080,
			ranges: []service.PortRange{
				{First: 8080, Last: 8080},
			},
			expect: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := core.IsPortInRange(tc.port, tc.ranges)
			require.Equal(t, tc.expect, result)
		})
	}
}

func TestIsConnectivityAllowed(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		allowedPorts []service.PortRange
		hosts        []string
		tnsNames     []service.TNSName
		expectErr    bool
		expect       error
	}{
		{
			name: "SSH is not enabled, so connectivity should be allowed",
			allowedPorts: []service.PortRange{
				{
					First: 80,
					Last:  80,
				},
			},
			hosts: []string{
				"dvh-p.adeo.no",
			},
			tnsNames: []service.TNSName{
				{
					TnsName:     "dvh-p",
					Name:        "DVH-P",
					Host:        "dvh-p.adeo.no",
					Description: "Datavarehus P",
				},
			},
			expectErr: false,
		},
		{
			name: "SSH is enabled, but host is not in TNS names, so connectivity should be allowed",
			allowedPorts: []service.PortRange{
				{
					First: 22,
					Last:  22,
				},
			},
			hosts: []string{
				"unknown-host.adeo.no",
			},
			tnsNames: []service.TNSName{
				{
					TnsName:     "dvh-p",
					Name:        "DVH-P",
					Host:        "dvh-p.adeo.no",
					Description: "Datavarehus P",
				},
			},
			expectErr: false,
		},
		{
			name: "SSH is enabled and host is in TNS names, so connectivity should be denied",
			allowedPorts: []service.PortRange{
				{
					First: 22,
					Last:  22,
				},
			},
			hosts: []string{
				"dvh-p.adeo.no",
			},
			tnsNames: []service.TNSName{
				{
					TnsName:     "dvh-p",
					Host:        "dvh-p.adeo.no",
					Name:        "DVH-P",
					Description: "Datavarehus P",
				},
			},
			expectErr: true,
			expect:    errors.New("connectivity to 'dvh-p.adeo.no' is not allowed while ssh is enabled"),
		},
		{
			name: "SSH is enabled, but host is DVH-I, so connectivity should be allowed",
			allowedPorts: []service.PortRange{
				{
					First: 22,
					Last:  22,
				},
			},
			hosts: []string{
				"dvh-i.adeo.no",
			},
			tnsNames: []service.TNSName{
				{
					TnsName:     "dvh-p",
					Name:        "DVH-P",
					Host:        "dvh-p.adeo.no",
					Description: "Datavarehus P",
				},
				{
					TnsName:     service.TNSNameDVHI,
					Name:        "DVH-I",
					Host:        "dvh-i.adeo.no",
					Description: "Datavarehus I",
				},
			},
			expectErr: false,
			expect:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := core.IsConnectivityAllowed(tc.allowedPorts, tc.hosts, tc.tnsNames)
			if !tc.expectErr {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			require.Equal(t, tc.expect.Error(), err.Error())
		})
	}
}
