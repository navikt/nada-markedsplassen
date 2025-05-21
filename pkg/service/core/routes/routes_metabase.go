package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
	"github.com/rs/zerolog"
)

type MetabaseEndpoints struct {
	MapDataset                    http.HandlerFunc
	MetabaseBigQueryDatasetStatus http.HandlerFunc
}

func NewMetabaseEndpoints(log zerolog.Logger, h *handlers.MetabaseHandler) *MetabaseEndpoints {
	return &MetabaseEndpoints{
		MapDataset:                    transport.For(h.MapDataset).RequestFromJSON().Build(log),
		MetabaseBigQueryDatasetStatus: transport.For(h.MetabaseBigQueryDatasetStatus).Build(log),
	}
}

func NewMetabaseRoutes(endpoints *MetabaseEndpoints, auth func(http.Handler) http.Handler) AddRoutesFn {
	return func(router chi.Router) {
		// Might otherwise conflict with DatasetRoutes in routes_dataproducts.go
		router.With(auth).Post("/api/datasets/{id}/map", endpoints.MapDataset)
		router.Get("/api/datasets/{id}/map_status", endpoints.MetabaseBigQueryDatasetStatus)
	}
}
