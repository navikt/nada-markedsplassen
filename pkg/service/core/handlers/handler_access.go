package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/auth"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
)

type AccessHandler struct {
	accessService service.AccessService
	gcpProjectID  string
}

func (h *AccessHandler) RevokeBigQueryAccessToDataset(ctx context.Context, r *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "AccessHandler.RevokeBigQueryAccessToDataset"

	id, err := uuid.Parse(r.URL.Query().Get("accessId"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("parsing id: %w", err))
	}

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	access, err := h.accessService.GetAccessToDataset(ctx, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	if access.Platform != service.AccessPlatformBigQuery {
		return nil, errs.E(op, fmt.Errorf("attempted to revoke BigQuery access for wrong platform %s", access.Platform))
	}

	err = h.accessService.EnsureUserIsAuthorizedToRevokeAccess(ctx, user, access)
	if err != nil {
		return nil, errs.E(op, err)
	}

	err = h.accessService.RevokeBigQueryAccessToDataset(ctx, access, h.gcpProjectID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *AccessHandler) GrantBigQueryAccessToDataset(ctx context.Context, _ *http.Request, in service.GrantAccessData) (*transport.Empty, error) {
	const op errs.Op = "AccessHandler.GrantBigQueryAccessToDataset"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	if err := h.accessService.EnsureUserIsDatasetOwner(ctx, user, in.DatasetID); err != nil {
		return nil, errs.E(op, err)
	}

	if in.Expires != nil && in.Expires.Before(time.Now()) {
		return nil, errs.E(errs.InvalidRequest, service.CodeExpiresInPast, op, fmt.Errorf("expires is in the past"))
	}

	err := h.accessService.GrantBigQueryAccessToDataset(ctx, user, in, h.gcpProjectID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *AccessHandler) GrantMetabaseAccessToDataset(ctx context.Context, _ *http.Request, in service.GrantAccessData) (*transport.Empty, error) {
	const op errs.Op = "AccessHandler.GrantMetabaseAccessToDataset"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	if err := h.accessService.EnsureUserIsDatasetOwner(ctx, user, in.DatasetID); err != nil {
		return nil, errs.E(op, err)
	}

	if in.Expires != nil && in.Expires.Before(time.Now()) {
		return nil, errs.E(errs.InvalidRequest, service.CodeExpiresInPast, op, fmt.Errorf("expires is in the past"))
	}

	if in.SubjectType != nil && *in.SubjectType != "user" {
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("only subjectType 'user' is supported for restricted metabase database access grants"))
	}

	err := h.accessService.GrantMetabaseAccessRestrictedToDataset(ctx, user, in)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *AccessHandler) RevokeMetabaseAllUsersAccessToDataset(ctx context.Context, r *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "AccessHandler.RevokeMetabaseAllUsersAccessToDataset"

	accessID, err := uuid.Parse(r.URL.Query().Get("accessId"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("parsing id: %w", err))
	}

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	access, err := h.accessService.GetAccessToDataset(ctx, accessID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	err = h.accessService.EnsureUserIsAuthorizedToRevokeAccess(ctx, user, access)
	if err != nil {
		return nil, errs.E(op, err)
	}

	err = h.accessService.RevokeMetabaseAllUsersAccessToDataset(ctx, access)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *AccessHandler) RevokeMetabaseRestrictedAccessToDataset(ctx context.Context, r *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "AccessHandler.RevokeMetabaseRestrictedAccessToDataset"

	id, err := uuid.Parse(r.URL.Query().Get("accessId"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("parsing id: %w", err))
	}

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	access, err := h.accessService.GetAccessToDataset(ctx, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	if strings.Split(access.Subject, ":")[0] != "user" {
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("only user subjectType is supported for metabase restricted database access revokes"))
	}

	err = h.accessService.EnsureUserIsAuthorizedToRevokeAccess(ctx, user, access)
	if err != nil {
		return nil, errs.E(op, err)
	}

	err = h.accessService.RevokeMetabaseRestrictedAccessToDataset(ctx, access)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *AccessHandler) GetAccessRequests(ctx context.Context, r *http.Request, _ any) (*service.AccessRequestsWrapper, error) {
	op := "AccessHandler.GetAccessRequests"

	id, err := uuid.Parse(r.URL.Query().Get("datasetId"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("parsing access request id: %w", err))
	}

	access, err := h.accessService.GetAccessRequests(ctx, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return access, nil
}

func (h *AccessHandler) ProcessAccessRequest(ctx context.Context, r *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "AccessHandler.ProcessAccessRequest"

	id, err := uuid.Parse(chi.URLParamFromCtx(ctx, "id"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("parsing access request id: %w", err))
	}

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	if err = h.accessService.EnsureUserIsAuthorizedToApproveRequest(ctx, user, id); err != nil {
		return nil, errs.E(errs.Unauthorized, op, fmt.Errorf("user not authorized to approve access request: %w", err))
	}

	reason := r.URL.Query().Get("reason")
	action := r.URL.Query().Get("action")

	switch action {
	case "approve":
		err := h.accessService.ApproveAccessRequest(ctx, user, id)
		if err != nil {
			return nil, errs.E(op, err)
		}

		return &transport.Empty{}, nil
	case "deny":
		err := h.accessService.DenyAccessRequest(ctx, user, id, &reason)
		if err != nil {
			return nil, errs.E(op, err)
		}

		return &transport.Empty{}, nil
	default:
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("invalid action: %s", action))
	}
}

func (h *AccessHandler) NewAccessRequest(ctx context.Context, _ *http.Request, in service.NewAccessRequestDTO) (*transport.Empty, error) {
	const op errs.Op = "AccessHandler.NewAccessRequest"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	err := h.accessService.CreateAccessRequest(ctx, user, in)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *AccessHandler) DeleteAccessRequest(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "AccessHandler.DeleteAccessRequest"

	id, err := uuid.Parse(chi.URLParamFromCtx(ctx, "id"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("parsing access request id: %w", err))
	}

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	err = h.accessService.DeleteAccessRequest(ctx, user, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *AccessHandler) UpdateAccessRequest(ctx context.Context, _ *http.Request, in service.UpdateAccessRequestDTO) (*transport.Empty, error) {
	const op errs.Op = "AccessHandler.UpdateAccessRequest"

	// FIXME: should we verify the user here
	err := h.accessService.UpdateAccessRequest(ctx, in)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func NewAccessHandler(
	service service.AccessService,
	gcpProjectID string,
) *AccessHandler {
	return &AccessHandler{
		accessService: service,
		gcpProjectID:  gcpProjectID,
	}
}
