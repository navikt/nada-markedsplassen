package emulator

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strconv"

	"cloud.google.com/go/compute/apiv1/computepb"
	"golang.org/x/exp/rand"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
)

type Emulator struct {
	router                *chi.Mux
	err                   error
	log                   zerolog.Logger
	server                *httptest.Server
	storeInstances        map[string][]*computepb.Instance
	storeFirewallPolicies map[string]*computepb.FirewallPolicy
}

func New(log zerolog.Logger) *Emulator {
	e := &Emulator{
		router:         chi.NewRouter(),
		log:            log,
		storeInstances: map[string][]*computepb.Instance{},
	}
	e.routes()
	return e
}

func (e *Emulator) SetFirewallPolicies(policies map[string]*computepb.FirewallPolicy) {
	e.storeFirewallPolicies = policies
}

func (e *Emulator) SetInstances(instances map[string][]*computepb.Instance) {
	e.storeInstances = instances
}

func (e *Emulator) routes() {
	e.router.With(e.debug).Get("/compute/v1/projects/{project}/zones/{zone}/instances", e.listInstances)
	e.router.With(e.debug).Get("/compute/v1/projects/{project}/regions/{region}/firewallPolicies/{name}", e.getFirewallPolicy)
	e.router.With(e.debug).NotFound(e.notFound)
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

func (e *Emulator) listInstances(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil
		return
	}

	zone := chi.URLParam(r, "zone")

	id := strconv.Itoa(rand.Int())

	instances, ok := e.storeInstances[zone]
	if !ok {
		instances = []*computepb.Instance{}
	}

	resp := &computepb.InstanceList{
		Id:    &id,
		Items: instances,
	}

	bytes, err := protojson.Marshal(resp)
	if err != nil {
		e.log.Error().Err(err).Msg("error marshaling response")
	}

	_, err = w.Write(bytes)
	if err != nil {
		e.log.Error().Err(err).Msg("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) getFirewallPolicy(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil
		return
	}

	project := chi.URLParam(r, "project")
	region := chi.URLParam(r, "region")
	name := chi.URLParam(r, "name")

	id := uint64(rand.Int())

	policy, ok := e.storeFirewallPolicies[fmt.Sprintf("%s-%s-%s", project, region, name)]
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	policy.Id = &id

	bytes, err := protojson.Marshal(policy)
	if err != nil {
		e.log.Error().Err(err).Msg("error marshaling response")
	}

	_, err = w.Write(bytes)
	if err != nil {
		e.log.Error().Err(err).Msg("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) GetRouter() *chi.Mux {
	return e.router
}

func (e *Emulator) Run() string {
	e.log.Info().Msg("starting compute engine emulator")
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
	e.log.Error().Str("request", string(request)).Msg("not found")
	http.Error(w, "not found", http.StatusNotFound)
}
