package emulator

import (
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"time"

	"github.com/jarcoal/httpmock"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Emulator struct {
	router *chi.Mux

	images map[string][]*artifactregistrypb.DockerImage

	err error

	log zerolog.Logger

	server *httptest.Server
}

func New(log zerolog.Logger) *Emulator {
	e := &Emulator{
		router: chi.NewRouter(),
		images: map[string][]*artifactregistrypb.DockerImage{},
		log:    log,
	}

	e.routes()

	return e
}

func (e *Emulator) routes() {
	e.router.With(e.debug).Get("/v1/projects/{project}/locations/{location}/repositories/{name}/dockerImages", e.listImages)

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
	e.images = make(map[string][]*artifactregistrypb.DockerImage)
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

func (e *Emulator) listImages(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")
	location := chi.URLParam(r, "location")
	name := chi.URLParam(r, "name")

	parent := fmt.Sprintf("projects/%s/locations/%s/repositories/%s", project, location, name)

	resp := &artifactregistrypb.ListDockerImagesResponse{}

	if images, ok := e.images[parent]; ok {
		resp.DockerImages = images

		err := response(w, resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		return
	}

	http.Error(w, "images not found", http.StatusNotFound)
}

func (e *Emulator) SetImages(images map[string][]*artifactregistrypb.DockerImage) {
	e.images = images
}

func response(w http.ResponseWriter, v proto.Message) error {
	w.Header().Set("Content-Type", "application/json")

	bytes, err := protojson.Marshal(v)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)

	return err
}

// FIXME: might not need this, but I suspect that we might
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
