package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/auth"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
)

type MetabaseDashboardsHandler struct {
	service service.MetabaseDashboardsService
}

func (h *MetabaseDashboardsHandler) CreateMetabaseDashboard(ctx context.Context, _ *http.Request, in service.NewInsightProduct) (*service.InsightProduct, error) {
	const op errs.Op = "MetabaseDashboardsHandler.CreateMetabaseDashboard"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	err := user.Validate()
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, err)
	}

	insightProduct, err := h.service.CreateMetabaseDashboard(ctx, user, in)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return insightProduct, nil
}

func (h *MetabaseDashboardsHandler) DeleteMetabaseDashboard(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "MetabaseDashboardsHandler.DeleteMetabaseDashboard"

	id, err := uuid.Parse(chi.URLParamFromCtx(ctx, "id"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, err)
	}

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	err = h.service.DeleteMetabaseDashboard(ctx, user, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func NewMetabaseDashboardsHandler(service service.MetabaseDashboardsService) *MetabaseDashboardsHandler {
	return &MetabaseDashboardsHandler{
		service: service,
	}
}
