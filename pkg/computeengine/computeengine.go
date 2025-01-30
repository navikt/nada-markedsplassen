package computeengine

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var ErrNotExists = errors.New("not exists")

var _ Operations = &Client{}

type Operations interface {
	// GetVirtualMachinesByLabel by label value returns all virtual machine with a given label.
	ListVirtualMachines(ctx context.Context, project string, zone []string, filterExpression string) ([]*VirtualMachine, error)

	// GetVirtualMachinesByLabel by label value returns all virtual machine with a given label.
	GetVirtualMachinesByLabel(ctx context.Context, project string, zones []string, label *Label) ([]*VirtualMachine, error)

	// GetFirewallRulesForRegionalPolicy returns all firewall rules for a specific policy.
	GetFirewallRulesForRegionalPolicy(ctx context.Context, project, region, name string) ([]*FirewallRule, error)
}

type Label struct {
	Key   string
	Value string
}

type FirewallRule struct {
	Name        string
	SecureTags  []string
	Description string
}

type VirtualMachine struct {
	Name               string
	ID                 uint64
	Zone               string
	FullyQualifiedZone string
	IPs                []string
}

type Client struct {
	apiEndpoint string
	disableAuth bool
}

func (c *Client) GetFirewallRulesForRegionalPolicy(ctx context.Context, project, region, name string) ([]*FirewallRule, error) {
	client, err := c.newFirewallPoliciesClient(ctx)
	if err != nil {
		return nil, err
	}

	req := &computepb.GetRegionNetworkFirewallPolicyRequest{
		Project:        project,
		Region:         region,
		FirewallPolicy: name,
	}

	policy, err := client.Get(ctx, req)
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return nil, fmt.Errorf("firewall policy %s: %w", name, ErrNotExists)
		}

		return nil, fmt.Errorf("getting firewall policy (%s): %w", name, err)
	}

	var rules []*FirewallRule

	for _, rule := range policy.GetRules() {
		var tagNames []string

		for _, tag := range rule.GetTargetSecureTags() {
			tagNames = append(tagNames, tag.GetName())
		}

		rules = append(rules, &FirewallRule{
			Name:        rule.GetRuleName(),
			SecureTags:  tagNames,
			Description: rule.GetDescription(),
		})
	}

	return rules, nil
}

func (c *Client) GetVirtualMachinesByLabel(ctx context.Context, project string, zones []string, label *Label) ([]*VirtualMachine, error) {
	return c.ListVirtualMachines(ctx, project, zones, fmt.Sprintf("labels.%s:%s", label.Key, label.Value))
}

func (c *Client) ListVirtualMachines(ctx context.Context, project string, zone []string, filter string) ([]*VirtualMachine, error) {
	var vms []*VirtualMachine

	for _, z := range zone {
		raw, err := c.listVirtualMachinesInZone(ctx, project, z, filter)
		if err != nil {
			return nil, err
		}

		for _, vm := range raw {
			ips := make([]string, len(vm.GetNetworkInterfaces()))
			for i, ni := range vm.GetNetworkInterfaces() {
				ips[i] = ni.GetNetworkIP()
			}

			vms = append(vms, &VirtualMachine{
				Name:               vm.GetName(),
				ID:                 vm.GetId(),
				FullyQualifiedZone: vm.GetZone(),
				Zone:               z,
				IPs:                ips,
			})
		}
	}

	return vms, nil
}

func (c *Client) listVirtualMachinesInZone(ctx context.Context, project, zone, filter string) ([]*computepb.Instance, error) {
	client, err := c.newInstancesClient(ctx)
	if err != nil {
		return nil, err
	}

	req := &computepb.ListInstancesRequest{
		Project: project,
		Zone:    zone,
		Filter:  &filter,
	}

	it := client.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("listing instances: %w", err)
	}

	var instances []*computepb.Instance

	for {
		instance, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("iterating instances: %w", err)
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

func (c *Client) newInstancesClient(ctx context.Context) (*compute.InstancesClient, error) {
	var options []option.ClientOption

	if c.disableAuth {
		options = append(options, option.WithoutAuthentication())
	}

	if c.apiEndpoint != "" {
		options = append(options, option.WithEndpoint(c.apiEndpoint))
	}

	client, err := compute.NewInstancesRESTClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("creating compute instances client: %w", err)
	}

	return client, nil
}

func (c *Client) newFirewallPoliciesClient(ctx context.Context) (*compute.RegionNetworkFirewallPoliciesClient, error) {
	var options []option.ClientOption

	if c.disableAuth {
		options = append(options, option.WithoutAuthentication())
	}

	if c.apiEndpoint != "" {
		options = append(options, option.WithEndpoint(c.apiEndpoint))
	}

	client, err := compute.NewRegionNetworkFirewallPoliciesRESTClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("creating firewall policies client: %w", err)
	}

	return client, nil
}

func NewClient(apiEndpoint string, disableAuth bool) *Client {
	return &Client{
		apiEndpoint: apiEndpoint,
		disableAuth: disableAuth,
	}
}
