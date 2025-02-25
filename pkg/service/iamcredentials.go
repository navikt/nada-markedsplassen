package service

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type IAMCredentialsAPI interface {
	SignJWT(ctx context.Context, serviceAccountEmail string, claims jwt.MapClaims) (*SignedJWT, error)
}

type SignedJWT struct {
	SignedJWT string
	KeyID     string
}

type DVHClaims struct {
	Ident               string
	IP                  string
	Databases           []string
	Reference           string
	PodName             string
	KnastContainerImage string
}

func (d *DVHClaims) ToMapClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"notifier":    "knast_notifier",
		"client_type": "knastv1",

		"ident":                 d.Ident,
		"ip":                    d.IP,
		"databases":             strings.Join(d.Databases, ","),
		"reference":             d.Reference,
		"pod_name":              d.PodName,
		"knast_container_image": d.KnastContainerImage,

		"exp": time.Now().Add(time.Minute * 5).Unix(),
		"iat": time.Now().Unix(),
	}
}
