package leaderelection

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	clientTimeout = 5 * time.Second
	numRetries    = 3
)

func IsLeader(ctx context.Context) (bool, error) {
	electorPath := os.Getenv("ELECTOR_PATH")
	if electorPath == "" {
		// local development
		return true, nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		return false, err
	}

	leader, err := getLeader(ctx, electorPath)
	if err != nil {
		return false, err
	}

	return hostname == leader, nil
}

func getLeader(ctx context.Context, electorPath string) (string, error) {
	resp, err := electorRequestWithRetry(ctx, electorPath, numRetries)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var electorResponse struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal(bodyBytes, &electorResponse); err != nil {
		return "", err
	}

	return electorResponse.Name, nil
}

func electorRequestWithRetry(ctx context.Context, electorPath string, numRetries int) (*http.Response, error) {
	client := http.Client{
		Timeout: clientTimeout,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+electorPath, nil)
	if err != nil {
		return nil, err
	}

	for i := 1; i <= numRetries; i++ {
		resp, err := client.Do(req)
		if err == nil {
			return resp, nil
		}

		time.Sleep(time.Second * time.Duration(i))
	}

	return nil, fmt.Errorf("no response from elector container after %v retries", numRetries)
}
