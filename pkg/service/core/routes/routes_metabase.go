package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
	"github.com/rs/zerolog"
)

type MetabaseEndpoints struct {
	GetMetabaseBigQueryRestrictedDatasetStatus http.HandlerFunc
	CreateMetabaseBigQueryRestrictedDataset    http.HandlerFunc
	GetMetabaseBigQueryOpenDatasetStatus       http.HandlerFunc
	CreateMetabaseBigQueryOpenDataset          http.HandlerFunc
}

func NewMetabaseEndpoints(log zerolog.Logger, h *handlers.MetabaseHandler) *MetabaseEndpoints {
	return &MetabaseEndpoints{
		GetMetabaseBigQueryRestrictedDatasetStatus: transport.For(h.GetMetabaseBigQueryRestrictedDatasetStatus).Build(log),
		CreateMetabaseBigQueryRestrictedDataset:    transport.For(h.CreateMetabaseBigQueryRestrictedDataset).Build(log),
	}
}

func NewMetabaseRoutes(endpoints *MetabaseEndpoints, auth func(http.Handler) http.Handler) AddRoutesFn {
	return func(router chi.Router) {
		// Might otherwise conflict with DatasetRoutes in routes_dataproducts.go
		router.Get("/api/datasets/{id}/bigquery_restricted", endpoints.GetMetabaseBigQueryRestrictedDatasetStatus)
		router.With(auth).Post("/api/datasets/{id}/bigquery_restricted", endpoints.GetMetabaseBigQueryRestrictedDatasetStatus)
		router.Get("/api/datasets/{id}/bigquery_open", endpoints.GetMetabaseBigQueryOpenDatasetStatus)
		router.With(auth).Post("/api/datasets/{id}/bigquery_open", endpoints.CreateMetabaseBigQueryOpenDataset)
	}
}
