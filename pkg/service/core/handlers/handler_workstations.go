package handlers

import (
	"context"
	"net/http"

	"github.com/navikt/nada-backend/pkg/auth"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
)

type WorkstationsHandler struct {
	service service.WorkstationsService
}

func (h *WorkstationsHandler) EnsureWorkstation(ctx context.Context, r *http.Request, input *service.WorkstationInput) (*service.WorkstationOutput, error) {
	const op errs.Op = "WorkstationsHandler.CreateWorkstation"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, op, errs.Str("no user in context"))
	}

	if !containsGroup(user.GoogleGroups, "nada@nav.no") {
		return nil, errs.E(errs.Unauthorized, op, errs.Str("the workstation feature is only available for members of nada@nav.no"))
	}

	workstation, err := h.service.EnsureWorkstation(ctx, user, input)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return workstation, nil
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

func (h *WorkstationsHandler) StartWorkstation(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
	return &transport.Empty{}, nil
}

func (h *WorkstationsHandler) StopWorkstation(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
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
