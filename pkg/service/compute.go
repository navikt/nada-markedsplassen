package service

import (
	"context"
)

type ComputeAPI interface {
	// ListVirtualMachines returns all virtual machine for given filter input.
	GetVirtualMachinesByLabel(ctx context.Context, zones []string, label *Label) ([]*VirtualMachine, error)

	// GetFirewallRulesForPolicy returns all firewall rules for a specific policy.
	GetFirewallRulesForPolicy(ctx context.Context, name string) ([]*FirewallRule, error)
}

type ComputeService interface {
	// GetAllowedFirewallTags returns all firewall tags available.
	GetAllowedFirewallTags(ctx context.Context) ([]string, error)
}

type VirtualMachine struct {
	Name               string
	ID                 uint64
	Zone               string
	FullyQualifiedZone string
}

type FirewallRule struct {
	Name        string
	SecureTags  []string
	Description string
}

type Label struct {
	Key   string
	Value string
}
