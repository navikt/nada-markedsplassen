package computeengine

import (
	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Operations interface {
	ListVirtualMachines(ctx context.Context, filter string) ([]*VirtualMachine, error)
}

type VirtualMachine struct {
	Name string
}

type Client struct {
	apiEndpoint string
	disableAuth bool
}

func (c *Client) ListVirtualMachines(ctx context.Context, project string, zone []string, filter string) ([]*VirtualMachine, error) {
	var vms []*VirtualMachine

	for _, z := range zone {
		raw, err := c.listVirtualMachinesInZone(ctx, project, z, filter)
		if err != nil {
			return nil, err
		}

		for _, vm := range raw {
			vms = append(vms, &VirtualMachine{
				Name: vm.GetName(),
			})
		}
	}

	return vms, nil
}

func (c *Client) listVirtualMachinesInZone(ctx context.Context, project, zone, filter string) ([]*computepb.Instance, error) {
	client, err := c.newClient(ctx)
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

func (c *Client) newClient(ctx context.Context) (*compute.InstancesClient, error) {
	var options []option.ClientOption

	if c.disableAuth {
		options = append(options, option.WithoutAuthentication())
	}

	if c.apiEndpoint != "" {
		options = append(options, option.WithEndpoint(c.apiEndpoint))
	}

	client, err := compute.NewInstancesRESTClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("creating compute client: %w", err)
	}

	return client, nil
}

func NewClient(apiEndpoint string, disableAuth bool) *Client {
	return &Client{
		apiEndpoint: apiEndpoint,
		disableAuth: disableAuth,
	}
}
