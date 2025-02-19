package datavarehus_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/navikt/nada-backend/pkg/datavarehus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockServer(statusCode int, responseBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	}))
}

func TestClient_GetTNSNames(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		expectErr    bool
		expect       any
	}{
		{
			name:       "Success",
			statusCode: http.StatusOK,
			responseBody: func() string {
				content, err := os.ReadFile("testdata/dmo_ops_tnsnames.json")
				require.NoError(t, err)

				return string(content)
			}(),
			expectErr: false,
			expect: []datavarehus.TNSName{
				{
					TnsName:     "DVH-P",
					Name:        "DVH-produksjon",
					Description: "Produksjonsdatabasen til Nav-datavarehus. Den benyttes til ETL/lagring og konsum av dataobjekter i fagsystemsonen.",
					Host:        "DM08-SCAN.ADEO.NO",
					Port:        "1521",
					ServiceName: "DWH_HA",
				},
				{
					TnsName:     "DVH-U",
					Name:        "DVH-utvikling",
					Description: "Kopi av produksjon (manuelt oppdatert) som benyttes til utvikling.",
					Host:        "DMV34-SCAN.ADEO.NO",
					Port:        "1521",
					ServiceName: "CCDWHU1_HA",
				},
			},
		},
		{
			name:         "Error",
			statusCode:   http.StatusInternalServerError,
			responseBody: `{"error": "internal server error"}`,
			expectErr:    true,
			expect:       `received status code 500 from datavarehus: error {"error": "internal server error"}`,
		},
		{
			name:         "HasMore",
			statusCode:   http.StatusOK,
			responseBody: `{"items": [], "hasMore": true}`,
			expectErr:    true,
			expect:       "more items available than returned, please implement pagination",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := mockServer(tt.statusCode, tt.responseBody)
			defer server.Close()

			client := datavarehus.New(server.URL, "clientID", "clientSecret")
			ctx := context.Background()

			got, err := client.GetTNSNames(ctx)
			if tt.expectErr {
				require.Error(t, err)
				assert.Equal(t, tt.expect, err.Error())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.expect, got)
			}
		})
	}
}

func TestClient_SendJWT(t *testing.T) {
	tests := []struct {
		name            string
		keyID           string
		signedJWT       string
		token           *datavarehus.Token
		tokenStatusCode int
		expect          any
		expectErr       bool
	}{
		{
			name:      "Success",
			keyID:     "78638r89867983r",
			signedJWT: "eye038r039083.293r473092809tr8903.39r857039270953",
			token: &datavarehus.Token{
				AccessToken: "valid_token",
				TokenType:   "bearer",
				ExpiresIn:   3600,
			},
			tokenStatusCode: http.StatusOK,
			expect: &datavarehus.Body{
				KeyID:     "78638r89867983r",
				SignedJWT: "eye038r039083.293r473092809tr8903.39r857039270953",
			},
			expectErr: false,
		},
		{
			name:      "Error",
			keyID:     "78638r89867983r",
			signedJWT: "eye038r039083.293r473092809tr8903.39r857039270953",
			token: &datavarehus.Token{
				AccessToken: "invalid_token",
				TokenType:   "bearer",
				ExpiresIn:   3600,
			},
			tokenStatusCode: http.StatusInternalServerError,
			expect:          "received status code 500 from datavarehus: error something bad",
			expectErr:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			token, err := json.Marshal(tc.token)
			require.NoError(t, err)

			httpmock.RegisterResponder("POST", "http://example.com/ords/dvh/oauth/token",
				httpmock.NewBytesResponder(tc.tokenStatusCode, token))

			httpmock.RegisterResponder("POST", "http://example.com/ords/dvh/dvh_dmo/wli_rest/notify",
				func(request *http.Request) (*http.Response, error) {
					require.Equal(t, fmt.Sprintf("Bearer %s", tc.token.AccessToken), request.Header.Get("Authorization"))
					require.Equal(t, "application/json", request.Header.Get("Content-Type"))

					if tc.expectErr {
						return httpmock.NewStringResponse(http.StatusInternalServerError, "something bad"), nil
					} else {
						b := &datavarehus.Body{}
						err := json.NewDecoder(request.Body).Decode(b)
						require.NoError(t, err)
						require.Equal(t, tc.expect, b)
						return httpmock.NewStringResponse(http.StatusNoContent, ""), nil
					}
				})

			client := datavarehus.New("http://example.com", "clientID", "clientSecret")
			ctx := context.Background()

			err = client.SendJWT(ctx, tc.keyID, tc.signedJWT)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Equal(t, tc.expect, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
