package emulator

import (
	"cloud.google.com/go/kms/apiv1/kmspb"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
)

type Emulator struct {
	router        *chi.Mux
	log           zerolog.Logger
	server        *httptest.Server
	symmetricKeys map[string][]byte
}

func New(log zerolog.Logger) *Emulator {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	e := &Emulator{
		router:        r,
		log:           log,
		symmetricKeys: map[string][]byte{},
	}

	e.routes()

	return e
}

func (e *Emulator) routes() {
	e.router.With(e.debug).Post("/v1/projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{keyName}:encrypt", e.encrypt)
	e.router.With(e.debug).Post("/v1/projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{keyName}:decrypt", e.decrypt)
	e.router.With(e.debug).NotFound(e.notFound)
}

func (e *Emulator) AddSymmetricKey(project, location, keyRing, keyName string, key []byte) {
	e.symmetricKeys[fmt.Sprintf("%s-%s-%s-%s", project, location, keyRing, keyName)] = key
}

func (e *Emulator) encrypt(w http.ResponseWriter, r *http.Request) {
	project := chi.URLParam(r, "project")
	location := chi.URLParam(r, "location")
	keyRing := chi.URLParam(r, "keyRing")
	keyName := chi.URLParam(r, "keyName")

	keyPath := fmt.Sprintf("%s-%s-%s-%s", project, location, keyRing, keyName)

	symmetricKey, ok := e.symmetricKeys[keyPath]
	if !ok {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	req := &kmspb.EncryptRequest{}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := protojson.Unmarshal(data, req); err != nil {
		http.Error(w, "Failed to parse request", http.StatusBadRequest)
		return
	}

	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		http.Error(w, fmt.Errorf("creating cipher: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	ciphertext := make([]byte, aes.BlockSize+len(req.Plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		http.Error(w, "Failed to generate IV", http.StatusInternalServerError)
		return
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], req.Plaintext)

	resp := &kmspb.EncryptResponse{
		Name:       req.Name,
		Ciphertext: ciphertext,
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

func (e *Emulator) decrypt(w http.ResponseWriter, r *http.Request) {
	project := chi.URLParam(r, "project")
	location := chi.URLParam(r, "location")
	keyRing := chi.URLParam(r, "keyRing")
	keyName := chi.URLParam(r, "keyName")

	keyPath := fmt.Sprintf("%s-%s-%s-%s", project, location, keyRing, keyName)

	symmetricKey, ok := e.symmetricKeys[keyPath]
	if !ok {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	req := &kmspb.DecryptRequest{}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	unm := protojson.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}
	if err := unm.Unmarshal(data, req); err != nil {
		http.Error(w, "Failed to parse request", http.StatusBadRequest)
		return
	}

	if len(req.Ciphertext) < aes.BlockSize {
		http.Error(w, "ciphertext too short", http.StatusBadRequest)
		return
	}

	// Extract IV
	iv := req.Ciphertext[:aes.BlockSize]
	ciphertext := req.Ciphertext[aes.BlockSize:]

	// Decrypt
	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		http.Error(w, "Failed to create cipher", http.StatusInternalServerError)
		return
	}

	stream := cipher.NewCFBDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	resp := &kmspb.DecryptResponse{
		Plaintext: plaintext,
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

func (e *Emulator) notFound(w http.ResponseWriter, r *http.Request) {
	request, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	e.log.Error().Str("request", string(request)).Msg("not found")
	http.Error(w, "not found", http.StatusNotFound)
}

func (e *Emulator) Run() string {
	e.server = httptest.NewServer(e.router)
	return e.server.URL
}

func (e *Emulator) Reset() {
	e.server.Close()
}
