package core

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.MetabaseDocumentsService = &metabaseDocumentsService{}

type metabaseDocumentsService struct {
	metabaseDocumentStorage service.MetabaseDocumentStorage
	metabaseAPI             service.MetabaseAPI

	host string
}

func NewMetabaseDocumentsService(
	storage service.MetabaseDocumentStorage,
	metabaseAPI service.MetabaseAPI,
	host string,
) *metabaseDocumentsService {
	return &metabaseDocumentsService{metabaseDocumentStorage: storage, metabaseAPI: metabaseAPI, host: host}
}

func (s *metabaseDocumentsService) DeleteMetabaseDocument(
	ctx context.Context,
	user *service.User,
	id uuid.UUID,
) error {
	const op errs.Op = "metabaseDocumentsService.DeleteMetabaseDocument"

	publicDocument, err := s.metabaseDocumentStorage.GetMetabaseDocument(ctx, id)
	if err != nil {
		return errs.E(op, errs.UserName(user.Email), err)
	}

	if err = ensureUserInGroup(user, publicDocument.Group); err != nil {
		return errs.E(errs.Unauthorized, op, errs.UserName(user.Email), fmt.Errorf("unauthorized"))
	}

	// Revoke the public link before removing the record, so a failure never
	// leaves a public document without an owner in Nada.
	err = s.metabaseAPI.DeletePublicDocumentLink(ctx, publicDocument.MetabaseID)
	if err != nil {
		return errs.E(op, errs.UserName(user.Email), err)
	}

	err = s.metabaseDocumentStorage.DeleteMetabaseDocument(ctx, id)
	if err != nil {
		return errs.E(op, errs.UserName(user.Email), err)
	}

	return nil
}

func (s *metabaseDocumentsService) CreateMetabaseDocument(
	ctx context.Context,
	user *service.User,
	input service.PublicMetabaseDocumentInput,
) (*service.PublicMetabaseDocumentOutput, error) {
	const op errs.Op = "metabaseDocumentsService.CreateMetabaseDocument"

	documentID, err := parseMetabaseDocumentID(input.Link, s.host)
	if err != nil {
		return nil, errs.E(op, err)
	}

	if err = ensureUserInGroup(user, input.Group); err != nil {
		return nil, errs.E(op, err)
	}

	if input.Keywords == nil {
		input.Keywords = []string{}
	}

	document, err := s.metabaseAPI.GetDocument(ctx, int(documentID))
	if err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	// Metabase represents the root collection as null for documents.
	collectionID := 0
	if document.CollectionID != nil {
		collectionID = *document.CollectionID
	}

	if err = checkCollectionWritePermissions(ctx, s.metabaseAPI, user, collectionID); err != nil {
		return nil, errs.E(errs.Unauthorized, op, errs.UserName(user.Email), err)
	}

	publicDocumentUUID, err := s.metabaseAPI.CreatePublicDocumentLink(ctx, int(documentID))
	if err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	mbd, err := s.metabaseDocumentStorage.CreateMetabaseDocument(ctx, &service.NewPublicMetabaseDocument{
		Input:            &input,
		CreatorEmail:     user.Email,
		Name:             document.Name,
		PublicDocumentID: publicDocumentUUID,
		MetabaseID:       documentID,
	})
	if err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	return PublicMetabaseDocumentToOutput(mbd, s.host), nil
}

func parseMetabaseDocumentID(link, host string) (int32, error) {
	const op errs.Op = "parseMetabaseDocumentID"
	return parseMetabaseResourceID(link, host, "document", service.CodeMetabaseDocumentInvalidLink, op)
}

func (s *metabaseDocumentsService) GetMetabaseDocument(ctx context.Context, id uuid.UUID) (*service.PublicMetabaseDocumentOutput, error) {
	const op errs.Op = "metabaseDocumentsService.GetMetabaseDocument"

	publicDocument, err := s.metabaseDocumentStorage.GetMetabaseDocument(ctx, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return PublicMetabaseDocumentToOutput(publicDocument, s.host), nil
}

func (s *metabaseDocumentsService) UpdateMetabaseDocument(ctx context.Context, user *service.User, id uuid.UUID, input service.PublicMetabaseDocumentEditInput) (*service.PublicMetabaseDocumentOutput, error) {
	const op errs.Op = "metabaseDocumentsService.UpdateMetabaseDocument"

	if input.Keywords == nil {
		input.Keywords = []string{}
	}

	publicDocument, err := s.metabaseDocumentStorage.GetMetabaseDocument(ctx, id)
	if err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	if err = ensureUserInGroup(user, publicDocument.Group); err != nil {
		return nil, errs.E(errs.Unauthorized, op, errs.UserName(user.Email), fmt.Errorf("unauthorized"))
	}

	mbd, err := s.metabaseDocumentStorage.UpdateMetabaseDocument(ctx, &service.EditPublicMetabaseDocument{
		ID:    id,
		Input: &input,
		Name:  publicDocument.Name,
	})
	if err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	return PublicMetabaseDocumentToOutput(mbd, s.host), nil
}

func PublicMetabaseDocumentToOutput(document *service.PublicMetabaseDocument, host string) *service.PublicMetabaseDocumentOutput {
	return &service.PublicMetabaseDocumentOutput{
		ID:               document.ID,
		Name:             document.Name,
		Description:      document.Description,
		Link:             fmt.Sprintf("%s/public/document/%s", host, document.PublicDocumentID),
		MetabaseLink:     fmt.Sprintf("%s/document/%d", host, document.MetabaseID),
		Keywords:         document.Keywords,
		Group:            document.Group,
		TeamkatalogenURL: document.TeamkatalogenURL,
		TeamID:           document.TeamID,
		CreatedBy:        document.CreatedBy,
		Created:          document.Created,
		LastModified:     document.LastModified,
	}
}
