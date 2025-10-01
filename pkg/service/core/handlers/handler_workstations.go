package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	md "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

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

type WorkstationResyncJob service.WorkstationResyncJob

func (j *WorkstationResyncJob) StatusCode() int {
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

func (h *WorkstationsHandler) GetWorkstationResyncJobs(ctx context.Context, _ *http.Request, _ any) (*service.WorkstationResyncJobs, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstationResyncJobs"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	jobs, err := h.service.GetWorkstationResyncJobsForUser(ctx, user.Ident)
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

func (h *WorkstationsHandler) GetWorkstationURLListGlobalAllow(ctx context.Context, _ *http.Request, _ any) (*service.WorkstationURLListGlobalAllow, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstationURLListGlobalAllow"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	globalAllow, err := h.service.GetWorkstationURLListGlobalAllow(ctx, user)
	if err != nil {
		return nil, errs.E(op, err)
	}
	return globalAllow, nil
}

func (h *WorkstationsHandler) GetWorkstationURLListForIdent(ctx context.Context, _ *http.Request, _ any) (*service.WorkstationURLListForIdent, error) {
	const op errs.Op = "WorkstationsHandler.GetWorkstationURLListForIdent"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	list, err := h.service.GetWorkstationURLListForIdent(ctx, user)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return list, nil
}

func (h *WorkstationsHandler) CreateWorkstationURLListItemForIdent(ctx context.Context, _ *http.Request, input *service.WorkstationURLListItem) (*transport.Created, error) {
	const op errs.Op = "WorkstationsHandler.CreateWorkstationURLListItemForIdent"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	_, err := h.service.CreateWorkstationURLListItemForIdent(ctx, user, input)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Created{}, nil
}

func (h *WorkstationsHandler) UpdateWorkstationURLListItemForIdent(ctx context.Context, _ *http.Request, input *service.WorkstationURLListItem) (*transport.OK, error) {
	const op errs.Op = "WorkstationsHandler.UpdateWorkstationURLListItemForIdent"
	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	_, err := h.service.UpdateWorkstationURLListItemForIdent(ctx, user, input)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.OK{}, nil
}

func (h *WorkstationsHandler) DeleteWorkstationURLListItemForIdent(ctx context.Context, r *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "WorkstationsHandler.DeleteWorkstationURLListItemForIdent"
	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	raw := chi.URLParam(r, "id")

	itemID, err := uuid.Parse(raw)
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, service.CodeInvalidURLListItemID, op, err)
	}

	err = h.service.DeleteWorkstationURLListItemForIdent(ctx, itemID)
	if err != nil {
		return nil, errs.E(op, err)
	}
	return &transport.Empty{}, nil
}

func (h *WorkstationsHandler) ActivateWorkstationURLListForIdent(ctx context.Context, _ *http.Request, input *service.WorkstationURLListItems) (*transport.Empty, error) {
	const op errs.Op = "WorkstationsHandler.ActivateWorkstationURLListForIdent"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	err := h.service.ActivateWorkstationURLListForIdent(ctx, user.Ident, input.ItemIDs)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *WorkstationsHandler) UpdateWorkstationURLListSettings(ctx context.Context, _ *http.Request, input *service.WorkstationURLListSettingsOpts) (*transport.OK, error) {
	const op errs.Op = "WorkstationsHandler.UpdateWorkstationURLListSettings"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	err := h.service.EnsureWorkstationURLListSettingsForIdent(ctx, user, input)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.OK{}, nil
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

func (h *WorkstationsHandler) RestartWorkstation(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "WorkstationJob.RestartWorkstation"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	err := h.service.RestartWorkstation(ctx, user, md.GetReqID(ctx))
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *WorkstationsHandler) CreateWorkstationResyncJob(ctx context.Context, _ *http.Request, _ any) (*WorkstationResyncJob, error) {
	const op errs.Op = "WorkstationJob.CreateWorkstationResyncJob"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	if !user.IsAdmin() {
		return nil, errs.E(errs.Unauthorized, op, errs.Str("resync workstation can only be initiated by admins"))
	}

	slug := chi.URLParamFromCtx(ctx, "slug")

	raw, err := h.service.CreateWorkstationResyncJob(ctx, slug)
	if err != nil {
		return nil, errs.E(op, err)
	}

	job := WorkstationResyncJob(*raw)

	return &job, nil
}

func (h *WorkstationsHandler) CreateResyncAllWorkstationsWorkflow(ctx context.Context, _ *http.Request, input *service.ResyncAll) (*transport.Empty, error) {
	const op errs.Op = "WorkstationsHandler.CreateResyncAllWorkstationsWorkflow"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	if !user.IsAdmin() {
		return nil, errs.E(errs.Unauthorized, op, errs.Str("resync workstation can only be initiated by admins"))
	}

	err := h.service.CreateWorkstationResyncAllWorkflow(ctx, user.Ident, input.Slugs)
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

func (h *WorkstationsHandler) ConfigWorkstationSSH(ctx context.Context, r *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "WorkstationsHandler.ConfigWorkstationSSH"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	allow := strings.ToLower(r.URL.Query().Get("allow")) == "true"

	err := h.service.CreateConfigWorkstationSSHJob(ctx, user.Ident, allow)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *WorkstationsHandler) GetConfigWorkstationSSHJob(ctx context.Context, _ *http.Request, _ any) (*service.ConfigWorkstationSSHJob, error) {
	const op errs.Op = "WorkstationsHandler.GetConfigWorkstationSSHJob"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	jobs, err := h.service.GetConfigWorkstationSSHJob(ctx, user.Ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return jobs, nil
}
