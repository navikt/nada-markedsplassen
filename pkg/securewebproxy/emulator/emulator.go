package emulator

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"google.golang.org/api/networksecurity/v1"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"path/filepath"
)

const (
	keyBits = 2048
)

type Emulator struct {
	router *chi.Mux

	err error

	urlLists map[string]*networksecurity.UrlList

	log zerolog.Logger

	server *httptest.Server
}

func New(log zerolog.Logger) *Emulator {
	e := &Emulator{
		router:   chi.NewRouter(),
		urlLists: make(map[string]*networksecurity.UrlList),
		log:      log,
	}

	e.routes()

	return e
}

func (e *Emulator) routes() {
	e.router.With(e.debug).Post("/v1/projects/{project}/locations/{location}/urlLists", e.createURLList)
	e.router.With(e.debug).Get("/v1/projects/{project}/locations/{location}/urlLists/{id}", e.getURLList)
	e.router.With(e.debug).Delete("/v1/projects/{project}/locations/{location}/urlLists/{id}", e.deleteURLList)

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

	if err := json.NewEncoder(w).Encode(req); err != nil {
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
