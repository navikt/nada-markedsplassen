package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/navikt/nada-backend/pkg/auth"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func newTexasServer(status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if body != "" {
			w.Write([]byte(body))
		}
	}))
}

func TestMiddleware_Handler_TexasIntrospectionFailure(t *testing.T) {
	tests := []struct {
		name                   string
		texasStatusCode        int
		TexasResponseBody      string
		expectDownstreamCalled bool
		expectStatusCode       int
		expectBody             string
	}{
		{
			name:                   "Texas Internal Server Error",
			texasStatusCode:        500,
			TexasResponseBody:      "Internal server error",
			expectDownstreamCalled: false,
			expectStatusCode:       401,
			expectBody:             `{"error": "Unauthorized"}`,
		},
		{
			name:                   "Texas Invalid Token",
			texasStatusCode:        200,
			TexasResponseBody:      `{"active": false, "error": "Invalid token"}`,
			expectDownstreamCalled: false,
			expectStatusCode:       401,
			expectBody:             `{"error": "Unauthorized"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			texasServer := newTexasServer(tt.texasStatusCode, tt.TexasResponseBody)
			defer texasServer.Close()

			texasClient := auth.NewTexasClient(
				http.DefaultClient,
				auth.TexasEndpoints{
					Introspect: texasServer.URL + "/introspect",
					Exchange:   texasServer.URL + "/exchange",
				},
				zerolog.Nop(),
			)

			middleware := auth.NewMiddleware(
				nil,
				nil,
				texasClient,
				[]string{},
				zerolog.Nop(),
			)

			downstreamCalled := false
			downstreamHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				downstreamCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.Handler(downstreamHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("authorization", "Bearer test-token")
			req = req.WithContext(context.Background())

			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectStatusCode, recorder.Code)
			assert.Contains(t, recorder.Body.String(), tt.expectBody)
			assert.Equal(t, tt.expectDownstreamCalled, downstreamCalled)
		})
	}
}
