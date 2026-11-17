package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

func (c *metabaseAPI) GetDocument(ctx context.Context, id int) (*service.MetabaseDocument, error) {
	const op errs.Op = "metabaseAPI.GetDocument"

	document := service.MetabaseDocument{}

	err := c.request(ctx, http.MethodGet, fmt.Sprintf("/document/%d", id), nil, nil, &document)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &document, nil
}

func (c *metabaseAPI) CreatePublicDocumentLink(ctx context.Context, id int) (uuid.UUID, error) {
	const op errs.Op = "metabaseAPI.CreatePublicDocumentLink"

	var publicDocument struct {
		ID uuid.UUID `json:"uuid"`
	}
	if err := c.request(ctx, http.MethodPost, fmt.Sprintf("/document/%d/public-link", id), nil, nil, &publicDocument); err != nil {
		return uuid.UUID{}, errs.E(op, fmt.Errorf("creating public link for document %d: %w", id, err))
	}

	return publicDocument.ID, nil
}

func (c *metabaseAPI) DeletePublicDocumentLink(ctx context.Context, id int) error {
	const op errs.Op = "metabaseAPI.DeletePublicDocumentLink"

	if err := c.request(ctx, http.MethodDelete, fmt.Sprintf("/document/%d/public-link", id), nil, nil, nil); err != nil {
		return errs.E(op, fmt.Errorf("deleting public link for document %d: %w", id, err))
	}

	return nil
}
