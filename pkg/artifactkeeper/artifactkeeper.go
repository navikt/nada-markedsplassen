package artifactkeeper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	tokensPath       = "/api/v1/auth/tokens"
	maxResponseBytes = 1 << 20
	maxAttempts      = 2
)

var (
	requestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "nada_backend",
		Subsystem: "artifact_keeper",
		Name:      "requests_total",
		Help:      "Artifact Keeper API requests by operation and result.",
	}, []string{"operation", "result"})
	requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "nada_backend",
		Subsystem: "artifact_keeper",
		Name:      "request_duration_seconds",
		Help:      "Artifact Keeper API request duration by operation.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"operation"})
	retriesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "nada_backend",
		Subsystem: "artifact_keeper",
		Name:      "retries_total",
		Help:      "Artifact Keeper API retries by operation.",
	}, []string{"operation"})
)

func Collectors() []prometheus.Collector {
	return []prometheus.Collector{requestsTotal, requestDuration, retriesTotal}
}

type Operations interface {
	ListTokens(ctx context.Context) ([]TokenMetadata, error)
	CreateToken(ctx context.Context, request CreateTokenRequest) (*CreatedToken, error)
	DeleteToken(ctx context.Context, id string) error
}

type Client struct {
	baseURL      *url.URL
	serviceToken string
	httpClient   *http.Client
}

type RepositorySelector struct {
	MatchPattern string `json:"match_pattern"`
}

type CreateTokenRequest struct {
	Name          string             `json:"name"`
	ExpiresInDays int                `json:"expires_in_days"`
	Scopes        []string           `json:"scopes"`
	RepoSelector  RepositorySelector `json:"repo_selector"`
}

type CreatedToken struct {
	ID    string `json:"id"`
	Token string `json:"token"`
	Name  string `json:"name"`
}

type TokenMetadata struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type listTokensResponse struct {
	Items      []TokenMetadata `json:"items"`
	Tokens     []TokenMetadata `json:"tokens"`
	NextCursor string          `json:"next_cursor"`
	NextPage   int             `json:"next_page"`
}

func New(apiURL, serviceToken string, httpClient *http.Client) (*Client, error) {
	parsed, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("parsing Artifact Keeper API URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("Artifact Keeper API URL must be absolute")
	}
	if serviceToken == "" {
		return nil, errors.New("Artifact Keeper service token is empty")
	}
	if httpClient == nil {
		return nil, errors.New("Artifact Keeper HTTP client is nil")
	}

	return &Client{baseURL: parsed, serviceToken: serviceToken, httpClient: httpClient}, nil
}

func (c *Client) ListTokens(ctx context.Context) ([]TokenMetadata, error) {
	var result []TokenMetadata
	cursor := ""
	page := 1
	for {
		query := url.Values{}
		if cursor != "" {
			query.Set("cursor", cursor)
		} else {
			query.Set("page", strconv.Itoa(page))
		}

		body, err := c.do(ctx, "list", http.MethodGet, tokensPath, query, nil)
		if err != nil {
			return nil, err
		}

		var direct []TokenMetadata
		if err := json.Unmarshal(body, &direct); err == nil {
			return append(result, direct...), nil
		}

		var response listTokensResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, errors.New("decoding Artifact Keeper token list response")
		}
		items := response.Items
		if items == nil {
			items = response.Tokens
		}
		result = append(result, items...)
		switch {
		case response.NextCursor != "":
			cursor = response.NextCursor
		case response.NextPage > page:
			page = response.NextPage
		default:
			return result, nil
		}
	}
}

func (c *Client) CreateToken(ctx context.Context, request CreateTokenRequest) (*CreatedToken, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, errors.New("encoding Artifact Keeper create token request")
	}

	responseBody, err := c.do(ctx, "create", http.MethodPost, tokensPath, nil, body)
	if err != nil {
		return nil, err
	}

	var response CreatedToken
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, errors.New("decoding Artifact Keeper create token response")
	}
	return &response, nil
}

func (c *Client) DeleteToken(ctx context.Context, id string) error {
	if id == "" || strings.Contains(id, "/") {
		return errors.New("invalid Artifact Keeper token ID")
	}
	_, err := c.do(ctx, "delete", http.MethodDelete, tokensPath+"/"+url.PathEscape(id), nil, nil)
	return err
}

func (c *Client) do(ctx context.Context, operation, method, requestPath string, query url.Values, body []byte) ([]byte, error) {
	started := time.Now()
	defer func() { requestDuration.WithLabelValues(operation).Observe(time.Since(started).Seconds()) }()

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			retriesTotal.WithLabelValues(operation).Inc()
			delay := 100*time.Millisecond + time.Duration(rand.IntN(100))*time.Millisecond
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				requestsTotal.WithLabelValues(operation, "cancelled").Inc()
				return nil, ctx.Err()
			case <-timer.C:
			}
		}

		requestURL := c.baseURL.ResolveReference(&url.URL{Path: requestPath})
		requestURL.RawQuery = query.Encode()
		req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), bytes.NewReader(body))
		if err != nil {
			return nil, errors.New("creating Artifact Keeper request")
		}
		req.Header.Set("Authorization", "Bearer "+c.serviceToken)
		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = errors.New("Artifact Keeper request failed")
			continue
		}
		responseBody, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
		closeErr := resp.Body.Close()
		if readErr != nil || closeErr != nil {
			lastErr = errors.New("reading Artifact Keeper response")
			continue
		}
		if len(responseBody) > maxResponseBytes {
			requestsTotal.WithLabelValues(operation, "invalid_response").Inc()
			return nil, errors.New("Artifact Keeper response exceeds size limit")
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			requestsTotal.WithLabelValues(operation, "success").Inc()
			return responseBody, nil
		}
		lastErr = fmt.Errorf("Artifact Keeper request returned HTTP %d", resp.StatusCode)
		if resp.StatusCode < 500 || attempt == maxAttempts-1 {
			requestsTotal.WithLabelValues(operation, "http_error").Inc()
			return nil, lastErr
		}
	}

	requestsTotal.WithLabelValues(operation, "transport_error").Inc()
	return nil, lastErr
}
