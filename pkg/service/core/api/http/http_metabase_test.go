package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	mbhttp "github.com/navikt/nada-backend/pkg/service/core/api/http"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMetabaseTestServer creates a minimal httptest server that handles the
// two endpoints exercised by the Metabase client in these tests:
//   - POST /session  — returns a fake session token
//   - PUT  /table    — returns 200 OK
//
// If closeAfterResponse is true the handler hijacks the underlying TCP
// connection and writes a complete HTTP response manually before closing,
// simulating Metabase's Jetty server silently closing keep-alive connections
// after its idle timeout.
func newMetabaseTestServer(t *testing.T, closeAfterResponse bool) (srv *httptest.Server, newConns *atomic.Int32) {
	t.Helper()

	newConns = new(atomic.Int32)

	srv = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !closeAfterResponse {
			switch r.URL.Path {
			case "/session":
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{"id": "test-session"})
			default:
				w.WriteHeader(http.StatusOK)
			}
			return
		}

		// When simulating Jetty's connection-close behaviour we must write a
		// complete, valid HTTP response manually. w.WriteHeader() only buffers
		// the status code inside the ResponseWriter — the bytes are not flushed
		// to the underlying connection until the handler returns or the first
		// w.Write() call, so hijacking immediately after WriteHeader would
		// leave the client with an incomplete response and an unexpected EOF.
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server does not support hijacking")
			return
		}

		conn, buf, _ := hj.Hijack()
		defer conn.Close()

		switch r.URL.Path {
		case "/session":
			body := `{"id":"test-session"}`
			_, _ = fmt.Fprintf(buf, "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
		default:
			_, _ = buf.WriteString("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
		}

		_ = buf.Flush()
		// Close without sending Connection: close, so the client does not know
		// the connection is gone until it tries to reuse it.
	}))

	srv.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			newConns.Add(1)
		}
	}

	srv.Start()
	t.Cleanup(srv.Close)

	return srv, newConns
}

// TestMetabaseClientOpensNewConnectionPerRequest asserts that every request
// uses a fresh TCP connection. This is the property that prevents stale
// keep-alive connections from causing EOF or "server closed idle connection"
// errors on non-idempotent requests (PUT/POST/DELETE), which Go does not
// auto-retry.
//
// With DisableKeepAlives: true the connection count must equal the number of
// requests. Without it, connections would be reused and the count would be
// lower, breaking the guarantee.
func TestMetabaseClientOpensNewConnectionPerRequest(t *testing.T) {
	t.Parallel()

	const nRequests = 5

	// Normal server — connections are never closed by the server so Go would
	// happily reuse them if keep-alives were enabled on the client.
	srv, newConns := newMetabaseTestServer(t, false)

	client := mbhttp.NewMetabaseHTTP(srv.URL, "user", "pass", "", false, false, zerolog.Nop())

	for range nRequests {
		require.NoError(t, client.HideTables(context.Background(), []int{1, 2, 3}))
	}

	// The first HideTables call establishes a session (POST /session) before
	// issuing the PUT /table request, so the expected connection count is
	// nRequests + 1.
	assert.Equal(t, nRequests+1, int(newConns.Load()))
}

// TestMetabaseClientSucceedsWhenServerClosesConnections asserts that all
// requests succeed even when the server forcefully closes the TCP connection
// after every response — without sending a Connection: close header. This
// replicates Metabase's Jetty server behaviour that previously caused
// intermittent failures on HideTables and ShowTables in periodic workers.
func TestMetabaseClientSucceedsWhenServerClosesConnections(t *testing.T) {
	t.Parallel()

	const nRequests = 5

	srv, newConns := newMetabaseTestServer(t, true)

	client := mbhttp.NewMetabaseHTTP(srv.URL, "user", "pass", "", false, false, zerolog.Nop())

	for range nRequests {
		require.NoError(t, client.HideTables(context.Background(), []int{1, 2, 3}))
	}

	// Every request must have used its own connection because the previous one
	// was closed by the server.
	assert.Equal(t, nRequests+1, int(newConns.Load()))
}
