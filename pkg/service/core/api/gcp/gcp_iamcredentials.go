package gcp

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/iamcredentials"
	"github.com/navikt/nada-backend/pkg/service"
)

type iamCredentialsAPI struct {
	ops iamcredentials.Operations
}

var _ service.IAMCredentialsAPI = &iamCredentialsAPI{}

func (a *iamCredentialsAPI) SignJWT(ctx context.Context, signer *service.ServiceAccount, claims jwt.MapClaims) (*service.SignedJWT, error) {
	const op errs.Op = "iamCredentialsAPI.SignJWT"

	jwt, err := a.ops.SignJWT(ctx, &iamcredentials.ServiceAccount{
		Email: signer.Email,
	}, claims)
	if err != nil {
		return nil, errs.E(err)
	}

	return &service.SignedJWT{
		KeyID:     jwt.KeyID,
		SignedJWT: jwt.SignedJWT,
	}, nil
}

func NewIAMCredentialsAPI(ops iamcredentials.Operations) *iamCredentialsAPI {
	return &iamCredentialsAPI{
		ops: ops,
	}
}
