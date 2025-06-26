package emulator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"time"

	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"cloud.google.com/go/iam/apiv1/iampb"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/jarcoal/httpmock"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

type Emulator struct {
	router *chi.Mux

	images      map[string][]*artifactregistrypb.DockerImage
	policies    map[string]*iampb.Policy
	attachments []*artifactregistrypb.Attachment

	err error

	log zerolog.Logger

	server *httptest.Server
}

func New(log zerolog.Logger) *Emulator {
	e := &Emulator{
		router:   chi.NewRouter(),
		images:   map[string][]*artifactregistrypb.DockerImage{},
		policies: map[string]*iampb.Policy{},
		log:      log,
	}

	e.routes()

	return e
}

func (e *Emulator) routes() {
	e.router.With(e.debug).Get("/v1/projects/{project}/locations/{location}/repositories/{name}/dockerImages", e.listImages)
	e.router.With(e.debug).Get("/v1/projects/{project}/locations/{location}/repositories/{name}/packages/{package}/tags/{tag}", e.getTag)
	e.router.With(e.debug).Get("/v1/projects/{project}/locations/{location}/repositories/{name}/attachments", e.listAttachments)
	e.router.With(e.debug).Get("/v1/projects/{project}/locations/{location}/repositories/{name}:getIamPolicy", e.getIamPolicy)
	e.router.With(e.debug).Post("/v1/projects/{project}/locations/{location}/repositories/{name}:setIamPolicy", e.setIamPolicy)
	e.router.With(e.debug).Get("/v1/projects/{project}/locations/{location}/repositories/{name}/files/{id}", e.downloadAttachment)
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
	e.log.Info().Msg("starting artifact registry emulator")

	e.server = httptest.NewServer(e)

	return e.server.URL
}

func (e *Emulator) Reset() {
	e.images = make(map[string][]*artifactregistrypb.DockerImage)
	e.policies = make(map[string]*iampb.Policy)
	e.server.Close()
}

func (e *Emulator) SetError(err error) {
	e.err = err
}

func (e *Emulator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.router.ServeHTTP(w, r)
}

func (e *Emulator) SetImageAttachment() {
	e.attachments = append(e.attachments, &artifactregistrypb.Attachment{
		Name: "projects/test/locations/europe-north1/repositories/knast-images/attachments/63ea9b26-b4fa-4f78-99a0-d833464bdb14",
		Type: "application/vnd.knast.container.config",
		Files: []string{
			"projects/test/locations/europe-north1/repositories/knast-images/files/sha256:ba01b3a4cd1eefe9275f0ef7e0882fc41b5a6fdf9a41420d4c9e568bfa8dbc3c",
		},
	})
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

func (e *Emulator) SetIamPolicy(name string, policy *iampb.Policy) {
	e.policies[name] = policy
}

func (e *Emulator) GetPolicies() map[string]*iampb.Policy {
	return e.policies
}

func (e *Emulator) setIamPolicy(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	id := chi.URLParam(r, "name")

	var req iampb.SetIamPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	e.policies[id] = req.Policy

	if err := json.NewEncoder(w).Encode(req.Policy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) getIamPolicy(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	id := chi.URLParam(r, "name")

	if policy, ok := e.policies[id]; ok {
		if err := json.NewEncoder(w).Encode(policy); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	http.Error(w, "policy not found", http.StatusNotFound)
}

func (e *Emulator) getTag(w http.ResponseWriter, r *http.Request) {
	project := chi.URLParam(r, "project")
	location := chi.URLParam(r, "location")
	repo := chi.URLParam(r, "repository")
	name := chi.URLParam(r, "name")
	tag := chi.URLParam(r, "tag")

	resp := &artifactregistrypb.Tag{
		Name:    fmt.Sprintf("projects/%s/locations/%s/repositories/%s/packages/%s/tags/%s", project, location, name, repo, tag),
		Version: fmt.Sprintf("projects/%s/locations/%s/repositories/%s/packages/%s/versions/sha256:26a04640990301944d8444873a848754c15c3e82f2348431f7ee224e6d006a6c", project, location, name, repo),
	}
	err := response(w, resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (e *Emulator) listAttachments(w http.ResponseWriter, r *http.Request) {
	resp := &artifactregistrypb.ListAttachmentsResponse{
		Attachments: e.attachments,
	}

	err := response(w, resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (e *Emulator) downloadAttachment(w http.ResponseWriter, r *http.Request) {
	config := map[string]any{
		"extra_readiness_checks": []map[string]any{
			{
				"path": "/healthcheck",
				"port": 80,
			},
		},
	}

	w.Header().Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

func (e *Emulator) ArtifactRegistryImageInspecter(uri string, statusCode int, log zerolog.Logger) *http.Client {
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
		}).Msg("inspect_image_request")

		return true
	}

	httpmock.ActivateNonDefault(client)

	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		fmt.Sprintf("https://%s", uri),
		httpmock.NewMatcher("log_request", matcher),
		httpmock.NewStringResponder(statusCode, ""),
	)

	return client
}
