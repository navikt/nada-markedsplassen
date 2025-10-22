package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
	"github.com/rs/zerolog"
)

type MetabaseDashboardEndpoints struct {
	CreateMetabaseDashboard http.HandlerFunc
	GetMetabaseDashboard    http.HandlerFunc
	DeleteMetabaseDashboard http.HandlerFunc
	UpdateMetabaseDashboard http.HandlerFunc
}

func NewMetabaseDashboardEndpoints(
	log zerolog.Logger,
	h *handlers.MetabaseDashboardsHandler,
) *MetabaseDashboardEndpoints {
	return &MetabaseDashboardEndpoints{
		CreateMetabaseDashboard: transport.For(h.CreateMetabaseDashboard).RequestFromJSON().Build(log),
		GetMetabaseDashboard:    transport.For(h.GetMetabaseDashboard).Build(log),
		DeleteMetabaseDashboard: transport.For(h.DeleteMetabaseDashboard).Build(log),
		UpdateMetabaseDashboard: transport.For(h.UpdateMetabaseDashboard).RequestFromJSON().Build(log),
	}
}

func NewMetabaseDashboardRouter(
	endpoints *MetabaseDashboardEndpoints,
	auth func(http.Handler) http.Handler,
) AddRoutesFn {
	return func(router chi.Router) {
		router.Route("/api/metabaseDashboards", func(r chi.Router) {
			r.Use(auth)
			r.Post("/new", endpoints.CreateMetabaseDashboard)
			r.Get("/{id}", endpoints.GetMetabaseDashboard)
			r.Put("/{id}", endpoints.UpdateMetabaseDashboard)
			r.Delete("/{id}", endpoints.DeleteMetabaseDashboard)
		})
	}
}
