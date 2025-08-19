package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
	"github.com/rs/zerolog"
)

type WorkstationsEndpoints struct {
	CreateWorkstationJob                         http.HandlerFunc
	GetWorkstationJob                            http.HandlerFunc
	GetWorkstationJobs                           http.HandlerFunc
	GetWorkstationResyncJobs                     http.HandlerFunc
	GetWorkstation                               http.HandlerFunc
	DeleteWorkstationByUser                      http.HandlerFunc
	DeleteWorkstationBySlug                      http.HandlerFunc
	StartWorkstation                             http.HandlerFunc
	GetWorkstationStartJobs                      http.HandlerFunc
	GetWorkstationStartJob                       http.HandlerFunc
	StopWorkstation                              http.HandlerFunc
	GetWorkstationOptions                        http.HandlerFunc
	GetWorkstationLogs                           http.HandlerFunc
	GetWorkstationZonalTagBindings               http.HandlerFunc
	ListWorkstations                             http.HandlerFunc
	UpdateWorkstationOnpremMapping               http.HandlerFunc
	GetWorkstationOnpremMapping                  http.HandlerFunc
	GetWorkstationURLList                        http.HandlerFunc
	GetWorkstationURLListForIdent                http.HandlerFunc
	CreateWorkstationURLListItemForIdent         http.HandlerFunc
	UpdateWorkstationURLListItemForIdent         http.HandlerFunc
	DeleteWorkstationURLListItemForIdent         http.HandlerFunc
	ScheduleWorkstationURLListActivationForIdent http.HandlerFunc
	UpdateWorkstationURLListSettings             http.HandlerFunc
	CreateWorkstationConnectivityWorkflow        http.HandlerFunc
	GetWorkstationConnectivityWorkflow           http.HandlerFunc
	RestartWorkstation                           http.HandlerFunc
	CreateResyncWorkstationJob                   http.HandlerFunc
	CreateResyncAllWorkstationsWorkflow          http.HandlerFunc
}

func NewWorkstationsEndpoints(log zerolog.Logger, h *handlers.WorkstationsHandler) *WorkstationsEndpoints {
	return &WorkstationsEndpoints{
		CreateWorkstationJob:                         transport.For(h.CreateWorkstationJob).RequestFromJSON().Build(log),
		GetWorkstationJob:                            transport.For(h.GetWorkstationJob).Build(log),
		GetWorkstationJobs:                           transport.For(h.GetWorkstationJobs).Build(log),
		GetWorkstationResyncJobs:                     transport.For(h.GetWorkstationResyncJobs).Build(log),
		GetWorkstation:                               transport.For(h.GetWorkstation).Build(log),
		DeleteWorkstationByUser:                      transport.For(h.DeleteWorkstationByUser).Build(log),
		DeleteWorkstationBySlug:                      transport.For(h.DeleteWorkstationBySlug).Build(log),
		StartWorkstation:                             transport.For(h.StartWorkstation).Build(log),
		GetWorkstationStartJob:                       transport.For(h.GetWorkstationStartJob).Build(log),
		GetWorkstationStartJobs:                      transport.For(h.GetWorkstationStartJobs).Build(log),
		StopWorkstation:                              transport.For(h.StopWorkstation).Build(log),
		GetWorkstationOptions:                        transport.For(h.GetWorkstationOptions).Build(log),
		GetWorkstationLogs:                           transport.For(h.GetWorkstationLogs).Build(log),
		GetWorkstationZonalTagBindings:               transport.For(h.GetWorkstationZonalTagBindings).Build(log),
		ListWorkstations:                             transport.For(h.ListWorkstations).Build(log),
		UpdateWorkstationOnpremMapping:               transport.For(h.UpdateWorkstationOnpremMapping).RequestFromJSON().Build(log),
		GetWorkstationOnpremMapping:                  transport.For(h.GetWorkstationOnpremMapping).Build(log),
		GetWorkstationURLListForIdent:                transport.For(h.GetWorkstationURLListForIdent).Build(log),
		ScheduleWorkstationURLListActivationForIdent: transport.For(h.ScheduleWorkstationURLListActivationForIdent).RequestFromJSON().Build(log),
		UpdateWorkstationURLListSettings:             transport.For(h.UpdateWorkstationURLListSettings).RequestFromJSON().Build(log),
		CreateWorkstationURLListItemForIdent:         transport.For(h.CreateWorkstationURLListItemForIdent).RequestFromJSON().Build(log),
		UpdateWorkstationURLListItemForIdent:         transport.For(h.UpdateWorkstationURLListItemForIdent).RequestFromJSON().Build(log),
		DeleteWorkstationURLListItemForIdent:         transport.For(h.DeleteWorkstationURLListItemForIdent).Build(log),
		CreateWorkstationConnectivityWorkflow:        transport.For(h.CreateWorkstationConnectivityWorkflow).RequestFromJSON().Build(log),
		GetWorkstationConnectivityWorkflow:           transport.For(h.GetWorkstationConnectivityWorkflow).Build(log),
		RestartWorkstation:                           transport.For(h.RestartWorkstation).Build(log),
		CreateResyncWorkstationJob:                   transport.For(h.CreateWorkstationResyncJob).Build(log),
		CreateResyncAllWorkstationsWorkflow:          transport.For(h.CreateResyncAllWorkstationsWorkflow).RequestFromJSON().Build(log),
	}
}

func NewWorkstationsRoutes(endpoints *WorkstationsEndpoints, auth func(http.Handler) http.Handler) AddRoutesFn {
	return func(router chi.Router) {
		router.With(auth).Route("/api/workstations", func(r chi.Router) {
			r.Post("/job", endpoints.CreateWorkstationJob)
			r.Get("/job", endpoints.GetWorkstationJobs)
			r.Get("/resyncjob", endpoints.GetWorkstationResyncJobs)
			r.Get("/job/{id}", endpoints.GetWorkstationJob)
			r.Get("/", endpoints.GetWorkstation)
			r.Delete("/", endpoints.DeleteWorkstationByUser)
			r.Delete("/{slug}", endpoints.DeleteWorkstationBySlug)
			r.Post("/start", endpoints.StartWorkstation)
			r.Get("/start", endpoints.GetWorkstationStartJobs)
			r.Get("/start/{id}", endpoints.GetWorkstationStartJob)
			r.Post("/stop", endpoints.StopWorkstation)
			r.Post("/restart", endpoints.RestartWorkstation)
			r.Post("/resync/{slug}", endpoints.CreateResyncWorkstationJob)
			r.Put("/urllist/activate", endpoints.ScheduleWorkstationURLListActivationForIdent)
			r.Get("/urllist", endpoints.GetWorkstationURLListForIdent)
			r.Post("/urllist", endpoints.CreateWorkstationURLListItemForIdent)
			r.Put("/urllist/{id}", endpoints.UpdateWorkstationURLListItemForIdent)
			r.Put("/urllist/settings", endpoints.UpdateWorkstationURLListSettings)
			r.Delete("/urllist/{id}", endpoints.DeleteWorkstationURLListItemForIdent)
			r.Get("/options", endpoints.GetWorkstationOptions)
			r.Get("/logs", endpoints.GetWorkstationLogs)
			r.Get("/bindings/tags", endpoints.GetWorkstationZonalTagBindings)
			r.Get("/list", endpoints.ListWorkstations)
			r.Put("/onpremhosts", endpoints.UpdateWorkstationOnpremMapping)
			r.Get("/onpremhosts", endpoints.GetWorkstationOnpremMapping)
			r.Get("/workflow/connectivity", endpoints.GetWorkstationConnectivityWorkflow)
			r.Post("/workflow/connectivity", endpoints.CreateWorkstationConnectivityWorkflow)
			r.Post("/workflow/resyncall", endpoints.CreateResyncAllWorkstationsWorkflow)
		})
	}
}
