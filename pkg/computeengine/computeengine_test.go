package computeengine_test

import (
	"context"
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"

	"github.com/navikt/nada-backend/pkg/computeengine"
	"github.com/navikt/nada-backend/pkg/computeengine/emulator"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string {
	return &s
}

func uint64Ptr(i uint64) *uint64 {
	return &i
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
			zones:     []string{"europe-north1-b", "europe-north1-c"},
			filter:    "",
			instances: map[string][]*computepb.Instance{},
			expect:    nil,
		},
		{
			name:    "one instance, in one zone",
			project: "test",
			zones:   []string{"europe-north1-b"},
			filter:  "",
			instances: map[string][]*computepb.Instance{
				"europe-north1-b": {
					{
						Name: strPtr("test-instance"),
						Id:   uint64Ptr(123),
						NetworkInterfaces: []*computepb.NetworkInterface{
							{
								NetworkIP: strPtr("1.2.3.4"),
							},
						},
						Zone: strPtr("https://www.googleapis.com/compute/v1/projects/knada-dev/zones/europe-north1-b"),
					},
				},
			},
			expect: []*computeengine.VirtualMachine{
				{
					Name:               "test-instance",
					ID:                 123,
					IPs:                []string{"1.2.3.4"},
					FullyQualifiedZone: "https://www.googleapis.com/compute/v1/projects/knada-dev/zones/europe-north1-b",
					Zone:               "europe-north1-b",
				},
			},
		},
		{
			name:    "instance in multiple zones",
			project: "test",
			zones:   []string{"europe-north1-b", "europe-north1-c"},
			filter:  "",
			instances: map[string][]*computepb.Instance{
				"europe-north1-b": {
					{
						Name: strPtr("test-instance-1"),
						Id:   uint64Ptr(123),
						NetworkInterfaces: []*computepb.NetworkInterface{
							{
								NetworkIP: strPtr("1.2.3.4"),
							},
						},
						Zone: strPtr("https://www.googleapis.com/compute/v1/projects/knada-dev/zones/europe-north1-b"),
					},
				},
				"europe-north1-c": {
					{
						Name: strPtr("test-instance-2"),
						Id:   uint64Ptr(1231234),
						NetworkInterfaces: []*computepb.NetworkInterface{
							{
								NetworkIP: strPtr("5.6.7.8"),
							},
						},
						Zone: strPtr("https://www.googleapis.com/compute/v1/projects/knada-dev/zones/europe-north1-c"),
					},
				},
			},
			expect: []*computeengine.VirtualMachine{
				{
					Name:               "test-instance-1",
					ID:                 123,
					IPs:                []string{"1.2.3.4"},
					FullyQualifiedZone: "https://www.googleapis.com/compute/v1/projects/knada-dev/zones/europe-north1-b",
					Zone:               "europe-north1-b",
				},
				{
					Name:               "test-instance-2",
					ID:                 1231234,
					IPs:                []string{"5.6.7.8"},
					FullyQualifiedZone: "https://www.googleapis.com/compute/v1/projects/knada-dev/zones/europe-north1-c",
					Zone:               "europe-north1-c",
				},
			},
		},
		{
			name:    "one instance from multiple zones",
			project: "test",
			zones:   []string{"europe-north1-b"},
			filter:  "",
			instances: map[string][]*computepb.Instance{
				"europe-north1-b": {
					{
						NetworkInterfaces: []*computepb.NetworkInterface{
							{
								NetworkIP: strPtr("1.2.3.4"),
							},
						},
						Name: strPtr("test-instance-1"),
						Id:   uint64Ptr(123),
						Zone: strPtr("https://www.googleapis.com/compute/v1/projects/knada-dev/zones/europe-north1-b"),
					},
				},
				"europe-north1-c": {
					{
						Name: strPtr("test-instance-2"),
						Id:   uint64Ptr(1231234),
						NetworkInterfaces: []*computepb.NetworkInterface{
							{
								NetworkIP: strPtr("5.6.7.8"),
							},
						},
						Zone: strPtr("https://www.googleapis.com/compute/v1/projects/knada-dev/zones/europe-north1-c"),
					},
				},
			},
			expect: []*computeengine.VirtualMachine{
				{
					Name:               "test-instance-1",
					ID:                 123,
					IPs:                []string{"1.2.3.4"},
					FullyQualifiedZone: "https://www.googleapis.com/compute/v1/projects/knada-dev/zones/europe-north1-b",
					Zone:               "europe-north1-b",
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
		project            string
		region             string
		name               string
		firewallPolicyName string
		firewallPolicies   map[string]*computepb.FirewallPolicy
		expect             any
		expectErr          bool
	}{
		{
			name:               "no firewall rules",
			project:            "test",
			region:             "europe-north1",
			firewallPolicyName: "finnes ikke",
			firewallPolicies:   map[string]*computepb.FirewallPolicy{},
			expect:             "firewall policy finnes ikke: not exists",
			expectErr:          true,
		},
		{
			name:               "one firewall rule",
			project:            "test",
			region:             "europe-north1",
			firewallPolicyName: "test-policy",
			firewallPolicies: map[string]*computepb.FirewallPolicy{
				"test-europe-north1-test-policy": {
					Name: strPtr("test-policy"),
					Rules: []*computepb.FirewallPolicyRule{
						{
							RuleName: strPtr("test-rule"),
						},
					},
				},
			},
			expect: []*computeengine.FirewallRule{
				{
					Name: "test-rule",
				},
			},
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

			got, err := c.GetFirewallRulesForRegionalPolicy(context.Background(), tc.project, tc.region, tc.firewallPolicyName)

			if tc.expectErr {
				require.Error(t, err)
				require.Equal(t, tc.expect, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expect, got)
			}
		})
	}
}
