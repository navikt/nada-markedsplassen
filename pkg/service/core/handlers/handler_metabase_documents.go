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

type MetabaseDocumentsHandler struct {
	service service.MetabaseDocumentsService
}

func (h *MetabaseDocumentsHandler) CreateMetabaseDocument(ctx context.Context, _ *http.Request, in service.PublicMetabaseDocumentInput) (*service.PublicMetabaseDocumentOutput, error) {
	const op errs.Op = "MetabaseDocumentsHandler.CreateMetabaseDocument"

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	err := user.Validate()
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, err)
	}

	document, err := h.service.CreateMetabaseDocument(ctx, user, in)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return document, nil
}

func (h *MetabaseDocumentsHandler) GetMetabaseDocument(ctx context.Context, _ *http.Request, _ any) (*service.PublicMetabaseDocumentOutput, error) {
	const op errs.Op = "MetabaseDocumentsHandler.GetMetabaseDocument"

	id, err := uuid.Parse(chi.URLParamFromCtx(ctx, "id"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, err)
	}

	document, err := h.service.GetMetabaseDocument(ctx, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return document, nil
}

func (h *MetabaseDocumentsHandler) UpdateMetabaseDocument(ctx context.Context, _ *http.Request, in service.PublicMetabaseDocumentEditInput) (*service.PublicMetabaseDocumentOutput, error) {
	const op errs.Op = "MetabaseDocumentsHandler.UpdateMetabaseDocument"

	id, err := uuid.Parse(chi.URLParamFromCtx(ctx, "id"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, err)
	}

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	err = user.Validate()
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, err)
	}

	document, err := h.service.UpdateMetabaseDocument(ctx, user, id, in)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return document, nil
}

func (h *MetabaseDocumentsHandler) DeleteMetabaseDocument(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "MetabaseDocumentsHandler.DeleteMetabaseDocument"

	id, err := uuid.Parse(chi.URLParamFromCtx(ctx, "id"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, err)
	}

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	err = h.service.DeleteMetabaseDocument(ctx, user, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func NewMetabaseDocumentsHandler(service service.MetabaseDocumentsService) *MetabaseDocumentsHandler {
	return &MetabaseDocumentsHandler{
		service: service,
	}
}
