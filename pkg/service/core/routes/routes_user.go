package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
	"github.com/rs/zerolog"
)

type UserEndpoints struct {
	GetUserData http.HandlerFunc
	GetTokens   http.HandlerFunc
}

func NewUserEndpoints(log zerolog.Logger, h *handlers.UserHandler) *UserEndpoints {
	return &UserEndpoints{
		GetUserData: transport.For(h.GetUserData).Build(log),
		GetTokens:   transport.For(h.GetTokens).Build(log),
	}
}

func NewUserRoutes(endpoints *UserEndpoints, auth func(http.Handler) http.Handler) AddRoutesFn {
	return func(router chi.Router) {
		router.Route("/api/userData", func(r chi.Router) {
			r.Use(auth)
			r.Get("/", endpoints.GetUserData)
		})
		router.Route("/api/tokens", func(r chi.Router) {
			r.Use(auth)
			r.Get("/", endpoints.GetTokens)
		})
	}
}
