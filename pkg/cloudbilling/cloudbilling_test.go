package cloudbilling_test

import (
	"context"
	"math"
	"testing"

	"cloud.google.com/go/billing/apiv1/billingpb"
	"github.com/navikt/nada-backend/pkg/cloudbilling"
	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/type/money"
)

// Ad-hoc test
func TestGetPrice(t *testing.T) {
	t.SkipNow()
	client := cloudbilling.NewClient()
	ctx := context.Background()
	hourlyCost, err := client.GetHourlyCostInNOKFromSKU(ctx)
	assert.NoError(t, err)
	t.Logf("%#v", hourlyCost)
}

// Test that prices are correctly extracted from upstream API responses.
func TestExtractPrice(t *testing.T) {
	client := cloudbilling.NewStaticClient(map[string]*billingpb.Sku{
		cloudbilling.SkuCpu: {
			SkuId: cloudbilling.SkuCpu,
			PricingInfo: []*billingpb.PricingInfo{
				{
					PricingExpression: &billingpb.PricingExpression{
						TieredRates: []*billingpb.PricingExpression_TierRate{
							{
								UnitPrice: &money.Money{
									Nanos: 123000000,
								},
							},
						},
					},
				},
			},
		},
		cloudbilling.SkuMemory: {
			SkuId: cloudbilling.SkuMemory,
			PricingInfo: []*billingpb.PricingInfo{
				{
					PricingExpression: &billingpb.PricingExpression{
						TieredRates: []*billingpb.PricingExpression_TierRate{
							{
								UnitPrice: &money.Money{
									Nanos: 234000000,
								},
							},
						},
					},
				},
			},
		},
	})
	ctx := context.Background()
	hourlyCost, err := client.GetHourlyCostInNOKFromSKU(ctx)
	assert.NoError(t, err)
	assert.Equal(t, &cloudbilling.VirtualMachineResourceHourlyCost{CPU: 0.123, Memory: 0.234}, hourlyCost)
}

// Test that prices are correctly extracted from upstream API responses.
func TestSkuNotFound(t *testing.T) {
	client := cloudbilling.NewStaticClient(map[string]*billingpb.Sku{})
	ctx := context.Background()
	hourlyCost, err := client.GetHourlyCostInNOKFromSKU(ctx)
	assert.ErrorIs(t, err, cloudbilling.ErrSkuNotFound)
	assert.Nil(t, hourlyCost)
}

func TestCostForConfiguration(t *testing.T) {
	hourlyCost := &cloudbilling.VirtualMachineResourceHourlyCost{
		CPU:    2,
		Memory: 1,
	}
	assert.Equal(t, 60.0, hourlyCost.CostForConfiguration(10, 40))
}

func TestCostForConfigurationNaN(t *testing.T) {
	hourlyCost := &cloudbilling.VirtualMachineResourceHourlyCost{
		CPU:    2,
		Memory: math.NaN(),
	}
	assert.True(t, math.IsNaN(hourlyCost.CostForConfiguration(10, 40)))
}
