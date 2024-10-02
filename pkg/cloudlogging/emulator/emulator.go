package emulator

import (
	"context"
	"net"

	logpb "cloud.google.com/go/logging/apiv2/loggingpb"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Emulator struct {
	server   *grpc.Server
	listener net.Listener
	log      zerolog.Logger

	logpb.LoggingServiceV2Server

	entries []*logpb.LogEntry
}

func NewEmulator(log zerolog.Logger) *Emulator {
	return &Emulator{
		log: log,
	}
}

func (e *Emulator) Start() (string, error) {
	var err error
	e.listener, err = net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	e.server = grpc.NewServer()
	logpb.RegisterLoggingServiceV2Server(e.server, e)
	reflection.Register(e.server)

	go func() {
		if err := e.server.Serve(e.listener); err != nil {
			panic(err)
		}
	}()

	return e.listener.Addr().String(), nil
}

func (e *Emulator) Stop() {
	e.server.GracefulStop()
	e.listener.Close()
}

func (e *Emulator) Reset() {
	e.entries = nil
}

func (e *Emulator) AppendLogEntries(entries []*logpb.LogEntry) {
	e.log.Info().Msgf("appending %d log entries", len(entries))

	e.entries = append(e.entries, entries...)
}

func (e *Emulator) ListLogEntries(_ context.Context, req *logpb.ListLogEntriesRequest) (*logpb.ListLogEntriesResponse, error) {
	e.log.Info().Fields(map[string]interface{}{
		"resource_names": req.GetResourceNames(),
		"filter":         req.GetFilter(),
		"order_by":       req.GetOrderBy(),
		"page_size":      req.GetPageSize(),
	}).Msg("listing log entries")

	return &logpb.ListLogEntriesResponse{
		Entries: e.entries,
	}, nil
}
