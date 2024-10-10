package computeengine_test

import (
	"cloud.google.com/go/compute/apiv1/computepb"
	"context"
	"github.com/navikt/nada-backend/pkg/computeengine"
	"github.com/navikt/nada-backend/pkg/computeengine/emulator"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"testing"
)

func strPtr(s string) *string {
	return &s
}

func TestNewInstancesClient(t *testing.T) {
	testCases := []struct {
		name      string
		project   string
		zones     []string
		filter    string
		instances map[string][]*computepb.Instance
		expect    []*computeengine.VirtualMachine
	}{
		{
			name:      "no instances",
			project:   "test",
			zones:     []string{"europe-west1-b", "europe-west1-c"},
			filter:    "",
			instances: map[string][]*computepb.Instance{},
			expect:    nil,
		},
		{
			name:    "one instance, in one zone",
			project: "test",
			zones:   []string{"europe-west1-b"},
			filter:  "",
			instances: map[string][]*computepb.Instance{
				"europe-west1-b": {
					{
						Name: strPtr("test-instance"),
					},
				},
			},
			expect: []*computeengine.VirtualMachine{
				{
					Name: "test-instance",
				},
			},
		},
		{
			name:    "instance in multiple zones",
			project: "test",
			zones:   []string{"europe-west1-b", "europe-west1-c"},
			filter:  "",
			instances: map[string][]*computepb.Instance{
				"europe-west1-b": {
					{
						Name: strPtr("test-instance-1"),
					},
				},
				"europe-west1-c": {
					{
						Name: strPtr("test-instance-2"),
					},
				},
			},
			expect: []*computeengine.VirtualMachine{
				{
					Name: "test-instance-1",
				},
				{
					Name: "test-instance-2",
				},
			},
		},
		{
			name:    "one instance from multiple zones",
			project: "test",
			zones:   []string{"europe-west1-b"},
			filter:  "",
			instances: map[string][]*computepb.Instance{
				"europe-west1-b": {
					{
						Name: strPtr("test-instance-1"),
					},
				},
				"europe-west1-c": {
					{
						Name: strPtr("test-instance-2"),
					},
				},
			},
			expect: []*computeengine.VirtualMachine{
				{
					Name: "test-instance-1",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.New(zerolog.NewConsoleWriter())
			e := emulator.New(log)
			e.SetInstances(tc.instances)
			url := e.Run()

			c := computeengine.NewClient(url, true)

			got, err := c.ListVirtualMachines(context.Background(), tc.project, tc.zones, tc.filter)
			require.NoError(t, err)
			require.Equal(t, tc.expect, got)
		})
	}
}
func TestNewFirewallPoliciesClient(t *testing.T) {
	testCases := []struct {
		name               string
		firewallPolicyName string
		firewallPolicies   map[string][]*computepb.FirewallPolicy
		expect             []computeengine.FirewallRule
	}{
		{
			name:               "no firewall rules",
			firewallPolicyName: "finnes ikke",
			firewallPolicies:   map[string][]*computepb.FirewallPolicy{},
			expect:             nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.New(zerolog.NewConsoleWriter())
			e := emulator.New(log)
			e.SetFirewallPolicies(tc.firewallPolicies)
			url := e.Run()

			c := computeengine.NewClient(url, true)

			got, err := c.GetFirewallRulesForPolicy(context.Background(), tc.firewallPolicyName)
			require.NoError(t, err)
			require.Equal(t, tc.expect, got)
		})
	}
}
