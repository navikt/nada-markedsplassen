package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

type TexasEndpoints struct {
	Exchange   string `yaml:"exchange"`
	Introspect string `yaml:"introspect"`
}

type TexasClient struct {
	client    *http.Client
	endpoints TexasEndpoints
	log       zerolog.Logger
}

func NewTexasClient(client *http.Client, endpoints TexasEndpoints, log zerolog.Logger) *TexasClient {
	return &TexasClient{
		client:    client,
		endpoints: endpoints,
		log:       log,
	}
}

const (
	ProviderAzureAD = "azuread"
	ContentTypeJSON = "application/json"
)

type TexasIntrospectRequest struct {
	IdentityProvider string `json:"identity_provider"`
	Token            string `json:"token"`
}

type TexasExchangeRequest struct {
	IdentityProvider string `json:"identity_provider"`
	UserToken        string `json:"user_token"`
	Target           string `json:"target"`
}

type TokenClaims struct {
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	NavIdent          string `json:"NAVident"`
	Exp               int64  `json:"exp"`
}

type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func (c *TexasClient) Exchange(ctx context.Context, token, provider, targetScope string) (string, error) {
	oboToken, err := post[Token](
		ctx,
		c.client,
		TexasExchangeRequest{
			IdentityProvider: provider,
			UserToken:        strings.TrimPrefix(token, "Bearer "),
			Target:           targetScope,
		},
		c.endpoints.Exchange,
	)
	if err != nil {
		return "", fmt.Errorf("token exchange failed %w", err)
	}

	return fmt.Sprintf("%s %s", oboToken.TokenType, oboToken.AccessToken), nil
}

func (c *TexasClient) Introspect(ctx context.Context, token, provider string) (TokenClaims, error) {
	return post[TokenClaims](
		ctx,
		c.client,
		TexasIntrospectRequest{
			IdentityProvider: provider,
			Token:            strings.TrimPrefix(token, "Bearer "),
		},
		c.endpoints.Introspect,
	)
}

func post[T Token | TokenClaims, R TexasExchangeRequest | TexasIntrospectRequest](ctx context.Context, client *http.Client, body R, endpoint string) (T, error) {
	var result T
	jsonData, err := json.Marshal(body)
	if err != nil {
		return result, fmt.Errorf("failed to marshal body %v", err)
	}

	request, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return result, fmt.Errorf("failed to create request %v", err)
	}
	request.Header.Add("content-type", ContentTypeJSON)

	response, err := client.Do(request)
	if err != nil {
		return result, fmt.Errorf("request to Texas failed %v", err)
	}
	if response.StatusCode != 200 {
		return result, fmt.Errorf("unexpected response from Texas %s", response.Status)
	}

	defer response.Body.Close()
	if err = json.NewDecoder(response.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("failed to decode response %v", err)
	}

	return result, nil
}
