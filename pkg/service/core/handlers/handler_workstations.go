package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

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

type WorkstationZonalTagBindingJobs service.WorkstationZonalTagBindingJobs

func (j *WorkstationZonalTagBindingJobs) StatusCode() int {
	return http.StatusAccepted
}

func (h *WorkstationsHandler) CreateWorkstationJob(ctx context.Context, _ *http.Request, input *service.WorkstationInput) (*WorkstationJob, error) {
	const op errs.Op = "WorkstationsHandler.CreateWorkstation"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
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

func IsAdmin(user *service.User) bool {
	return user.GoogleGroups.Contains("nada@nav.no")
}

func (h *WorkstationsHandler) DeleteWorkstationBySlug(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "WorkstationsHandler.DeleteWorkstationBySlug"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	if !IsAdmin(user) {
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

func (h *WorkstationsHandler) CreateWorkstationZonalTagBindingJobs(ctx context.Context, _ *http.Request, _ any) (*WorkstationZonalTagBindingJobs, error) {
	const op errs.Op = "WorkstationsHandler.CreateWorkstationZonalTagBindingJobs"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	jobs, err := h.service.CreateWorkstationZonalTagBindingJobsForUser(ctx, user.Ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	raw := WorkstationZonalTagBindingJobs(service.WorkstationZonalTagBindingJobs{
		Jobs: jobs,
	})

	return &raw, nil
}

func (h *WorkstationsHandler) GetWorkstationZonalTagBindingJobs(ctx context.Context, _ *http.Request, _ any) (*service.WorkstationZonalTagBindingJobs, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstationZonalTagBindingJobs"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	jobs, err := h.service.GetWorkstationZonalTagBindingJobsForUser(ctx, user.Ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &service.WorkstationZonalTagBindingJobs{
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

	err := h.service.StopWorkstation(ctx, user)
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

func NewWorkstationsHandler(service service.WorkstationsService) *WorkstationsHandler {
	return &WorkstationsHandler{
		service: service,
	}
}
