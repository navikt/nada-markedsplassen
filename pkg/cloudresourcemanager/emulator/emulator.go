package emulator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"time"

	"github.com/jarcoal/httpmock"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"google.golang.org/api/cloudresourcemanager/v1"
)

type Emulator struct {
	router *chi.Mux

	policies map[string]*cloudresourcemanager.Policy

	err error

	log zerolog.Logger

	server *httptest.Server
}

func New(log zerolog.Logger) *Emulator {
	e := &Emulator{
		router:   chi.NewRouter(),
		policies: map[string]*cloudresourcemanager.Policy{},
		log:      log,
	}

	e.routes()

	return e
}

func (e *Emulator) routes() {
	e.router.With(e.debug).Post("/v1/projects/{project}:getIamPolicy", e.getIamPolicy)
	e.router.With(e.debug).Post("/v1/projects/{project}:setIamPolicy", e.setIamPolicy)

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

func (e *Emulator) GetRouter() *chi.Mux {
	return e.router
}

func (e *Emulator) Run() string {
	e.log.Info().Msg("starting cloud resource manager emulator")

	e.server = httptest.NewServer(e)

	return e.server.URL
}

func (e *Emulator) Reset() {
	e.policies = make(map[string]*cloudresourcemanager.Policy)
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

func (e *Emulator) getIamPolicy(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")

	var req cloudresourcemanager.GetIamPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	fmt.Println("project", project)

	if policy, ok := e.policies[project]; ok {
		if err := json.NewEncoder(w).Encode(policy); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	http.Error(w, "policy not found", http.StatusNotFound)
}

func (e *Emulator) setIamPolicy(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")

	var req cloudresourcemanager.SetIamPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	e.policies[project] = req.Policy

	w.WriteHeader(http.StatusNoContent)
}

func (e *Emulator) SetPolicy(project string, policy *cloudresourcemanager.Policy) {
	e.policies[project] = policy
}

func (e *Emulator) GetPolicy(project string) *cloudresourcemanager.Policy {
	return e.policies[project]
}

func (e *Emulator) TagBindingPolicyClient(zones []string, statusCode int, log zerolog.Logger) *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout: 60 * time.Second,
		},
	}

	matcher := func(r *http.Request) bool {
		body, _ := io.ReadAll(r.Body)

		log.Info().Fields(map[string]interface{}{
			"method": r.Method,
			"url":    r.URL.String(),
			"body":   string(body),
		}).Msg("add_tag_binding_request")

		return true
	}

	httpmock.ActivateNonDefault(client)

	for _, z := range zones {
		httpmock.RegisterMatcherResponder(
			http.MethodPost,
			fmt.Sprintf("https://%s-cloudresourcemanager.googleapis.com/v3/tagBindings", z),
			httpmock.NewMatcher("log_request", matcher),
			httpmock.NewStringResponder(statusCode, ""),
		)
	}

	return client
}
