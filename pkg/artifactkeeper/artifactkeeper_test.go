package artifactkeeper_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/navikt/nada-backend/pkg/artifactkeeper"
	"github.com/stretchr/testify/require"
)

func TestClientCreatesListsAndDeletesTokensWithBearerAuth(t *testing.T) {
	t.Parallel()

	const serviceToken = "service-marker-secret"
	var deleted bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer "+serviceToken, r.Header.Get("Authorization"))
		require.Empty(t, r.Header.Get("Cookie"))
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/auth/tokens":
			_ = json.NewEncoder(w).Encode(map[string]any{"items": []map[string]string{{"id": "old", "name": "knast:test"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/auth/tokens":
			var request artifactkeeper.CreateTokenRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
			require.Equal(t, artifactkeeper.CreateTokenRequest{
				Name: "knast:test", ExpiresInDays: 1, Scopes: []string{"read:artifacts"},
				RepoSelector: artifactkeeper.RepositorySelector{MatchPattern: "knast-pypi"},
			}, request)
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "new", "name": request.Name, "token": "workstation-marker-secret"})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/auth/tokens/old":
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	client, err := artifactkeeper.New(server.URL, serviceToken, &http.Client{Timeout: time.Second})
	require.NoError(t, err)
	tokens, err := client.ListTokens(context.Background())
	require.NoError(t, err)
	require.Equal(t, []artifactkeeper.TokenMetadata{{ID: "old", Name: "knast:test"}}, tokens)
	created, err := client.CreateToken(context.Background(), artifactkeeper.CreateTokenRequest{
		Name: "knast:test", ExpiresInDays: 1, Scopes: []string{"read:artifacts"}, RepoSelector: artifactkeeper.RepositorySelector{MatchPattern: "knast-pypi"},
	})
	require.NoError(t, err)
	require.Equal(t, "new", created.ID)
	require.NoError(t, client.DeleteToken(context.Background(), "old"))
	require.True(t, deleted)
}

func TestClientDoesNotExposeResponseBodyInErrors(t *testing.T) {
	t.Parallel()
	const marker = "never-log-or-return-this-token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, marker, http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)
	client, err := artifactkeeper.New(server.URL, "service-token", server.Client())
	require.NoError(t, err)
	_, err = client.CreateToken(context.Background(), artifactkeeper.CreateTokenRequest{})
	require.Error(t, err)
	require.NotContains(t, err.Error(), marker)
}

func TestClientRetriesServerErrorsOnce(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) == 1 {
			http.Error(w, "temporary", http.StatusServiceUnavailable)
			return
		}
		_ = json.NewEncoder(w).Encode([]artifactkeeper.TokenMetadata{})
	}))
	t.Cleanup(server.Close)
	client, err := artifactkeeper.New(server.URL, "service-token", server.Client())
	require.NoError(t, err)

	_, err = client.ListTokens(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, 2, requests.Load())
}

func TestClientDoesNotRetryClientErrors(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)
	client, err := artifactkeeper.New(server.URL, "service-token", server.Client())
	require.NoError(t, err)

	_, err = client.ListTokens(context.Background())
	require.Error(t, err)
	require.EqualValues(t, 1, requests.Load())
}
