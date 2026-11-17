package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/database/gensql"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.MetabaseDocumentStorage = &metabaseDocumentStorage{}

type metabaseDocumentStorage struct {
	db *database.Repo
}

func (m *metabaseDocumentStorage) CreateMetabaseDocument(ctx context.Context, mbDocument *service.NewPublicMetabaseDocument) (*service.PublicMetabaseDocument, error) {
	const op errs.Op = "metabaseDocumentStorage.CreateMetabaseDocument"

	document, err := m.db.Querier.CreatePublicDocument(ctx, gensql.CreatePublicDocumentParams{
		Name:             mbDocument.Name,
		Description:      ptrToNullString(mbDocument.Input.Description),
		OwnerGroup:       mbDocument.Input.Group,
		PublicDocumentID: mbDocument.PublicDocumentID,
		MetabaseID:       mbDocument.MetabaseID,
		CreatedBy:        mbDocument.CreatorEmail,
		Keywords:         mbDocument.Input.Keywords,
		TeamkatalogenUrl: ptrToNullString(mbDocument.Input.TeamkatalogenURL),
		TeamID:           uuidPtrToNullUUID(mbDocument.Input.TeamID),
	})
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return metabaseDocumentFromSQL(&document), nil
}

func (m *metabaseDocumentStorage) UpdateMetabaseDocument(ctx context.Context, mbDocument *service.EditPublicMetabaseDocument) (*service.PublicMetabaseDocument, error) {
	const op errs.Op = "metabaseDocumentStorage.UpdateMetabaseDocument"

	document, err := m.db.Querier.UpdatePublicDocument(ctx, gensql.UpdatePublicDocumentParams{
		ID:               mbDocument.ID,
		Name:             mbDocument.Name,
		Description:      ptrToNullString(mbDocument.Input.Description),
		Keywords:         mbDocument.Input.Keywords,
		TeamkatalogenUrl: ptrToNullString(mbDocument.Input.TeamkatalogenURL),
		TeamID:           uuidPtrToNullUUID(mbDocument.Input.TeamID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.E(errs.NotExist, service.CodeDatabase, op, err)
		}
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return metabaseDocumentFromSQL(&document), nil
}

func (m *metabaseDocumentStorage) GetMetabaseDocument(ctx context.Context, id uuid.UUID) (*service.PublicMetabaseDocument, error) {
	const op errs.Op = "metabaseDocumentStorage.GetMetabaseDocument"

	document, err := m.db.Querier.GetPublicDocument(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.E(errs.NotExist, service.CodeDatabase, op, err)
		}
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return metabaseDocumentFromSQL(&document), nil
}

func (m *metabaseDocumentStorage) GetMetabaseDocumentForGroups(ctx context.Context, groups []string) ([]*service.PublicMetabaseDocument, error) {
	const op errs.Op = "metabaseDocumentStorage.GetMetabaseDocumentForGroups"

	documents, err := m.db.Querier.GetPublicDocumentsForGroups(ctx, groups)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	publicDocuments := []*service.PublicMetabaseDocument{}
	for _, document := range documents {
		publicDocuments = append(publicDocuments, metabaseDocumentFromSQL(&document))
	}

	return publicDocuments, nil
}

func (m *metabaseDocumentStorage) DeleteMetabaseDocument(ctx context.Context, id uuid.UUID) error {
	const op errs.Op = "metabaseDocumentStorage.DeleteMetabaseDocument"

	err := m.db.Querier.DeletePublicDocument(ctx, id)
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func metabaseDocumentFromSQL(document *gensql.MetabaseDocument) *service.PublicMetabaseDocument {
	return &service.PublicMetabaseDocument{
		ID:               document.ID,
		Name:             document.Name,
		Description:      nullStringToPtr(document.Description),
		Group:            document.Group,
		PublicDocumentID: document.PublicDocumentID,
		MetabaseID:       int(document.MetabaseID),
		CreatedBy:        document.CreatedBy,
		Created:          document.Created,
		LastModified:     document.LastModified,
		Keywords:         document.Keywords,
		TeamkatalogenURL: nullStringToPtr(document.TeamkatalogenUrl),
		TeamID:           nullUUIDToUUIDPtr(document.TeamID),
	}
}

func NewMetabaseDocumentStorage(db *database.Repo) *metabaseDocumentStorage {
	return &metabaseDocumentStorage{
		db: db,
	}
}
