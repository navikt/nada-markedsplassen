package auth

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
)

type MiddlewareHandler func(http.Handler) http.Handler

type contextKey int

const ContextUserKey contextKey = 1

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

type Middleware struct {
	groupsCache  *groupsCacher
	azureGroups  *AzureGroupClient
	googleGroups *GoogleGroupClient
	texas        *TexasClient
	knastGroups  []string
	log          zerolog.Logger
}

func NewMiddleware(
	azureGroups *AzureGroupClient,
	googleGroups *GoogleGroupClient,
	texas *TexasClient,
	knastGroups []string,
	log zerolog.Logger,
) *Middleware {
	return &Middleware{
		azureGroups:  azureGroups,
		googleGroups: googleGroups,
		groupsCache: &groupsCacher{
			cache: map[string]groupsCacheValue{},
		},
		texas:       texas,
		knastGroups: knastGroups,
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
