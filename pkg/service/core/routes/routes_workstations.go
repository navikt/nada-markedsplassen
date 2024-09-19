package routes

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
	"github.com/rs/zerolog"
)

type WorkstationsEndpoints struct {
	EnsureWorkstation http.HandlerFunc
	GetWorkstation    http.HandlerFunc
	DeleteWorkstation http.HandlerFunc
}

func NewWorkstationsEndpoints(log zerolog.Logger, h *handlers.WorkstationsHandler) *WorkstationsEndpoints {
	return &WorkstationsEndpoints{
		EnsureWorkstation: transport.For(h.EnsureWorkstation).RequestFromJSON().Build(log),
		GetWorkstation:    transport.For(h.GetWorkstation).Build(log),
		DeleteWorkstation: transport.For(h.DeleteWorkstation).Build(log),
	}
}

func NewWorkstationsRoutes(endpoints *WorkstationsEndpoints, auth func(http.Handler) http.Handler) AddRoutesFn {
	return func(router chi.Router) {
		router.With(auth).Route("/api/workstations", func(r chi.Router) {
			r.Post("/", endpoints.EnsureWorkstation)
			r.Get("/", endpoints.GetWorkstation)
			r.Delete("/", endpoints.DeleteWorkstation)
		})
	}
}
