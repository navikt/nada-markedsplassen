package cloudbilling

import (
	"context"
	"errors"
	"fmt"
	"math"

	billing "cloud.google.com/go/billing/apiv1"
	"cloud.google.com/go/billing/apiv1/billingpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// from https://cloud.google.com/skus?hl=en&filter=n2d&currency=NOK
const (
	SkuComputeEngineGroup = "services/6F81-5844-456A" // "Compute Engine" service, from https://cloud.google.com/skus/sku-groups/on-demand-vms
	SkuCpu                = "4791-064D-3189"
	SkuMemory             = "F579-1568-AEED"
)

var ErrSkuNotFound = errors.New("SKU not present in dataset")

type Operations interface {
	GetHourlyCostInNOKFromSKU(ctx context.Context) (*VirtualMachineResourceHourlyCost, error)
}

var _ Operations = &client{}

// VirtualMachineResourceHourlyCost contains the hourly unit cost of the CPU and memory components of a virtual machine.
// i.e., cost per one unit per hour.
type VirtualMachineResourceHourlyCost struct {
	CPU    float64
	Memory float64
}

// CostForConfiguration - given a virtual machine with X cores and Y GB memory, returns the total hourly cost.
// Returns NaN if some cost info does not exist.
func (c VirtualMachineResourceHourlyCost) CostForConfiguration(cpuAmount, memoryAmount uint) float64 {
	return float64(cpuAmount)*c.CPU + float64(memoryAmount)*c.Memory
}

type client struct {
	fetchCallbackFn func(ctx context.Context) (map[string]*billingpb.Sku, error)
}

func NewClient() Operations {
	cli := &client{}
	cli.fetchCallbackFn = cli.getComputeEngineGroupSkus
	return cli
}

func NewStaticClient(skuData map[string]*billingpb.Sku) Operations {
	return &client{
		fetchCallbackFn: func(ctx context.Context) (map[string]*billingpb.Sku, error) {
			return skuData, nil
		},
	}
}

func (c *client) internalClient(ctx context.Context) (*billing.CloudCatalogClient, error) {
	var options []option.ClientOption

	cli, err := billing.NewCloudCatalogRESTClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("creating cloud billing client: %w", err)
	}

	return cli, nil
}

func (c *client) getComputeEngineGroupSkus(ctx context.Context) (map[string]*billingpb.Sku, error) {
	cli, err := c.internalClient(ctx)
	if err != nil {
		return nil, err
	}

	request := &billingpb.ListSkusRequest{
		Parent:       SkuComputeEngineGroup,
		CurrencyCode: "NOK",
	}

	skuIterator := cli.ListSkus(ctx, request)

	skus := make(map[string]*billingpb.Sku)

	for {
		skuPricing, err := skuIterator.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if err != nil {
			return nil, err
		}

		skus[skuPricing.SkuId] = skuPricing
	}

	return skus, nil
}

func (c *client) GetHourlyCostInNOKFromSKU(ctx context.Context) (*VirtualMachineResourceHourlyCost, error) {
	skus, err := c.fetchCallbackFn(ctx)
	if err != nil {
		return nil, err
	}

	for _, sku := range []string{SkuCpu, SkuMemory} {
		if _, ok := skus[sku]; !ok {
			return nil, fmt.Errorf("%w: %q", ErrSkuNotFound, sku)
		}
	}

	return &VirtualMachineResourceHourlyCost{
		CPU:    getInnerCost(skus[SkuCpu]),
		Memory: getInnerCost(skus[SkuMemory]),
	}, nil
}

// Helper function to extract the price from a SKU.
func getInnerCost(sku *billingpb.Sku) float64 {
	for _, pricing := range sku.PricingInfo {
		for _, rate := range pricing.PricingExpression.TieredRates {
			return float64(rate.UnitPrice.Nanos) / 1e9
		}
	}
	return math.NaN()
}
