package datavarehus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/ords/dvh/dvh_dmo/dmo_ops_tnsnames.json", c.host), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("received status code %d from datavarehus: unable to parse response err: %w", res.StatusCode, err)
		}

		return nil, fmt.Errorf("received status code %d from datavarehus: error %s", res.StatusCode, string(resBody))
	}

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
	if c.token.Expired() {
		token, err := c.getAccessToken(ctx)
		if err != nil {
			return err
		}

		c.token = token
	}

	body := &Body{
		KeyID:     keyID,
		SignedJWT: signedJWT,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/ords/dvh/dvh_dmo/snart/kommer", c.host), bytes.NewReader(b))
	if err != nil {
		return err
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.AccessToken))

	res, err := c.httpClient.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()

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
		return nil, err
	}

	q := req.URL.Query()
	q.Add("grant_type", "client_credentials")
	req.URL.RawQuery = q.Encode()

	req.SetBasicAuth(c.clientID, c.clientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	token := &Token{}
	err = json.NewDecoder(resp.Body).Decode(token)
	if err != nil {
		return nil, err
	}
	fmt.Println(token)
	token.expires = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	fmt.Println(token)

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
