package auth

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/navikt/nada-backend/pkg/database/gensql"
)

type MiddlewareHandler func(http.Handler) http.Handler

type contextKey int

const ContextUserKey contextKey = 1

type CertificateList []*x509.Certificate

type KeyDiscovery struct {
	Keys []Key `json:"keys"`
}

type EncodedCertificate string

type Key struct {
	Kid string               `json:"kid"`
	X5c []EncodedCertificate `json:"x5c"`
}

func FetchCertificates(discoveryURL string, log zerolog.Logger) (map[string]CertificateList, error) {
	log.Info().Msgf("Discover Microsoft signing certificates from %s", discoveryURL)
	azureKeyDiscovery, err := DiscoverURL(discoveryURL)
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Decoding certificates for %d keys", len(azureKeyDiscovery.Keys))
	azureCertificates, err := azureKeyDiscovery.Map()
	if err != nil {
		return nil, err
	}
	return azureCertificates, nil
}

// Map transform a KeyDiscovery object into a dictionary with "kid" as key
// and lists of decoded X509 certificates as values.
//
// Returns an error if any certificate does not decode.
func (k *KeyDiscovery) Map() (result map[string]CertificateList, err error) {
	result = make(map[string]CertificateList)

	for _, key := range k.Keys {
		certList := make(CertificateList, 0)
		for _, encodedCertificate := range key.X5c {
			certificate, err := encodedCertificate.Decode()
			if err != nil {
				return nil, err
			}
			certList = append(certList, certificate)
		}
		result[key.Kid] = certList
	}

	return result, err
}

// Decode a base64 encoded certificate into a X509 structure.
func (c EncodedCertificate) Decode() (*x509.Certificate, error) {
	stream := strings.NewReader(string(c))
	decoder := base64.NewDecoder(base64.StdEncoding, stream)
	key, err := io.ReadAll(decoder)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(key)
}

func DiscoverURL(url string) (*KeyDiscovery, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return Discover(response.Body)
}

func Discover(reader io.Reader) (*KeyDiscovery, error) {
	document, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	keyDiscovery := &KeyDiscovery{}
	err = json.Unmarshal(document, keyDiscovery)

	return keyDiscovery, err
}

func GetUser(ctx context.Context) *service.User {
	user := ctx.Value(ContextUserKey)
	if user == nil {
		return nil
	}

	return user.(*service.User)
}

func SetUser(ctx context.Context, user *service.User) context.Context {
	return context.WithValue(ctx, ContextUserKey, user)
}

type SessionRetriever interface {
	GetSession(ctx context.Context, token string) (*Session, error)
}

type Middleware struct {
	keyDiscoveryURL string
	tokenVerifier   *oidc.IDTokenVerifier
	groupsCache     *groupsCacher
	azureGroups     *AzureGroupClient
	googleGroups    *GoogleGroupClient
	texas           *TexasClient
	knastGroups     []string
	queries         *gensql.Queries
	log             zerolog.Logger
}

func newMiddleware(
	keyDiscoveryURL string,
	tokenVerifier *oidc.IDTokenVerifier,
	azureGroups *AzureGroupClient,
	googleGroups *GoogleGroupClient,
	texas *TexasClient,
	knastGroups []string,
	querier *gensql.Queries,
	log zerolog.Logger,
) *Middleware {
	return &Middleware{
		keyDiscoveryURL: keyDiscoveryURL,
		tokenVerifier:   tokenVerifier,
		azureGroups:     azureGroups,
		googleGroups:    googleGroups,
		groupsCache: &groupsCacher{
			cache: map[string]groupsCacheValue{},
		},
		texas:       texas,
		knastGroups: knastGroups,
		queries:     querier,
		log:         log,
	}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return m.handle(next)
}

func (m *Middleware) handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		token := r.Header.Get("authorization")

		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := m.texas.Introspect(ctx, token, ProviderAzureAD)
		if err != nil {
			m.log.Error().Err(err).Msg("Validation of token failed")
		}

		user := &service.User{
			Name:   claims.Name,
			Email:  claims.PreferredUsername,
			Ident:  claims.NavIdent,
			Expiry: time.Unix(claims.Exp, 0),
		}

		if err := m.addGroupsToUser(ctx, token, user); err != nil {
			m.log.Error().Err(err).Msg("Unable to add groups")
			w.Header().Add("Content-Type", "application/json")
			http.Error(w, `{"error": "Unable fetch users groups."}`, http.StatusInternalServerError)
			return
		}

		user.IsKnastUser = m.userInOneOfGroups(user.AzureGroups, m.knastGroups)

		r = r.WithContext(context.WithValue(ctx, ContextUserKey, user))

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) addGroupsToUser(ctx context.Context, token string, u *service.User) error {
	err := m.addAzureGroups(ctx, token, u)
	if err != nil {
		return fmt.Errorf("unable to add azure groups: %w", err)
	}

	err = m.addGoogleGroups(ctx, u)
	if err != nil {
		return fmt.Errorf("unable to add google groups: %w", err)
	}

	return nil
}

func (m *Middleware) addAzureGroups(ctx context.Context, token string, u *service.User) error {
	groups, ok := m.groupsCache.GetAzureGroups(u.Email)
	if ok {
		u.AzureGroups = groups
		return nil
	}

	groups, err := m.azureGroups.GroupsForUser(ctx, token, u.Email)
	if err != nil {
		return fmt.Errorf("getting groups for user: %w", err)
	}

	m.groupsCache.SetAzureGroups(u.Email, groups)
	u.AzureGroups = groups
	return nil
}

func (m *Middleware) addGoogleGroups(ctx context.Context, u *service.User) error {
	groups, ok := m.groupsCache.GetGoogleGroups(u.Email)
	if !ok {
		var err error
		groups, err = m.googleGroups.Groups(ctx, &u.Email)
		if err != nil {
			return fmt.Errorf("getting groups for user: %w", err)
		}

		m.groupsCache.SetGoogleGroups(u.Email, groups)
	}
	u.GoogleGroups = groups

	allGroups, ok := m.groupsCache.GetGoogleGroups("all")
	if !ok {
		var err error
		allGroups, err = m.googleGroups.Groups(ctx, nil)
		if err != nil {
			return fmt.Errorf("getting all groups: %w", err)
		}

		m.groupsCache.SetGoogleGroups("all", allGroups)
	}
	u.AllGoogleGroups = allGroups

	return nil
}

func (m *Middleware) userInOneOfGroups(userGroups service.AzureGroups, allowedGroupIDs []string) bool {
	for _, ug := range userGroups {
		if slices.Contains(allowedGroupIDs, ug.ObjectID) {
			return true
		}
	}

	return false
}
