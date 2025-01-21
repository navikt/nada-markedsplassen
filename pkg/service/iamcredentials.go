package service

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type IAMCredentialsAPI interface {
	SignJWT(ctx context.Context, signer *ServiceAccount, claims jwt.MapClaims) (*SignedJWT, error)
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

// FIXME: We need to modify the fields in this JWT to match what DVH are expecting
// https://datatracker.ietf.org/doc/html/rfc7519#section-4.1
func (d *DVHClaims) ToMapClaims(ident, ip, podName string) jwt.MapClaims {
	return jwt.MapClaims{
		"ident":                 d.Ident,
		"ip":                    d.IP,
		"databases":             strings.Join(d.Databases, ","),
		"reference":             d.Reference,
		"pod_name":              podName,
		"knast_container_image": d.KnastContainerImage,

		"exp": time.Now().Add(time.Minute * 5).Unix(),
		"iat": time.Now().Unix(),
	}
}
