package postgres

import (
	"context"

	"github.com/navikt/nada-backend/pkg/cache"
	"github.com/navikt/nada-backend/pkg/cloudbilling"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.CloudBillingAPI = &machineCostCache{}

type machineCostCache struct {
	api   service.CloudBillingAPI
	cache cache.Cacher
}

const cacheKey = "machinecost:nok"

func (c *machineCostCache) GetHourlyCostInNOKFromSKU(ctx context.Context) (*cloudbilling.VirtualMachineResourceHourlyCost, error) {
	const op errs.Op = "machineCostCache.GetHourlyCostInNOKFromSKU"

	hourlyCost := &cloudbilling.VirtualMachineResourceHourlyCost{}
	valid := c.cache.Get(cacheKey, &hourlyCost)
	if valid {
		return hourlyCost, nil
	}

	hourlyCost, err := c.api.GetHourlyCostInNOKFromSKU(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	c.cache.Set(cacheKey, hourlyCost)

	return hourlyCost, nil
}

func NewMachineCostCache(api service.CloudBillingAPI, cache cache.Cacher) service.CloudBillingAPI {
	return &machineCostCache{
		api:   api,
		cache: cache,
	}
}
