package emulator

import (
	"cloud.google.com/go/iam/credentials/apiv1/credentialspb"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
)

type Keypair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	KeyID      string
}

type Emulator struct {
	router  *chi.Mux
	log     zerolog.Logger
	server  *httptest.Server
	signers map[string]*Keypair
}

func New(log zerolog.Logger) *Emulator {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	e := &Emulator{
		router:  r,
		log:     log,
		signers: map[string]*Keypair{},
	}
	e.routes()
	return e
}

func (e *Emulator) AddSigner(email, publicKey, privateKey string) {
	priv, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		e.log.Error().Err(err).Msg("error parsing private key")
	}

	pub, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
	if err != nil {
		e.log.Error().Err(err).Msg("error parsing public key")
	}

	e.signers[email] = &Keypair{
		PrivateKey: priv,
		PublicKey:  pub,
		KeyID:      generateKeyID(pub),
	}
}

func (e *Emulator) routes() {
	e.router.With(e.debug).Post("/v1/projects/-/serviceAccounts/{email}:signJwt", e.signJWT)
	e.router.With(e.debug).NotFound(e.notFound)
}

func (e *Emulator) signJWT(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")

	keypair, ok := e.signers[email]
	if !ok {
		http.Error(w, "Signer not found", http.StatusNotFound)
		return
	}

	req := credentialspb.SignJwtRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to parse request", http.StatusBadRequest)
		return
	}

	var claims jwt.MapClaims
	if err := json.NewDecoder(strings.NewReader(req.Payload)).Decode(&claims); err != nil {
		http.Error(w, "Failed to parse claims", http.StatusBadRequest)
		return
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Sign the token with the private key
	signedToken, err := token.SignedString(keypair.PrivateKey)
	if err != nil {
		http.Error(w, "Failed to sign token", http.StatusInternalServerError)
		return
	}

	fmt.Println(keypair.KeyID)

	bytes, err := protojson.Marshal(&credentialspb.SignJwtResponse{
		SignedJwt: signedToken,
		KeyId:     keypair.KeyID,
	})
	if err != nil {
		e.log.Error().Err(err).Msg("error marshaling response")
	}

	_, err = w.Write(bytes)
	if err != nil {
		e.log.Error().Err(err).Msg("error writing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func generateKeyID(pub *rsa.PublicKey) string {
	pubASN1, _ := x509.MarshalPKIXPublicKey(pub)
	hash := sha1.Sum(pubASN1)
	return fmt.Sprintf("%x", hash)
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
