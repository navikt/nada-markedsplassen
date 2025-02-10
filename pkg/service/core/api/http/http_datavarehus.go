package http

import (
	"context"
	"fmt"
	"strings"

	"github.com/navikt/nada-backend/pkg/datavarehus"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
)

type datavarehusAPI struct {
	ops datavarehus.Operations
	log zerolog.Logger
}

var _ service.DatavarehusAPI = &datavarehusAPI{}

func (c *datavarehusAPI) GetTNSNames(ctx context.Context) ([]service.TNSName, error) {
	const op errs.Op = "datavarehusAPI.GetTNSNames"

	raw, err := c.ops.GetTNSNames(ctx)
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeDatavarehus, op, err)
	}

	tnsHosts := make([]service.TNSName, len(raw))
	for idx, tns := range raw {
		tnsHosts[idx] = service.TNSName{
			TnsName:     tns.TnsName,
			Name:        tns.Name,
			Description: tns.Description,
			Host:        strings.ToLower(tns.Host),
			Port:        tns.Port,
			ServiceName: tns.ServiceName,
		}
	}

	return tnsHosts, nil
}

func (c *datavarehusAPI) SendJWT(ctx context.Context, keyID, signedJWT string) error {
	const op errs.Op = "datavarehusAPI.SendJWT"

	err := c.ops.SendJWT(ctx, keyID, signedJWT)
	if err != nil {
		return errs.E(errs.IO, service.CodeDatavarehus, op, fmt.Errorf("sending JWT to Datavarehus: %w", err))
	}

	c.log.Info().Fields(map[string]any{
		"keyID": keyID,
		"jwt":   signedJWT,
	}).Msg("Successfully sent JWT to Datavarehus")

	return nil
}

func NewDatavarehusAPI(ops datavarehus.Operations, log zerolog.Logger) service.DatavarehusAPI {
	return &datavarehusAPI{
		ops: ops,
		log: log,
	}
}
