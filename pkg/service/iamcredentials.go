package service

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type IAMCredentialsAPI interface {
	SignJWT(ctx context.Context, signer *ServiceAccount, claims jwt.MapClaims) (*SignedJWT, error)
}

type SignedJWT struct {
	SignedJWT string
	KeyID     string
}

// FIXME: We need to modify the fields in this JWT to match what DVH are expecting
// https://datatracker.ietf.org/doc/html/rfc7519#section-4.1
func NewDVHJWTClaims(ident, ip, email string) jwt.MapClaims {
	return jwt.MapClaims{
		"ident": ident,
		"ip":    ip,

		"iss": email,

		"exp": time.Now().Add(time.Minute * 5).Unix(),
		"iat": time.Now().Unix(),
	}
}
