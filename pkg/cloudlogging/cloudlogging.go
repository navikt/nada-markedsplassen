package cloudlogging

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"cloud.google.com/go/logging/logadmin"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var (
	ErrNotExist = errors.New("not exist")
	ErrExist    = errors.New("already exists")
)

var _ Operations = &Client{}

type ListLogEntriesOpts struct {
	ResourceNames []string
	Filter        string
}

type HTTPRequest struct {
	UserAgent string
	URL       *url.URL
	Method    string
}

type LogEntry struct {
	HTTPRequest *HTTPRequest
	Timestamp   time.Time
}

type Operations interface {
	ListLogEntries(ctx context.Context, project string, opts *ListLogEntriesOpts) ([]*LogEntry, error)
}

type Client struct {
	apiEndpoint string
	disableAuth bool
}

func (c *Client) ListLogEntries(ctx context.Context, project string, opts *ListLogEntriesOpts) ([]*LogEntry, error) {
	client, err := c.newClient(ctx, project)
	if err != nil {
		return nil, err
	}

	iter := client.Entries(ctx,
		logadmin.ResourceNames(opts.ResourceNames),
		logadmin.Filter(opts.Filter),
		logadmin.NewestFirst(),
	)

	entries := []*LogEntry{}
	for {
		entry, err := iter.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		entries = append(entries, &LogEntry{
			HTTPRequest: &HTTPRequest{
				UserAgent: entry.HTTPRequest.Request.UserAgent(),
				URL:       entry.HTTPRequest.Request.URL,
				Method:    entry.HTTPRequest.Request.Method,
			},
			Timestamp: entry.Timestamp,
		})
	}

	return entries, nil
}

func (c *Client) newClient(ctx context.Context, project string) (*logadmin.Client, error) {
	var options []option.ClientOption

	if c.apiEndpoint != "" {
		options = append(options, option.WithEndpoint(c.apiEndpoint))
	}

	if c.disableAuth {
		options = append(options,
			option.WithoutAuthentication(),
		)
	}

	client, err := logadmin.NewClient(ctx, project, options...)
	if err != nil {
		return nil, fmt.Errorf("creating workstations client: %w", err)
	}

	return client, nil
}
