package service

import (
	"context"
	"net/url"
	"time"
)

type CloudLoggingAPI interface {
	ListLogEntries(ctx context.Context, project string, opts *ListLogEntriesOpts) ([]*LogEntry, error)
}

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
