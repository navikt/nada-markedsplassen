package emulator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"google.golang.org/api/networksecurity/v1"
)

type Emulator struct {
	router *chi.Mux

	err error

	urlLists    map[string]*networksecurity.UrlList
	policyRules map[string]map[string]*networksecurity.GatewaySecurityPolicyRule

	log zerolog.Logger

	server *httptest.Server
}

func New(log zerolog.Logger) *Emulator {
	e := &Emulator{
		router:      chi.NewRouter(),
		urlLists:    make(map[string]*networksecurity.UrlList),
		policyRules: make(map[string]map[string]*networksecurity.GatewaySecurityPolicyRule),
		log:         log,
	}

	e.routes()

	return e
}

func (e *Emulator) routes() {
	e.router.With(e.debug).Post("/v1/projects/{project}/locations/{location}/urlLists", e.createURLList)
	e.router.With(e.debug).Get("/v1/projects/{project}/locations/{location}/urlLists/{id}", e.getURLList)
	e.router.With(e.debug).Patch("/v1/projects/{project}/locations/{location}/urlLists/{id}", e.updateURLList)
	e.router.With(e.debug).Delete("/v1/projects/{project}/locations/{location}/urlLists/{id}", e.deleteURLList)

	e.router.With(e.debug).Get("/v1/projects/{project}/locations/{location}/gatewaySecurityPolicies/{policy}/rules", e.listSecurityPolicyRule)
	e.router.With(e.debug).Post("/v1/projects/{project}/locations/{location}/gatewaySecurityPolicies/{policy}/rules", e.createSecurityPolicyRule)
	e.router.With(e.debug).Get("/v1/projects/{project}/locations/{location}/gatewaySecurityPolicies/{policy}/rules/{rule}", e.getSecurityPolicyRule)
	e.router.With(e.debug).Patch("/v1/projects/{project}/locations/{location}/gatewaySecurityPolicies/{policy}/rules/{rule}", e.updateSecurityPolicyRule)
	e.router.With(e.debug).Delete("/v1/projects/{project}/locations/{location}/gatewaySecurityPolicies/{policy}/rules/{rule}", e.deleteSecurityPolicyRule)

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
	e.log.Info().Msg("starting service account emulator")

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
	e.log.Info().Str("method", r.Method).Str("url", r.URL.String()).Msg("incoming request")
	e.router.ServeHTTP(w, r)
}

func (e *Emulator) SetURLList(project, location, id string, urlList *networksecurity.UrlList) {
	name := fmt.Sprintf("%s/%s/%s", project, location, id)
	e.urlLists[name] = urlList
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

func (e *Emulator) getURLList(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")
	location := chi.URLParam(r, "location")
	id := chi.URLParam(r, "id")

	name := fmt.Sprintf("%s/%s/%s", project, location, id)

	if sa, ok := e.urlLists[name]; ok {
		if err := json.NewEncoder(w).Encode(sa); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	http.Error(w, "service account not found", http.StatusNotFound)
}

func (e *Emulator) createURLList(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	var req *networksecurity.UrlList
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	project := chi.URLParam(r, "project")
	location := chi.URLParam(r, "location")
	id := r.URL.Query().Get("urlListId")
	if id == "" {
		id = filepath.Base(req.Name)
	}

	name := fmt.Sprintf("%s/%s/%s", project, location, id)

	if _, hasURLList := e.urlLists[name]; hasURLList {
		http.Error(w, "url list already exists", http.StatusConflict)

		return
	}

	e.urlLists[name] = req

	op := &networksecurity.Operation{
		Done: true,
	}

	if err := json.NewEncoder(w).Encode(op); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) updateURLList(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")
	location := chi.URLParam(r, "location")
	id := chi.URLParam(r, "id")

	name := fmt.Sprintf("%s/%s/%s", project, location, id)

	var req *networksecurity.UrlList
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if _, hasURLList := e.urlLists[name]; !hasURLList {
		http.Error(w, "url list does not exist", http.StatusNotFound)

		return
	}

	e.urlLists[name].UpdateTime = time.Now().String()
	e.urlLists[name].Values = req.Values

	op := &networksecurity.Operation{
		Done: true,
	}

	if err := json.NewEncoder(w).Encode(op); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) deleteURLList(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")
	location := chi.URLParam(r, "location")
	id := chi.URLParam(r, "id")

	name := fmt.Sprintf("%s/%s/%s", project, location, id)

	if _, ok := e.urlLists[name]; ok {
		delete(e.urlLists, name)
		w.WriteHeader(http.StatusNoContent)

		return
	}

	http.Error(w, "url list not found", http.StatusNotFound)
}

func (e *Emulator) listSecurityPolicyRule(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")
	location := chi.URLParam(r, "location")
	policy := chi.URLParam(r, "policy")

	policyName := fmt.Sprintf("%s/%s/%s", project, location, policy)

	if p, ok := e.policyRules[policyName]; ok {
		pr := make([]*networksecurity.GatewaySecurityPolicyRule, 0, len(p))
		for _, rule := range p {
			pr = append(pr, rule)
		}

		resp := &networksecurity.ListGatewaySecurityPolicyRulesResponse{
			GatewaySecurityPolicyRules: pr,
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	http.Error(w, "security policy not found", http.StatusNotFound)
}

func (e *Emulator) getSecurityPolicyRule(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")
	location := chi.URLParam(r, "location")
	policy := chi.URLParam(r, "policy")
	ruleName := chi.URLParam(r, "rule")

	policyName := fmt.Sprintf("%s/%s/%s", project, location, policy)

	if _, ok := e.policyRules[policyName]; !ok {
		http.Error(w, "security policy not found", http.StatusNotFound)

		return
	}

	if pr, ok := e.policyRules[policyName][ruleName]; ok {
		if err := json.NewEncoder(w).Encode(pr); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	http.Error(w, "security policy not found", http.StatusNotFound)
}

func (e *Emulator) createSecurityPolicyRule(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	policyRule := &networksecurity.GatewaySecurityPolicyRule{}
	if err := json.NewDecoder(r.Body).Decode(policyRule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	project := chi.URLParam(r, "project")
	location := chi.URLParam(r, "location")
	policy := chi.URLParam(r, "policy")
	ruleName := r.URL.Query().Get("gatewaySecurityPolicyRuleId")

	if ruleName == "" {
		ruleName = filepath.Base(policyRule.Name)
	}

	policyName := fmt.Sprintf("%s/%s/%s", project, location, policy)

	if _, exists := e.policyRules[policyName]; !exists {
		e.policyRules[policyName] = make(map[string]*networksecurity.GatewaySecurityPolicyRule)
	}

	if _, exists := e.policyRules[policyName][ruleName]; exists {
		http.Error(w, "policy rule already exists", http.StatusConflict)

		return
	}

	policyRule.CreateTime = time.Now().String()
	e.policyRules[policyName][ruleName] = policyRule

	op := &networksecurity.Operation{
		Done: true,
	}

	if err := json.NewEncoder(w).Encode(op); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (e *Emulator) updateSecurityPolicyRule(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")
	location := chi.URLParam(r, "location")
	policy := chi.URLParam(r, "policy")
	ruleName := chi.URLParam(r, "rule")
	policyName := fmt.Sprintf("%s/%s/%s", project, location, policy)

	updatedPolicyRule := &networksecurity.GatewaySecurityPolicyRule{}
	if err := json.NewDecoder(r.Body).Decode(updatedPolicyRule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	if _, exists := e.policyRules[policyName]; !exists {
		http.Error(w, "policy does not exists", http.StatusNotFound)

		return
	}

	if _, exists := e.policyRules[policyName][ruleName]; !exists {
		http.Error(w, "policy rule does not exists", http.StatusNotFound)

		return
	}

	e.policyRules[policyName][ruleName].UpdateTime = time.Now().String()
	e.policyRules[policyName][ruleName].SessionMatcher = updatedPolicyRule.SessionMatcher
	e.policyRules[policyName][ruleName].ApplicationMatcher = updatedPolicyRule.ApplicationMatcher

	op := &networksecurity.Operation{
		Done: true,
	}

	if err := json.NewEncoder(w).Encode(op); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (e *Emulator) deleteSecurityPolicyRule(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")
	location := chi.URLParam(r, "location")
	policy := chi.URLParam(r, "policy")
	ruleName := chi.URLParam(r, "rule")
	policyName := fmt.Sprintf("%s/%s/%s", project, location, policy)

	if _, ok := e.policyRules[policyName]; !ok {
		http.Error(w, "security policy not found", http.StatusNotFound)

		return
	}

	if _, exists := e.policyRules[policyName][ruleName]; exists {
		delete(e.policyRules[policyName], ruleName)
		w.WriteHeader(http.StatusNoContent)

		return
	}

	http.Error(w, "security policy rule not found", http.StatusNotFound)
}
