package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
)

const (
	AzureGraphMemberOfEndpoint = `https://graph.microsoft.com/v1.0/me/memberOf/microsoft.graph.group?$select=mail,groupTypes,displayName,id`
	CacheDuration              = 1 * time.Hour
)

type AzureGroupClient struct {
	Client      *http.Client
	TexasClient *TexasClient
	log         zerolog.Logger
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type MemberOfResponse struct {
	Groups   []MemberOfGroup `json:"value"`
	NextLink *string         `json:"@odata.nextLink,omitempty"`
}

type MemberOfGroup struct {
	DisplayName string   `json:"displayName"`
	Mail        string   `json:"mail"`
	ObjectID    string   `json:"id"`
	GroupTypes  []string `json:"groupTypes"`
}

func NewAzureGroups(client *http.Client, texasClient *TexasClient, log zerolog.Logger) *AzureGroupClient {
	return &AzureGroupClient{
		Client:      client,
		TexasClient: texasClient,
		log:         log,
	}
}

func (a *AzureGroupClient) GroupsForUser(ctx context.Context, token, email string) (service.AzureGroups, error) {
	bearerToken, err := a.getBearerTokenOnBehalfOfUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("getting bearer token: %w", err)
	}

	var allGroups service.AzureGroups
	nextURL := AzureGraphMemberOfEndpoint

	// Handle pagination by following @odata.nextLink until all groups are fetched
	for nextURL != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, nextURL, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Add("Authorization", bearerToken)
		response, err := a.Client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("performing request: %w", err)
		}
		defer response.Body.Close()

		var memberOfResponse MemberOfResponse
		if err := json.NewDecoder(response.Body).Decode(&memberOfResponse); err != nil {
			return nil, fmt.Errorf("decoding response: %w", err)
		}

		// Process current page of groups
		for _, entry := range memberOfResponse.Groups {
			allGroups = append(allGroups, service.AzureGroup{
				Name:     entry.DisplayName,
				Email:    strings.ToLower(entry.Mail),
				ObjectID: entry.ObjectID,
			})
		}

		// Check if there are more pages
		if memberOfResponse.NextLink != nil {
			nextURL = *memberOfResponse.NextLink
		} else {
			nextURL = ""
		}
	}

	return allGroups, nil
}

func (a *AzureGroupClient) getBearerTokenOnBehalfOfUser(ctx context.Context, token string) (string, error) {
	token, err := a.TexasClient.Exchange(
		ctx,
		token,
		ProviderAzureAD,
		"https://graph.microsoft.com/.default",
	)
	if err != nil {
		return "", fmt.Errorf("exhanging token: %w", err)
	}

	a.log.Debug().Msg("Successfully retrieved on-behalf-of token")
	return token, nil
}
