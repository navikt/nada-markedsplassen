package routes

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
	"github.com/rs/zerolog"
)

type WorkstationsEndpoints struct {
	EnsureWorkstation        http.HandlerFunc
	GetWorkstation           http.HandlerFunc
	DeleteWorkstation        http.HandlerFunc
	StartWorkstation         http.HandlerFunc
	StopWorkstation          http.HandlerFunc
	UpdateWorkstationURLList http.HandlerFunc
}

func NewWorkstationsEndpoints(log zerolog.Logger, h *handlers.WorkstationsHandler) *WorkstationsEndpoints {
	return &WorkstationsEndpoints{
		EnsureWorkstation:        transport.For(h.EnsureWorkstation).RequestFromJSON().Build(log),
		GetWorkstation:           transport.For(h.GetWorkstation).Build(log),
		DeleteWorkstation:        transport.For(h.DeleteWorkstation).Build(log),
		StartWorkstation:         transport.For(h.StartWorkstation).Build(log),
		StopWorkstation:          transport.For(h.StopWorkstation).Build(log),
		UpdateWorkstationURLList: transport.For(h.UpdateWorkstationURLList).RequestFromJSON().Build(log),
	}
}

func NewWorkstationsRoutes(endpoints *WorkstationsEndpoints, auth func(http.Handler) http.Handler) AddRoutesFn {
	return func(router chi.Router) {
		router.With(auth).With(transport.UserInfo()).Route("/api/test/workstations", func(r chi.Router) {
			r.Get("/", endpoints.GetWorkstation)
		})

		router.With(auth).Route("/api/workstations", func(r chi.Router) {
			r.Post("/", endpoints.EnsureWorkstation)
			r.Get("/", endpoints.GetWorkstation)
			r.Delete("/", endpoints.DeleteWorkstation)
			r.Post("/start", endpoints.StartWorkstation)
			r.Post("/stop", endpoints.StopWorkstation)
			r.Put("/urllist", endpoints.UpdateWorkstationURLList)
		})
	}
}
