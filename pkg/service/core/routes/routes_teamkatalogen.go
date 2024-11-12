package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
	"github.com/rs/zerolog"
)

type TeamkatalogenEndpoints struct {
	SearchTeamKatalogen http.HandlerFunc
}

func NewTeamkatalogenEndpoints(log zerolog.Logger, h *handlers.TeamkatalogenHandler) *TeamkatalogenEndpoints {
	return &TeamkatalogenEndpoints{
		SearchTeamKatalogen: transport.For(h.SearchTeamKatalogen).Build(log),
	}
}

func NewTeamkatalogenRoutes(endpoints *TeamkatalogenEndpoints) AddRoutesFn {
	return func(router chi.Router) {
		router.Get("/api/teamkatalogen", endpoints.SearchTeamKatalogen)
	}
}
