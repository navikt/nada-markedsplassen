package emulator

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"google.golang.org/api/iam/v1"
)

const (
	keyBits = 2048
)

type Emulator struct {
	router *chi.Mux

	serviceAccounts    map[string]*iam.ServiceAccount
	serviceAccountKeys map[string][]*iam.ServiceAccountKey
	policies           map[string]*iam.Policy

	err error

	log zerolog.Logger

	server *httptest.Server
}

func New(log zerolog.Logger) *Emulator {
	e := &Emulator{
		router:             chi.NewRouter(),
		serviceAccounts:    map[string]*iam.ServiceAccount{},
		serviceAccountKeys: map[string][]*iam.ServiceAccountKey{},
		policies:           map[string]*iam.Policy{},
		log:                log,
	}

	e.routes()

	return e
}

func (e *Emulator) routes() {
	e.router.With(e.debug).Post("/v1/projects/{project}/serviceAccounts", e.createServiceAccount)
	e.router.With(e.debug).Get("/v1/projects/{project}/serviceAccounts/{id}", e.getServiceAccount)
	e.router.With(e.debug).Delete("/v1/projects/{project}/serviceAccounts/{id}", e.deleteServiceAccount)
	e.router.With(e.debug).Get("/v1/projects/{project}/serviceAccounts", e.getServiceAccounts)
	e.router.With(e.debug).Get("/v1/projects/{project}/serviceAccounts/{id}/keys", e.listServiceAccountKeys)
	e.router.With(e.debug).Post("/v1/projects/{project}/serviceAccounts/{id}/keys", e.createServiceAccountKey)
	e.router.With(e.debug).Delete("/v1/projects/{project}/serviceAccounts/{id}/keys/{keyID}", e.deleteServiceAccountKey)
	e.router.With(e.debug).Post("/v1/projects/{project}/serviceAccounts/{id}:getIamPolicy", e.getIamPolicy)
	e.router.With(e.debug).Post("/v1/projects/{project}/serviceAccounts/{id}:setIamPolicy", e.setIamPolicy)

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
	e.serviceAccounts = make(map[string]*iam.ServiceAccount)
	e.serviceAccountKeys = make(map[string][]*iam.ServiceAccountKey)
	e.server.Close()
}

func (e *Emulator) GetServiceAccounts() map[string]*iam.ServiceAccount {
	return e.serviceAccounts
}

func (e *Emulator) GetServiceAccountKeys() map[string][]*iam.ServiceAccountKey {
	return e.serviceAccountKeys
}

func (e *Emulator) SetServiceAccount(name string, sa *iam.ServiceAccount) {
	e.serviceAccounts[name] = sa
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

func (e *Emulator) SetIamPolicy(name string, policy *iam.Policy) {
	e.policies[name] = policy
}

func (e *Emulator) GetPolicies() map[string]*iam.Policy {
	return e.policies
}

func (e *Emulator) setIamPolicy(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	id := chi.URLParam(r, "id")

	var req iam.SetIamPolicyRequest
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

	id := chi.URLParam(r, "id")

	if policy, ok := e.policies[id]; ok {
		if err := json.NewEncoder(w).Encode(policy); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	http.Error(w, "policy not found", http.StatusNotFound)
}

func (e *Emulator) SetServiceAccountKeys(name string, keys []*iam.ServiceAccountKey) {
	e.serviceAccountKeys[name] = keys
}

func (e *Emulator) deleteServiceAccountKey(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")
	id := chi.URLParam(r, "id")
	keyID := chi.URLParam(r, "keyID")

	name := serviceAccountNameFromEmail(project, id)

	if keys, ok := e.serviceAccountKeys[name]; ok {
		for i, key := range keys {
			if key.Name == serviceAccountKeyName(project, id, keyID) {
				e.serviceAccountKeys[name] = append(keys[:i], keys[i+1:]...)

				w.WriteHeader(http.StatusNoContent)

				return
			}
		}
	}

	http.Error(w, "service account key not found", http.StatusNotFound)
}

func generateRSAKey() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, &privateKey.PublicKey, nil
}

func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) ([]byte, error) {
	key, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: key,
	})

	return privateKeyPEM, nil
}

func encodePublicKeyToPEM(publicKey *rsa.PublicKey) string {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return ""
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM)
}

func (e *Emulator) createServiceAccountKey(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")
	id := chi.URLParam(r, "id")

	name := serviceAccountNameFromEmail(project, id)

	priv, pub, err := generateRSAKey()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	privateKey, err := encodePrivateKeyToPEM(priv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	if sa, ok := e.serviceAccounts[name]; ok {
		key := &iam.ServiceAccountKey{
			Name:           "projects/" + project + "/serviceAccounts/" + sa.Email + "/keys/" + strconv.Itoa(len(e.serviceAccountKeys[name])+1), // nolint: goconst
			PrivateKeyData: base64.StdEncoding.EncodeToString(privateKey),
			PublicKeyData:  encodePublicKeyToPEM(pub),
			KeyAlgorithm:   "KEY_ALG_RSA_2048",
			KeyOrigin:      "GOOGLE_PROVIDED",
			KeyType:        "USER_MANAGED",
		}

		if err := json.NewEncoder(w).Encode(key); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		key.PrivateKeyData = ""

		e.serviceAccountKeys[name] = append(e.serviceAccountKeys[name], key)

		return
	}

	http.Error(w, "service account not found", http.StatusNotFound)
}

func (e *Emulator) listServiceAccountKeys(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")
	id := chi.URLParam(r, "id")

	name := serviceAccountNameFromEmail(project, id)

	var response struct {
		Keys []*iam.ServiceAccountKey `json:"keys"`
	}

	if keys, ok := e.serviceAccountKeys[name]; ok {
		response.Keys = append(response.Keys, keys...)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) getServiceAccounts(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	var response struct {
		Accounts []*iam.ServiceAccount `json:"accounts"`
	}

	for _, sa := range e.serviceAccounts {
		response.Accounts = append(response.Accounts, sa)
	}

	slices.SortFunc(response.Accounts, func(i, j *iam.ServiceAccount) int {
		return strings.Compare(i.Email, j.Email)
	})

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) getServiceAccount(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")
	id := chi.URLParam(r, "id")

	name := serviceAccountNameFromEmail(project, id)

	if sa, ok := e.serviceAccounts[name]; ok {
		if err := json.NewEncoder(w).Encode(sa); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	http.Error(w, "service account not found", http.StatusNotFound)
}

func (e *Emulator) createServiceAccount(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")

	var req iam.CreateServiceAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	name := serviceAccountName(project, req.AccountId)

	if _, hasServiceAccount := e.serviceAccounts[name]; hasServiceAccount {
		http.Error(w, "service account already exists", http.StatusConflict)

		return
	}

	sa := &iam.ServiceAccount{
		Description: req.ServiceAccount.Description,
		DisplayName: req.ServiceAccount.DisplayName,
		Email:       emailFromAccountID(project, req.AccountId),
		Name:        name,
		ProjectId:   project,
		UniqueId:    strconv.Itoa(len(e.serviceAccounts) + 1),
	}

	e.serviceAccounts[name] = sa

	if err := json.NewEncoder(w).Encode(sa); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (e *Emulator) deleteServiceAccount(w http.ResponseWriter, r *http.Request) {
	if e.err != nil {
		http.Error(w, e.err.Error(), http.StatusInternalServerError)
		e.err = nil

		return
	}

	project := chi.URLParam(r, "project")
	id := chi.URLParam(r, "id")

	name := serviceAccountNameFromEmail(project, id)

	if _, ok := e.serviceAccounts[name]; ok {
		delete(e.serviceAccounts, name)
		w.WriteHeader(http.StatusNoContent)

		return
	}

	http.Error(w, "service account not found", http.StatusNotFound)
}

func serviceAccountNameFromEmail(project, email string) string {
	return "projects/" + project + "/serviceAccounts/" + email
}

func serviceAccountName(project, accountID string) string {
	return "projects/" + project + "/serviceAccounts/" + emailFromAccountID(project, accountID)
}

func emailFromAccountID(project, accountID string) string {
	return accountID + "@" + project + ".iam.gserviceaccount.com"
}

func serviceAccountKeyName(project, accountID, keyID string) string {
	return "projects/" + project + "/serviceAccounts/" + accountID + "/keys/" + keyID
}
