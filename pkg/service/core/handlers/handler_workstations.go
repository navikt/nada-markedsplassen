package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

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

func (h *WorkstationsHandler) CreateWorkstationJob(ctx context.Context, _ *http.Request, input *service.WorkstationInput) (*WorkstationJob, error) {
	const op errs.Op = "WorkstationsHandler.CreateWorkstation"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, op, errs.Str("no user in context"))
	}

	if !containsGroup(user.GoogleGroups, "nada@nav.no") {
		return nil, errs.E(errs.Unauthorized, op, errs.Str("the workstation feature is only available for members of nada@nav.no"))
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
		return nil, errs.E(errs.Unauthenticated, op, errs.Str("no user in context"))
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
		return nil, errs.E(errs.Unauthenticated, op, errs.Str("no user in context"))
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
		return nil, errs.E(errs.Unauthenticated, op, errs.Str("no user in context"))
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
		return nil, errs.E(errs.Unauthenticated, op, errs.Str("no user in context"))
	}

	logs, err := h.service.GetWorkstationLogs(ctx, user)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return logs, nil
}

func (h *WorkstationsHandler) DeleteWorkstation(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "WorkstationsHandler.DeleteWorkstation"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, op, errs.Str("no user in context"))
	}

	err := h.service.DeleteWorkstation(ctx, user)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *WorkstationsHandler) UpdateWorkstationURLList(ctx context.Context, _ *http.Request, input *service.WorkstationURLList) (*transport.Empty, error) {
	const op errs.Op = "WorkstationsHandler.UpdateWorkstationURLList"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, op, errs.Str("no user in context"))
	}

	err := h.service.UpdateWorkstationURLList(ctx, user, input)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *WorkstationsHandler) StartWorkstation(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "WorkstationsHandler.StartWorkstation"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, op, errs.Str("no user in context"))
	}

	err := h.service.StartWorkstation(ctx, user)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *WorkstationsHandler) StopWorkstation(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "WorkstationsHandler.StopWorkstation"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, op, errs.Str("no user in context"))
	}

	err := h.service.StopWorkstation(ctx, user)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func NewWorkstationsHandler(service service.WorkstationsService) *WorkstationsHandler {
	return &WorkstationsHandler{
		service: service,
	}
}

func containsGroup(groups service.Groups, groupEmail string) bool {
	for _, g := range groups {
		if g.Email == groupEmail {
			return true
		}
	}

	return false
}
