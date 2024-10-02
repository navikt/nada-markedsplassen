package cloudlogging_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	logpb "cloud.google.com/go/logging/apiv2/loggingpb"

	"github.com/navikt/nada-backend/pkg/cloudlogging"
	"github.com/navikt/nada-backend/pkg/cloudlogging/emulator"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	ltype "google.golang.org/genproto/googleapis/logging/type"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestClient_ListLogEntries(t *testing.T) {
	now := timestamppb.Now()

	testCases := []struct {
		name      string
		project   string
		opts      *cloudlogging.ListLogEntriesOpts
		entries   []*logpb.LogEntry
		expect    []*cloudlogging.LogEntry
		expectErr bool
	}{
		{
			name:    "no entries",
			project: "test",
			opts:    &cloudlogging.ListLogEntriesOpts{},
			expect:  nil,
		},
		{
			name:    "one entry",
			project: "test",
			opts:    &cloudlogging.ListLogEntriesOpts{},
			entries: []*logpb.LogEntry{
				{
					Timestamp: now,
					HttpRequest: &ltype.HttpRequest{
						RequestMethod: http.MethodGet,
						RequestUrl:    "https://github.com/navikt",
						UserAgent:     "Chrome 91",
					},
				},
			},
			expect: []*cloudlogging.LogEntry{
				{
					HTTPRequest: &cloudlogging.HTTPRequest{
						Method: http.MethodGet,
						URL: func() *url.URL {
							u, err := url.Parse("https://github.com/navikt")
							require.NoError(t, err)

							return u
						}(),
						UserAgent: "Chrome 91",
					},
					Timestamp: now.AsTime(),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log := zerolog.New(zerolog.NewConsoleWriter())
			ctx := context.Background()

			e := emulator.NewEmulator(log)
			addr, err := e.Start()
			require.NoError(t, err)

			if len(tc.entries) > 0 {
				e.AppendLogEntries(tc.entries)
			}

			c := cloudlogging.NewClient(addr, true)
			got, err := c.ListLogEntries(ctx, tc.project, tc.opts)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expect, got)
			}
		})
	}
}
