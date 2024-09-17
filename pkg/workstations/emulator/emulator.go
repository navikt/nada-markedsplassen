package emulator

import (
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"cloud.google.com/go/workstations/apiv1/workstationspb"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/sanity-io/litter"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"time"
)

type Emulator struct {
	router *chi.Mux

	err error

	log zerolog.Logger

	server *httptest.Server
}

func New(log zerolog.Logger) *Emulator {
	e := &Emulator{
		router: chi.NewRouter(),
		log:    log,
	}

	e.routes()

	return e
}

func (e *Emulator) routes() {
	e.router.With(e.debug).Post("/v1/projects/{project}/locations/{location}/workstationClusters/{cluster}/workstationConfigs", e.CreateWorkstationConfig)
	e.router.With(e.debug).Get("/v1/projects/{project}/locations/{location}/workstationClusters/{cluster}/workstationConfigs/{name}", e.CreateWorkstationConfig)

	e.router.NotFound(e.notFound)
}

func (e *Emulator) debug(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		e.log.Debug().Str("request", string(request)).Msg("request")

		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)

		response, err := httputil.DumpResponse(rec.Result(), true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		e.log.Debug().Str("response", string(response)).Msg("response")

		for k, v := range rec.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(rec.Code)
		w.Write(rec.Body.Bytes())
	})
}

func (e *Emulator) CreateWorkstationConfig(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	req := &workstationspb.WorkstationConfig{}

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		e.log.Error().Err(err).Msg("error reading request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	m := protojson.UnmarshalOptions{AllowPartial: true}
	err = m.Unmarshal(bytes, req)
	if err != nil {
		e.log.Error().Err(err).Msg("error encoding request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	now := time.Now()

	req.CreateTime = &timestamppb.Timestamp{
		Seconds: now.Unix(),
	}

	fmt.Print(litter.Sdump(req))

	into := &anypb.Any{}
	err = anypb.MarshalFrom(into, req, proto.MarshalOptions{})

	op := &longrunningpb.Operation{
		Name:     "/v1/projects/x/locations/y/workstationClusters/z/workstationConfigs/hey",
		Metadata: nil,
		Done:     true,
		Result: &longrunningpb.Operation_Response{
			Response: into,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	bytes, err = protojson.Marshal(op)
	if err != nil {
		e.log.Error().Err(err).Msg("error encoding response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	_, err = w.Write(bytes)
	if err != nil {
		e.log.Error().Err(err).Msg("error encoding response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) GetRouter() *chi.Mux {
	return e.router
}

func (e *Emulator) Run() string {
	e.log.Info().Msg("starting cloud workstation emulator")

	e.server = httptest.NewServer(e)

	return e.server.URL
}

func (e *Emulator) Reset() {
	e.server.Close()
}

func (e *Emulator) SetError(err error) {
	e.err = err
}

func (e *Emulator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.router.ServeHTTP(w, r)
}

func (e *Emulator) notFound(w http.ResponseWriter, r *http.Request) {
	request, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	e.log.Warn().Str("request", string(request)).Msg("not found")

	http.Error(w, "not found", http.StatusNotFound)
}

func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}
