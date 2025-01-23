package handlers

import (
	"context"
	"net/http"

	"github.com/navikt/nada-backend/pkg/errs"

	"github.com/navikt/nada-backend/pkg/service"
)

type OnpremMappingHandler struct {
	onpremMappingService service.OnpremMappingService
}

func (h *OnpremMappingHandler) GetClassifiedHosts(ctx context.Context, r *http.Request, _ any) (*service.ClassifiedHosts, error) {
	const op errs.Op = "OnpremMappingHandler.GetClassifiedHosts"

	result, err := h.onpremMappingService.GetClassifiedHosts(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return result, nil
}

func NewOnpremMappingHandler(s service.OnpremMappingService) *OnpremMappingHandler {
	return &OnpremMappingHandler{onpremMappingService: s}
}
