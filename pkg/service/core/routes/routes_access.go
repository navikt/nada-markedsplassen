package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
	"github.com/rs/zerolog"
)

type AccessEndpoints struct {
	GetAccessRequests                       http.HandlerFunc
	ProcessAccessRequest                    http.HandlerFunc
	CreateAccessRequest                     http.HandlerFunc
	DeleteAccessRequest                     http.HandlerFunc
	UpdateAccessRequest                     http.HandlerFunc
	GrantBigQueryAccessToDataset            http.HandlerFunc
	RevokeBigQueryAccessToDataset           http.HandlerFunc
	GrantMetabaseAccessToDataset            http.HandlerFunc
	RevokeRestrictedMetabaseAccessToDataset http.HandlerFunc
	RevokeMetabaseAllUsersAccessToDataset   http.HandlerFunc
}

func NewAccessEndpoints(log zerolog.Logger, h *handlers.AccessHandler) *AccessEndpoints {
	return &AccessEndpoints{
		GetAccessRequests:             transport.For(h.GetAccessRequests).Build(log),
		ProcessAccessRequest:          transport.For(h.ProcessAccessRequest).Build(log),
		CreateAccessRequest:           transport.For(h.NewAccessRequest).RequestFromJSON().Build(log),
		DeleteAccessRequest:           transport.For(h.DeleteAccessRequest).Build(log),
		UpdateAccessRequest:           transport.For(h.UpdateAccessRequest).RequestFromJSON().Build(log),
		GrantBigQueryAccessToDataset:  transport.For(h.GrantBigQueryAccessToDataset).RequestFromJSON().Build(log),
		RevokeBigQueryAccessToDataset: transport.For(h.RevokeBigQueryAccessToDataset).Build(log),

		GrantMetabaseAccessToDataset:            transport.For(h.GrantMetabaseAccessToDataset).RequestFromJSON().Build(log),
		RevokeRestrictedMetabaseAccessToDataset: transport.For(h.RevokeMetabaseRestrictedAccessToDataset).Build(log),
		RevokeMetabaseAllUsersAccessToDataset:   transport.For(h.RevokeMetabaseAllUsersAccessToDataset).Build(log),
	}
}

func NewAccessRoutes(endpoints *AccessEndpoints, auth func(http.Handler) http.Handler) AddRoutesFn {
	return func(router chi.Router) {
		router.Route("/api/accessRequests", func(r chi.Router) {
			r.Use(auth)
			r.Get("/", endpoints.GetAccessRequests)
			r.Post("/process/{id}", endpoints.ProcessAccessRequest)
			r.Post("/new", endpoints.CreateAccessRequest)
			r.Delete("/{id}", endpoints.DeleteAccessRequest)
			// FIXME: dont seem to use the ID in the URL
			r.Put("/{id}", endpoints.UpdateAccessRequest)
		})

		router.Route("/api/accesses/bigquery", func(r chi.Router) {
			r.Use(auth)
			r.Post("/grant", endpoints.GrantBigQueryAccessToDataset)
			r.Post("/revoke", endpoints.RevokeBigQueryAccessToDataset)
		})

		router.Route("/api/accesses/metabase", func(r chi.Router) {
			r.Use(auth)
			r.Post("/grant", endpoints.GrantMetabaseAccessToDataset)
			r.Post("/revoke", endpoints.RevokeRestrictedMetabaseAccessToDataset)
			r.Post("/revokeAllUsers", endpoints.RevokeMetabaseAllUsersAccessToDataset)
		})
	}
}
