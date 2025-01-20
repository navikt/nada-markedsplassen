package iamcredentials

import (
	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"cloud.google.com/go/iam/credentials/apiv1/credentialspb"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/api/option"
)

type Operations interface {
	SignJWT(ctx context.Context, signer *ServiceAccount, claims jwt.MapClaims) (*SignedJWT, error)
}

type ServiceAccount struct {
	Email string
}

func (s *ServiceAccount) FullyQualifiedName() string {
	return fmt.Sprintf("projects/-/serviceAccounts/%s", s.Email)
}

type SignedJWT struct {
	SignedJWT string
	KeyID     string
}

type Client struct {
	apiEndpoint string
	disableAuth bool
}

func (c *Client) SignJWT(ctx context.Context, signer *ServiceAccount, claims jwt.MapClaims) (*SignedJWT, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(claims)
	if err != nil {
		return nil, fmt.Errorf("marshalling payload: %w", err)
	}

	raw, err := client.SignJwt(ctx, &credentialspb.SignJwtRequest{
		Name:    signer.FullyQualifiedName(),
		Payload: string(data),
	})
	if err != nil {
		return nil, fmt.Errorf("signing jwt: %w", err)
	}

	return &SignedJWT{
		SignedJWT: raw.SignedJwt,
		KeyID:     raw.KeyId,
	}, nil

}

func (c *Client) newClient(ctx context.Context) (*credentials.IamCredentialsClient, error) {
	var options []option.ClientOption

	if c.disableAuth {
		fmt.Println("Disabling authentication")
		options = append(options, option.WithoutAuthentication())
	}

	if c.apiEndpoint != "" {
		fmt.Printf("Using endpoint: %s\n", c.apiEndpoint)
		options = append(options, option.WithEndpoint(c.apiEndpoint))
	}

	client, err := credentials.NewIamCredentialsRESTClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("creating new IAM credentials client: %w", err)
	}

	return client, nil
}

func New(apiEndpoint string, disableAuth bool) *Client {
	return &Client{
		apiEndpoint: apiEndpoint,
		disableAuth: disableAuth,
	}
}
