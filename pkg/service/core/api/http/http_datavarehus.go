package http

import (
	"context"

	"github.com/navikt/nada-backend/pkg/datavarehus"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

type datavarehusAPI struct {
	client *datavarehus.Client
}

var _ service.DatavarehusAPI = &datavarehusAPI{}

func (c *datavarehusAPI) GetTNSNames(ctx context.Context) ([]service.TNSName, error) {
	const op errs.Op = "datavarehusAPI.GetTNSNames"

	raw, err := c.client.GetTNSNames(ctx)
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeDatavarehus, op, err)
	}

	tnsHosts := make([]service.TNSName, len(raw))
	for idx, tns := range raw {
		tnsHosts[idx] = service.TNSName{
			TnsName:     tns.TnsName,
			Name:        tns.Name,
			Description: tns.Description,
			Host:        tns.Host,
			Port:        tns.Port,
			ServiceName: tns.ServiceName,
		}
	}

	return tnsHosts, nil
}

func (c *datavarehusAPI) SendJWT(ctx context.Context, keyID, signedJWT string) error {
	const op errs.Op = "datavarehusAPI.SendJWT"

	err := c.client.SendJWT(ctx, keyID, signedJWT)
	if err != nil {
		return errs.E(errs.IO, service.CodeDatavarehus, op, err)
	}

	return nil
}

func NewDatavarehusAPI(client *datavarehus.Client) service.DatavarehusAPI {
	return &datavarehusAPI{
		client: client,
	}
}
