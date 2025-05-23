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
	DeleteMetabaseBigQueryRestrictedDataset    http.HandlerFunc

	GetMetabaseBigQueryOpenDatasetStatus http.HandlerFunc
	CreateMetabaseBigQueryOpenDataset    http.HandlerFunc
	DeleteMetabaseBigQueryOpenDataset    http.HandlerFunc
}

func NewMetabaseEndpoints(log zerolog.Logger, h *handlers.MetabaseHandler) *MetabaseEndpoints {
	return &MetabaseEndpoints{
		GetMetabaseBigQueryRestrictedDatasetStatus: transport.For(h.GetMetabaseBigQueryRestrictedDatasetStatus).Build(log),
		CreateMetabaseBigQueryRestrictedDataset:    transport.For(h.CreateMetabaseBigQueryRestrictedDataset).Build(log),
		DeleteMetabaseBigQueryRestrictedDataset:    transport.For(h.DeleteMetabaseBigQueryRestrictedDataset).Build(log),

		GetMetabaseBigQueryOpenDatasetStatus: transport.For(h.GetMetabaseBigQueryOpenDatasetStatus).Build(log),
		CreateMetabaseBigQueryOpenDataset:    transport.For(h.CreateMetabaseBigQueryOpenDataset).Build(log),
		DeleteMetabaseBigQueryOpenDataset:    transport.For(h.DeleteMetabaseBigQueryOpenDataset).Build(log),
	}
}

func NewMetabaseRoutes(endpoints *MetabaseEndpoints, auth func(http.Handler) http.Handler) AddRoutesFn {
	return func(router chi.Router) {
		// Might otherwise conflict with DatasetRoutes in routes_dataproducts.go
		router.Get("/api/datasets/{id}/bigquery_restricted", endpoints.GetMetabaseBigQueryRestrictedDatasetStatus)
		router.With(auth).Post("/api/datasets/{id}/bigquery_restricted", endpoints.CreateMetabaseBigQueryRestrictedDataset)
		router.With(auth).Delete("/api/datasets/{id}/bigquery_restricted", endpoints.DeleteMetabaseBigQueryRestrictedDataset)

		router.Get("/api/datasets/{id}/bigquery_open", endpoints.GetMetabaseBigQueryOpenDatasetStatus)
		router.With(auth).Post("/api/datasets/{id}/bigquery_open", endpoints.CreateMetabaseBigQueryOpenDataset)
		router.With(auth).Delete("/api/datasets/{id}/bigquery_open", endpoints.DeleteMetabaseBigQueryOpenDataset)
	}
}
