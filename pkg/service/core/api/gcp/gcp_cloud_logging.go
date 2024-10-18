package gcp

import (
	"context"

	"github.com/navikt/nada-backend/pkg/cloudlogging"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.CloudLoggingAPI = &cloudLoggingAPI{}

type cloudLoggingAPI struct {
	client cloudlogging.Operations
}

func (a *cloudLoggingAPI) ListLogEntries(ctx context.Context, project string, opts *service.ListLogEntriesOpts) ([]*service.LogEntry, error) {
	const op errs.Op = "cloudLoggingAPI.ListLogEntries"

	entries, err := a.client.ListLogEntries(ctx, project, &cloudlogging.ListLogEntriesOpts{
		ResourceNames: opts.ResourceNames,
		Filter:        opts.Filter,
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	var result []*service.LogEntry
	for _, e := range entries {
		result = append(result, &service.LogEntry{
			Timestamp: e.Timestamp,
			HTTPRequest: &service.HTTPRequest{
				UserAgent: e.HTTPRequest.UserAgent,
				URL:       e.HTTPRequest.URL,
				Method:    e.HTTPRequest.Method,
			},
		})
	}

	return result, nil
}

func NewCloudLoggingAPI(client cloudlogging.Operations) *cloudLoggingAPI {
	return &cloudLoggingAPI{client: client}
}
