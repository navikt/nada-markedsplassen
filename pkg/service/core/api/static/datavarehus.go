package static

import (
	"context"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
)

type datavarehusAPI struct {
	log      zerolog.Logger
	tnsNames []service.TNSName
}

func (s *datavarehusAPI) GetTNSNames(ctx context.Context) ([]service.TNSName, error) {
	s.log.Info().Msg("Would have called get TNS names")

	return s.tnsNames, nil
}

func (s *datavarehusAPI) SendJWT(ctx context.Context, keyID, signedJWT string) error {
	s.log.Info().Fields(map[string]any{
		"keyID":     keyID,
		"signedJWT": signedJWT,
	}).Msg("Would have sent JWT to Datavarehus")

	return nil
}

func NewDatavarehusAPI(log zerolog.Logger, tnsNames []service.TNSName) *datavarehusAPI {
	return &datavarehusAPI{
		tnsNames: tnsNames,
		log:      log,
	}
}
