package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	md "github.com/go-chi/chi/v5/middleware"

	"github.com/navikt/nada-backend/pkg/auth"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
)

type WorkstationsHandler struct {
	service service.WorkstationsService
}

type WorkstationJob service.WorkstationJob

func (j *WorkstationJob) StatusCode() int {
	return http.StatusAccepted
}

type WorkstationStartJob service.WorkstationStartJob

func (j *WorkstationStartJob) StatusCode() int {
	return http.StatusAccepted
}

type WorkstationZonalTagBindingsJob service.WorkstationZonalTagBindingsJob

func (j *WorkstationZonalTagBindingsJob) StatusCode() int {
	return http.StatusAccepted
}

func (h *WorkstationsHandler) CreateWorkstationJob(ctx context.Context, _ *http.Request, input *service.WorkstationInput) (*WorkstationJob, error) {
	const op errs.Op = "WorkstationsHandler.CreateWorkstation"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}
	if !user.IsKnastUser {
		return nil, errs.E(errs.Unauthorized, service.CodeNotNotAllowed, op, errs.Str("user not allowed to create Knast"))
	}

	raw, err := h.service.CreateWorkstationJob(ctx, user, input)
	if err != nil {
		return nil, errs.E(op, err)
	}

	job := WorkstationJob(*raw)

	return &job, nil
}

func (h *WorkstationsHandler) GetWorkstationJob(ctx context.Context, r *http.Request, _ any) (*service.WorkstationJob, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstationJob"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	raw := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, errs.E(errs.Invalid, op, err)
	}

	job, err := h.service.GetWorkstationJob(ctx, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	if job.Ident != user.Ident {
		return nil, errs.E(errs.Unauthorized, op, errs.Str("you are not authorized to view this job"))
	}

	return job, nil
}

func (h *WorkstationsHandler) GetWorkstationJobs(ctx context.Context, _ *http.Request, _ any) (*service.WorkstationJobs, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstationJobs"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	jobs, err := h.service.GetWorkstationJobsForUser(ctx, user.Ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return jobs, nil
}

func (h *WorkstationsHandler) GetWorkstation(ctx context.Context, _ *http.Request, _ any) (*service.WorkstationOutput, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstation"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	workstation, err := h.service.GetWorkstation(ctx, user)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return workstation, nil
}

func (h *WorkstationsHandler) GetWorkstationOptions(ctx context.Context, _ *http.Request, _ any) (*service.WorkstationOptions, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstationOptions"

	options, err := h.service.GetWorkstationOptions(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return options, nil
}

func (h *WorkstationsHandler) GetWorkstationLogs(ctx context.Context, _ *http.Request, _ any) (*service.WorkstationLogs, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstationsLogs"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	logs, err := h.service.GetWorkstationLogs(ctx, user)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return logs, nil
}

func (h *WorkstationsHandler) DeleteWorkstationByUser(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "WorkstationsHandler.DeleteWorkstationByUser"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	err := h.service.DeleteWorkstationByUser(ctx, user)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *WorkstationsHandler) DeleteWorkstationBySlug(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "WorkstationsHandler.DeleteWorkstationBySlug"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	if !user.IsAdmin() {
		return nil, errs.E(errs.Unauthorized, op, errs.Str("delete workstation is only available for members of nada@nav.no"))
	}

	slug := chi.URLParamFromCtx(ctx, "slug")

	err := h.service.DeleteWorkstationBySlug(ctx, slug)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *WorkstationsHandler) UpdateWorkstationURLList(ctx context.Context, _ *http.Request, input *service.WorkstationURLList) (*transport.Empty, error) {
	const op errs.Op = "WorkstationsHandler.UpdateWorkstationURLList"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	err := h.service.UpdateWorkstationURLList(ctx, user, input)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *WorkstationsHandler) StartWorkstation(ctx context.Context, _ *http.Request, _ any) (*WorkstationStartJob, error) {
	const op errs.Op = "WorkstationsHandler.StartWorkstation"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	raw, err := h.service.CreateWorkstationStartJob(ctx, user)
	if err != nil {
		return nil, errs.E(op, err)
	}

	job := WorkstationStartJob(*raw)

	return &job, nil
}

func (h *WorkstationsHandler) GetWorkstationStartJob(ctx context.Context, r *http.Request, _ any) (*service.WorkstationStartJob, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstationStartJob"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	raw := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, errs.E(errs.Invalid, op, err)
	}

	job, err := h.service.GetWorkstationStartJob(ctx, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	if job.Ident != user.Ident {
		return nil, errs.E(errs.Unauthorized, op, errs.Str("you are not authorized to view this job"))
	}

	return job, nil
}

func (h *WorkstationsHandler) GetWorkstationStartJobs(ctx context.Context, _ *http.Request, _ any) (*service.WorkstationStartJobs, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstationStartJobs"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	jobs, err := h.service.GetWorkstationStartJobsForUser(ctx, user.Ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return jobs, nil
}

func (h *WorkstationsHandler) GetWorkstationZonalTagBindings(ctx context.Context, _ *http.Request, _ any) (*service.EffectiveTags, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstationZonalTagBindings"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	tags, err := h.service.GetWorkstationZonalTagBindings(ctx, user.Ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &service.EffectiveTags{
		Tags: tags,
	}, nil
}

func (h *WorkstationsHandler) CreateWorkstationZonalTagBindingsJob(ctx context.Context, _ *http.Request, input *service.WorkstationOnpremAllowList) (*WorkstationZonalTagBindingsJob, error) {
	const op errs.Op = "WorkstationsHandler.CreateWorkstationZonalTagBindingsJob"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	job, err := h.service.CreateWorkstationZonalTagBindingsJobForUser(ctx, user.Ident, md.GetReqID(ctx), input)
	if err != nil {
		return nil, errs.E(op, err)
	}

	raw := WorkstationZonalTagBindingsJob(*job)

	return &raw, nil
}

func (h *WorkstationsHandler) GetWorkstationZonalTagBindingsJobs(ctx context.Context, _ *http.Request, _ any) (*service.WorkstationZonalTagBindingsJobs, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstationZonalTagBindingsJobs"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	jobs, err := h.service.GetWorkstationZonalTagBindingsJobsForUser(ctx, user.Ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &service.WorkstationZonalTagBindingsJobs{
		Jobs: jobs,
	}, nil
}

func (h *WorkstationsHandler) ListWorkstations(ctx context.Context, _ *http.Request, _ any) ([]*service.WorkstationOutput, error) {
	const op errs.Op = "WorkstationsHandler.ListWorkstations"

	workstations, err := h.service.ListWorkstations(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return workstations, nil
}

func (h *WorkstationsHandler) StopWorkstation(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "WorkstationsHandler.StopWorkstation"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	err := h.service.StopWorkstation(ctx, user, md.GetReqID(ctx))
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *WorkstationsHandler) UpdateWorkstationOnpremMapping(ctx context.Context, _ *http.Request, input *service.WorkstationOnpremAllowList) (*transport.Empty, error) {
	const op errs.Op = "WorkstationsHandler.UpdateWorkstationOnpremMapping"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	err := h.service.UpdateWorkstationOnpremMapping(ctx, user, input)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *WorkstationsHandler) GetWorkstationOnpremMapping(ctx context.Context, _ *http.Request, _ any) (*service.WorkstationOnpremAllowList, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstationOnpremMapping"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	mapping, err := h.service.GetWorkstationOnpremMapping(ctx, user)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return mapping, nil
}

func (h *WorkstationsHandler) GetWorkstationURLList(ctx context.Context, _ *http.Request, _ any) (*service.WorkstationURLList, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstationURLList"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	list, err := h.service.GetWorkstationURLList(ctx, user)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return list, nil
}

func (h *WorkstationsHandler) CreateWorkstationConnectivityWorkflow(ctx context.Context, _ *http.Request, input *service.WorkstationOnpremAllowList) (*service.WorkstationConnectivityWorkflow, error) {
	const op errs.Op = "WorkstationsHandler.CreateWorkstationConnectivityWorkflow"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	job, err := h.service.CreateWorkstationConnectivityWorkflow(ctx, user.Ident, md.GetReqID(ctx), input.Hosts)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return job, nil
}

func (h *WorkstationsHandler) GetWorkstationConnectivityWorkflow(ctx context.Context, _ *http.Request, _ any) (*service.WorkstationConnectivityWorkflow, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstationConnectivityWorkflow"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	workflow, err := h.service.GetWorkstationConnectivityWorkflow(ctx, user.Ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return workflow, nil
}

func NewWorkstationsHandler(service service.WorkstationsService) *WorkstationsHandler {
	return &WorkstationsHandler{
		service: service,
	}
}
