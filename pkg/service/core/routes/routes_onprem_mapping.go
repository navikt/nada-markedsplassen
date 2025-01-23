package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
	"github.com/rs/zerolog"
)

type OnpremMappingEndpoints struct {
	GetClassifiedHosts http.HandlerFunc
}

func NewOnpremMappingEndpoints(log zerolog.Logger, h *handlers.OnpremMappingHandler) *OnpremMappingEndpoints {
	return &OnpremMappingEndpoints{
		GetClassifiedHosts: transport.For(h.GetClassifiedHosts).Build(log),
	}
}

func NewOnpremMappingRoutes(endpoints *OnpremMappingEndpoints) AddRoutesFn {
	return func(router chi.Router) {
		router.Route("/api/onpremMapping", func(r chi.Router) {
			r.Get("/", endpoints.GetClassifiedHosts)
		})
	}
}
