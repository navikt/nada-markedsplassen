package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/rs/zerolog"
)

type RiverEndpoints struct {
	Handle http.HandlerFunc
}

func NewRiverEndpoints(_ zerolog.Logger, h *handlers.RiverHandler) *RiverEndpoints {
	return &RiverEndpoints{
		Handle: h.Handle,
	}
}

func NewRiverRoutes(endpoints *RiverEndpoints, auth func(http.Handler) http.Handler) AddRoutesFn {
	return func(router chi.Router) {
		router.Route("/river", func(r chi.Router) {
			r.Use(auth)
			r.Handle("/", endpoints.Handle)
			r.Handle("/*", endpoints.Handle)
		})
	}
}
