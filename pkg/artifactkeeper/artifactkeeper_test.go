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

func TestClientCreatesTokenWithBearerAuth(t *testing.T) {
	t.Parallel()

	const serviceToken = "service-marker-secret"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer "+serviceToken, r.Header.Get("Authorization"))
		require.Empty(t, r.Header.Get("Cookie"))
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/auth/tokens":
			var request artifactkeeper.CreateTokenRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
			require.Equal(t, artifactkeeper.CreateTokenRequest{
				Name: "knast:test", ExpiresInDays: 1, Scopes: []string{"read:artifacts"},
				RepoSelector: artifactkeeper.RepositorySelector{MatchLabels: map[string]string{"knast-default": "true"}},
			}, request)
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "new", "name": request.Name, "token": "workstation-marker-secret"})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	client, err := artifactkeeper.New(server.URL, serviceToken, &http.Client{Timeout: time.Second})
	require.NoError(t, err)
	created, err := client.CreateToken(context.Background(), artifactkeeper.CreateTokenRequest{
		Name: "knast:test", ExpiresInDays: 1, Scopes: []string{"read:artifacts"}, RepoSelector: artifactkeeper.RepositorySelector{MatchLabels: map[string]string{"knast-default": "true"}},
	})
	require.NoError(t, err)
	require.Equal(t, "new", created.ID)
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
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "new", "name": "knast:test", "token": "token"})
	}))
	t.Cleanup(server.Close)
	client, err := artifactkeeper.New(server.URL, "service-token", server.Client())
	require.NoError(t, err)

	_, err = client.CreateToken(context.Background(), artifactkeeper.CreateTokenRequest{})
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

	_, err = client.CreateToken(context.Background(), artifactkeeper.CreateTokenRequest{})
	require.Error(t, err)
	require.EqualValues(t, 1, requests.Load())
}
