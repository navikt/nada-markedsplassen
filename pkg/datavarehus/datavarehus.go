package datavarehus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

var (
	datavarehusRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "nada_backend",
		Subsystem: "datavarehus",
		Name:      "requests_total",
		Help:      "Total number of requests to datavarehus API",
	}, []string{"method", "endpoint", "status_code"})

	datavarehusRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "nada_backend",
		Subsystem: "datavarehus",
		Name:      "request_duration_seconds",
		Help:      "Duration of requests to datavarehus API",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "endpoint"})

	datavarehusTNSNamesHealthy = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nada_backend",
		Subsystem: "datavarehus",
		Name:      "tnsnames_healthy",
		Help:      "Whether the dmo_ops_tnsnames.json endpoint returns 200 OK (1 = healthy, 0 = unhealthy)",
	})
)

// Collectors returns the prometheus collectors for datavarehus metrics.
func Collectors() []prometheus.Collector {
	return []prometheus.Collector{
		datavarehusRequestsTotal,
		datavarehusRequestDuration,
		datavarehusTNSNamesHealthy,
	}
}

type Operations interface {
	GetTNSNames(ctx context.Context) ([]TNSName, error)
	SendJWT(ctx context.Context, keyID, signedJWT string) error
}

type Client struct {
	httpClient *http.Client

	host         string
	clientID     string
	clientSecret string

	token *Token
}

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`

	expires time.Time
}

func (t *Token) Expired() bool {
	return time.Now().After(t.expires)
}

type Body struct {
	KeyID     string `json:"key_id"`
	SignedJWT string `json:"signed_jwt"`
}

type TNSNames struct {
	Items []TNSName `json:"items"`

	HasMore bool `json:"hasMore"`

	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`

	Links []Link `json:"links"`
}

type TNSName struct {
	TnsName     string `json:"tns_name"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Host        string `json:"host"`
	Port        string `json:"port"`
	ServiceName string `json:"service_name"`
}

type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

func (c *Client) GetTNSNames(ctx context.Context) ([]TNSName, error) {
	start := time.Now()
	endpoint := "dmo_ops_tnsnames.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/ords/dvh/dvh_dmo/dmo_ops_tnsnames.json", c.host), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		datavarehusRequestsTotal.WithLabelValues(http.MethodGet, endpoint, "error").Inc()
		datavarehusTNSNamesHealthy.Set(0)
		return nil, err
	}

	datavarehusRequestDuration.WithLabelValues(http.MethodGet, endpoint).Observe(time.Since(start).Seconds())
	datavarehusRequestsTotal.WithLabelValues(http.MethodGet, endpoint, strconv.Itoa(res.StatusCode)).Inc()

	if res.StatusCode != http.StatusOK {
		datavarehusTNSNamesHealthy.Set(0)
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("received status code %d from datavarehus: unable to parse response err: %w", res.StatusCode, err)
		}

		return nil, fmt.Errorf("received status code %d from datavarehus: error %s", res.StatusCode, string(resBody))
	}

	datavarehusTNSNamesHealthy.Set(1)
	defer res.Body.Close()

	n := &TNSNames{}
	err = json.NewDecoder(res.Body).Decode(n)
	if err != nil {
		return nil, err
	}

	if n.HasMore {
		return nil, fmt.Errorf("more items available than returned, please implement pagination")
	}

	return n.Items, nil
}

func (c *Client) SendJWT(ctx context.Context, keyID, signedJWT string) error {
	start := time.Now()
	endpoint := "wli_rest/notify"

	if c.token.Expired() {
		token, err := c.getAccessToken(ctx)
		if err != nil {
			return fmt.Errorf("renewing access token: %w", err)
		}

		c.token = token
	}

	body := &Body{
		KeyID:     keyID,
		SignedJWT: signedJWT,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("unable to marshal jwt body: %w", err)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/ords/dvh/dvh_dmo/wli_rest/notify", c.host), bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.AccessToken))

	res, err := c.httpClient.Do(r)
	if err != nil {
		datavarehusRequestsTotal.WithLabelValues(http.MethodPost, endpoint, "error").Inc()
		return fmt.Errorf("unable to send jwt to datavarehus: %w", err)
	}
	defer res.Body.Close()

	datavarehusRequestDuration.WithLabelValues(http.MethodPost, endpoint).Observe(time.Since(start).Seconds())
	datavarehusRequestsTotal.WithLabelValues(http.MethodPost, endpoint, strconv.Itoa(res.StatusCode)).Inc()

	if res.StatusCode > 299 {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("received status code %d from datavarehus: unable to parse response err: %w", res.StatusCode, err)
		}

		return fmt.Errorf("received status code %d from datavarehus: error %s", res.StatusCode, string(resBody))
	}

	return nil
}

func (c *Client) getAccessToken(ctx context.Context) (*Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/ords/dvh/oauth/token", c.host), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("grant_type", "client_credentials")
	req.URL.RawQuery = q.Encode()

	req.SetBasicAuth(c.clientID, c.clientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to get access token: %w", err)
	}
	defer resp.Body.Close()

	token := &Token{}
	err = json.NewDecoder(resp.Body).Decode(token)
	if err != nil {
		return nil, fmt.Errorf("unable to decode access token: %w", err)
	}

	token.expires = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	return token, nil
}

func New(host, clientID, clientSecret string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		host:         host,
		clientID:     clientID,
		clientSecret: clientSecret,
		token:        &Token{},
	}
}

// RunHealthChecker starts a background goroutine that checks if dmo_ops_tnsnames.json
// returns 200 OK every 30 seconds and updates the datavarehusTNSNamesHealthy metric.
func (c *Client) RunHealthChecker(ctx context.Context, log zerolog.Logger) {
	log.Debug().Str("host", c.host).Msg("datavarehus health checker started")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Run immediately on start
	c.checkHealth(ctx, log)

	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("datavarehus health checker stopped")
			return
		case <-ticker.C:
			c.checkHealth(ctx, log)
		}
	}
}

func (c *Client) checkHealth(ctx context.Context, log zerolog.Logger) {
	url := fmt.Sprintf("%s/ords/dvh/dvh_dmo/dmo_ops_tnsnames.json", c.host)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("datavarehus health check: failed to create request")
		datavarehusTNSNamesHealthy.Set(0)
		return
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("datavarehus health check: request failed")
		datavarehusTNSNamesHealthy.Set(0)
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		datavarehusTNSNamesHealthy.Set(1)
		log.Debug().Str("url", url).Msg("datavarehus health check: healthy")
	} else {
		datavarehusTNSNamesHealthy.Set(0)
		log.Warn().Str("url", url).Int("status_code", res.StatusCode).Msg("datavarehus health check: unhealthy")
	}
}
