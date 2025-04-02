package service

import (
	"context"
	"errors"
)

var (
	ErrMultipleVMs = errors.New("multiple virtual machines found")
	ErrNoVMs       = errors.New("no virtual machines found")
)

type ComputeAPI interface {
	// GetVirtualMachineByLabel returns a single virtual machine for given filter input, or an error if
	// none or more than one virtual machine are found.
	GetVirtualMachineByLabel(ctx context.Context, zones []string, label *Label) (*VirtualMachine, error)

	// GetVirtualMachinesByLabel returns all virtual machine for given filter input.
	GetVirtualMachinesByLabel(ctx context.Context, zones []string, label *Label) ([]*VirtualMachine, error)

	// GetFirewallRulesForRegionalPolicy returns all firewall rules for a specific policy.
	GetFirewallRulesForRegionalPolicy(ctx context.Context, project, region, name string) ([]*FirewallRule, error)
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
	IPs                []string
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
