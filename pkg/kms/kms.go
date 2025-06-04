package kms

import (
	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"context"
	"fmt"
	"google.golang.org/api/option"
)

type Operations interface {
	Encrypt(ctx context.Context, id *KeyIdentifier, plaintext []byte) ([]byte, error)
	Decrypt(ctx context.Context, id *KeyIdentifier, ciphertext []byte) ([]byte, error)
}

type KeyIdentifier struct {
	Project  string
	Location string
	Keyring  string
	KeyName  string
}

func (i KeyIdentifier) ResourceName() string {
	return fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", i.Project, i.Location, i.Keyring, i.KeyName)
}

type Client struct {
	apiEndpoint string
	disableAuth bool
}

func (c *Client) Encrypt(ctx context.Context, id *KeyIdentifier, plaintext []byte) ([]byte, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	raw, err := client.Encrypt(ctx, &kmspb.EncryptRequest{
		Name:      id.ResourceName(),
		Plaintext: plaintext,
	})
	if err != nil {
		return nil, fmt.Errorf("encrypting data: %w", err)
	}

	return raw.Ciphertext, nil
}

func (c *Client) Decrypt(ctx context.Context, id *KeyIdentifier, ciphertext []byte) ([]byte, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	raw, err := client.Decrypt(ctx, &kmspb.DecryptRequest{
		Name:       id.ResourceName(),
		Ciphertext: ciphertext,
	})
	if err != nil {
		return nil, fmt.Errorf("decrypting data: %w", err)
	}

	return raw.Plaintext, nil
}

func (c *Client) newClient(ctx context.Context) (*kms.KeyManagementClient, error) {
	var options []option.ClientOption

	if c.disableAuth {
		options = append(options, option.WithoutAuthentication())
	}

	if c.apiEndpoint != "" {
		options = append(options, option.WithEndpoint(c.apiEndpoint))
	}

	client, err := kms.NewKeyManagementRESTClient(ctx, options...)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func NewClient(apiEndpoint string, disableAuth bool) *Client {
	return &Client{
		apiEndpoint: apiEndpoint,
		disableAuth: disableAuth,
	}
}
