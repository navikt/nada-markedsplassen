package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
	"github.com/rs/zerolog"
)

type MetabaseDocumentEndpoints struct {
	CreateMetabaseDocument http.HandlerFunc
	GetMetabaseDocument    http.HandlerFunc
	UpdateMetabaseDocument http.HandlerFunc
	DeleteMetabaseDocument http.HandlerFunc
}

func NewMetabaseDocumentEndpoints(log zerolog.Logger, h *handlers.MetabaseDocumentsHandler) *MetabaseDocumentEndpoints {
	return &MetabaseDocumentEndpoints{
		CreateMetabaseDocument: transport.For(h.CreateMetabaseDocument).RequestFromJSON().Build(log),
		GetMetabaseDocument:    transport.For(h.GetMetabaseDocument).Build(log),
		UpdateMetabaseDocument: transport.For(h.UpdateMetabaseDocument).RequestFromJSON().Build(log),
		DeleteMetabaseDocument: transport.For(h.DeleteMetabaseDocument).Build(log),
	}
}

func NewMetabaseDocumentRouter(endpoints *MetabaseDocumentEndpoints, auth func(http.Handler) http.Handler) AddRoutesFn {
	return func(router chi.Router) {
		router.Route("/api/metabase-documents", func(r chi.Router) {
			r.Use(auth)
			r.Post("/", endpoints.CreateMetabaseDocument)
			r.Get("/{id}", endpoints.GetMetabaseDocument)
			r.Put("/{id}", endpoints.UpdateMetabaseDocument)
			r.Delete("/{id}", endpoints.DeleteMetabaseDocument)
		})
	}
}
