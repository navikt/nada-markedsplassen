package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
	"github.com/rs/zerolog"
)

type WorkstationsEndpoints struct {
	CreateWorkstationJob                 http.HandlerFunc
	GetWorkstationJob                    http.HandlerFunc
	GetWorkstationJobs                   http.HandlerFunc
	GetWorkstation                       http.HandlerFunc
	DeleteWorkstationByUser              http.HandlerFunc
	DeleteWorkstationBySlug              http.HandlerFunc
	StartWorkstation                     http.HandlerFunc
	GetWorkstationStartJobs              http.HandlerFunc
	GetWorkstationStartJob               http.HandlerFunc
	StopWorkstation                      http.HandlerFunc
	UpdateWorkstationURLList             http.HandlerFunc
	GetWorkstationOptions                http.HandlerFunc
	GetWorkstationLogs                   http.HandlerFunc
	CreateWorkstationZonalTagBindingJobs http.HandlerFunc
	GetWorkstationZonalTagBindingJobs    http.HandlerFunc
	GetWorkstationZonalTagBindings       http.HandlerFunc
	ListWorkstations                     http.HandlerFunc
	UpdateWorkstationOnpremMapping       http.HandlerFunc
	GetWorkstationOnpremMapping          http.HandlerFunc
}

func NewWorkstationsEndpoints(log zerolog.Logger, h *handlers.WorkstationsHandler) *WorkstationsEndpoints {
	return &WorkstationsEndpoints{
		CreateWorkstationJob:                 transport.For(h.CreateWorkstationJob).RequestFromJSON().Build(log),
		GetWorkstationJob:                    transport.For(h.GetWorkstationJob).Build(log),
		GetWorkstationJobs:                   transport.For(h.GetWorkstationJobs).Build(log),
		GetWorkstation:                       transport.For(h.GetWorkstation).Build(log),
		DeleteWorkstationByUser:              transport.For(h.DeleteWorkstationByUser).Build(log),
		DeleteWorkstationBySlug:              transport.For(h.DeleteWorkstationBySlug).Build(log),
		StartWorkstation:                     transport.For(h.StartWorkstation).Build(log),
		GetWorkstationStartJob:               transport.For(h.GetWorkstationStartJob).Build(log),
		GetWorkstationStartJobs:              transport.For(h.GetWorkstationStartJobs).Build(log),
		StopWorkstation:                      transport.For(h.StopWorkstation).Build(log),
		UpdateWorkstationURLList:             transport.For(h.UpdateWorkstationURLList).RequestFromJSON().Build(log),
		GetWorkstationOptions:                transport.For(h.GetWorkstationOptions).Build(log),
		GetWorkstationLogs:                   transport.For(h.GetWorkstationLogs).Build(log),
		CreateWorkstationZonalTagBindingJobs: transport.For(h.CreateWorkstationZonalTagBindingJobs).Build(log),
		GetWorkstationZonalTagBindingJobs:    transport.For(h.GetWorkstationZonalTagBindingJobs).Build(log),
		GetWorkstationZonalTagBindings:       transport.For(h.GetWorkstationZonalTagBindings).Build(log),
		ListWorkstations:                     transport.For(h.ListWorkstations).Build(log),
		UpdateWorkstationOnpremMapping:       transport.For(h.UpdateWorkstationOnpremMapping).RequestFromJSON().Build(log),
		GetWorkstationOnpremMapping:          transport.For(h.GetWorkstationOnpremMapping).Build(log),
	}
}

func NewWorkstationsRoutes(endpoints *WorkstationsEndpoints, auth func(http.Handler) http.Handler) AddRoutesFn {
	return func(router chi.Router) {
		router.With(auth).Route("/api/workstations", func(r chi.Router) {
			r.Post("/job", endpoints.CreateWorkstationJob)
			r.Get("/job", endpoints.GetWorkstationJobs)
			r.Get("/job/{id}", endpoints.GetWorkstationJob)
			r.Get("/", endpoints.GetWorkstation)
			r.Delete("/", endpoints.DeleteWorkstationByUser)
			r.Delete("/{slug}", endpoints.DeleteWorkstationBySlug)
			r.Post("/start", endpoints.StartWorkstation)
			r.Get("/start", endpoints.GetWorkstationStartJobs)
			r.Get("/start/{id}", endpoints.GetWorkstationStartJob)
			r.Post("/stop", endpoints.StopWorkstation)
			r.Put("/urllist", endpoints.UpdateWorkstationURLList)
			r.Get("/options", endpoints.GetWorkstationOptions)
			r.Get("/logs", endpoints.GetWorkstationLogs)
			r.Post("/bindings", endpoints.CreateWorkstationZonalTagBindingJobs)
			r.Get("/bindings", endpoints.GetWorkstationZonalTagBindingJobs)
			r.Get("/bindings/tags", endpoints.GetWorkstationZonalTagBindings)
			r.Get("/list", endpoints.ListWorkstations)
			r.Put("/onpremhosts", endpoints.UpdateWorkstationOnpremMapping)
			r.Get("/onpremhosts", endpoints.GetWorkstationOnpremMapping)
		})
	}
}
