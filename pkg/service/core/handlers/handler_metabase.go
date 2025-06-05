package handlers

import (
	"context"
	"fmt"
	"github.com/navikt/nada-backend/pkg/auth"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

type MetabaseHandler struct {
	service service.MetabaseService
}

func (h *MetabaseHandler) GetMetabaseBigQueryOpenDatasetStatus(ctx context.Context, _ *http.Request, _ any) (*service.MetabaseBigQueryDatasetStatus, error) {
	const op errs.Op = "MetabaseHandler.GetOpenMetabaseBigQueryDatabaseWorkflow"

	datasetID, err := uuid.Parse(chi.URLParamFromCtx(ctx, "id"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("parsing id: %w", err))
	}

	status, err := h.service.GetOpenMetabaseBigQueryDatabaseWorkflow(ctx, datasetID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return status, nil
}

func (h *MetabaseHandler) CreateMetabaseBigQueryOpenDataset(ctx context.Context, _ *http.Request, _ any) (*service.MetabaseBigQueryDatasetStatus, error) {
	const op errs.Op = "MetabaseHandler.CreateOpenMetabaseBigQueryDatabaseWorkflow"

	datasetID, err := uuid.Parse(chi.URLParamFromCtx(ctx, "id"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("parsing id: %w", err))
	}

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	status, err := h.service.CreateOpenMetabaseBigqueryDatabaseWorkflow(ctx, user, datasetID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return status, nil
}

func (h *MetabaseHandler) DeleteMetabaseBigQueryOpenDataset(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "MetabaseHandler.DeleteOpenMetabaseBigQueryDatabaseWorkflow"

	datasetID, err := uuid.Parse(chi.URLParamFromCtx(ctx, "id"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("parsing id: %w", err))
	}

	err = h.service.DeleteOpenMetabaseBigqueryDatabase(ctx, datasetID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *MetabaseHandler) GetMetabaseBigQueryRestrictedDatasetStatus(ctx context.Context, _ *http.Request, _ any) (*service.MetabaseBigQueryDatasetStatus, error) {
	const op errs.Op = "MetabaseHandler.GetRestrictedMetabaseBigQueryDatabaseWorkflow"

	datasetID, err := uuid.Parse(chi.URLParamFromCtx(ctx, "id"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("parsing id: %w", err))
	}

	status, err := h.service.GetRestrictedMetabaseBigQueryDatabaseWorkflow(ctx, datasetID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return status, nil
}

func (h *MetabaseHandler) CreateMetabaseBigQueryRestrictedDataset(ctx context.Context, _ *http.Request, _ any) (*service.MetabaseBigQueryDatasetStatus, error) {
	const op errs.Op = "MetabaseHandler.CreateRestrictedMetabaseBigQueryDatabaseWorkflow"

	datasetID, err := uuid.Parse(chi.URLParamFromCtx(ctx, "id"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("parsing id: %w", err))
	}

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	status, err := h.service.CreateRestrictedMetabaseBigqueryDatabaseWorkflow(ctx, user, datasetID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return status, nil
}

func (h *MetabaseHandler) DeleteMetabaseBigQueryRestrictedDataset(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "MetabaseHandler.DeleteRestrictedMetabaseBigQueryDatabaseWorkflow"

	datasetID, err := uuid.Parse(chi.URLParamFromCtx(ctx, "id"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("parsing id: %w", err))
	}

	err = h.service.DeleteRestrictedMetabaseBigqueryDatabase(ctx, datasetID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func (h *MetabaseHandler) ClearMetabaseBigqueryWorkflowJobs(ctx context.Context, _ *http.Request, _ any) (*transport.Empty, error) {
	const op errs.Op = "MetabaseHandler.ClearMetabaseBigqueryWorkflowJobs"

	datasetID, err := uuid.Parse(chi.URLParamFromCtx(ctx, "id"))
	if err != nil {
		return nil, errs.E(errs.InvalidRequest, op, fmt.Errorf("parsing id: %w", err))
	}

	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	err = h.service.ClearMetabaseBigqueryWorkflowJobs(ctx, user, datasetID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &transport.Empty{}, nil
}

func NewMetabaseHandler(service service.MetabaseService) *MetabaseHandler {
	return &MetabaseHandler{
		service: service,
	}
}
