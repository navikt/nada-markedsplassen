package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MetabaseDocumentStorage interface {
	CreateMetabaseDocument(ctx context.Context, mbDocument *NewPublicMetabaseDocument) (*PublicMetabaseDocument, error)
	UpdateMetabaseDocument(ctx context.Context, mbDocument *EditPublicMetabaseDocument) (*PublicMetabaseDocument, error)
	GetMetabaseDocument(ctx context.Context, id uuid.UUID) (*PublicMetabaseDocument, error)
	GetMetabaseDocumentForGroups(ctx context.Context, groups []string) ([]*PublicMetabaseDocument, error)
	DeleteMetabaseDocument(ctx context.Context, id uuid.UUID) error
}

type MetabaseDocumentsService interface {
	CreateMetabaseDocument(
		ctx context.Context,
		user *User,
		input PublicMetabaseDocumentInput,
	) (*PublicMetabaseDocumentOutput, error)
	GetMetabaseDocument(
		ctx context.Context,
		id uuid.UUID,
	) (*PublicMetabaseDocumentOutput, error)
	UpdateMetabaseDocument(
		ctx context.Context,
		user *User,
		id uuid.UUID,
		input PublicMetabaseDocumentEditInput,
	) (*PublicMetabaseDocumentOutput, error)
	DeleteMetabaseDocument(
		ctx context.Context,
		user *User,
		id uuid.UUID,
	) error
}

type PublicMetabaseDocumentInput struct {
	Description      *string    `json:"description,omitempty"`
	Link             string     `json:"link"`
	Keywords         []string   `json:"keywords"`
	Group            string     `json:"group"`
	TeamkatalogenURL *string    `json:"teamkatalogenURL,omitempty"`
	ProductAreaID    *uuid.UUID `json:"productAreaID,omitempty"`
	TeamID           *uuid.UUID `json:"teamID,omitempty"`
}

type PublicMetabaseDocumentEditInput struct {
	Description      *string    `json:"description,omitempty"`
	Keywords         []string   `json:"keywords"`
	TeamkatalogenURL *string    `json:"teamkatalogenURL,omitempty"`
	ProductAreaID    *uuid.UUID `json:"productAreaID,omitempty"`
	TeamID           *uuid.UUID `json:"teamID,omitempty"`
}

type NewPublicMetabaseDocument struct {
	Input            *PublicMetabaseDocumentInput
	CreatorEmail     string
	Name             string
	PublicDocumentID uuid.UUID
	MetabaseID       int32
}

type EditPublicMetabaseDocument struct {
	ID    uuid.UUID
	Input *PublicMetabaseDocumentEditInput
	Name  string
}

type PublicMetabaseDocumentOutput struct {
	ID               uuid.UUID  `json:"id"`
	Name             string     `json:"name"`
	Description      *string    `json:"description,omitempty"`
	Link             string     `json:"link"`
	MetabaseLink     string     `json:"metabaseLink"`
	Keywords         []string   `json:"keywords"`
	Group            string     `json:"group"`
	TeamkatalogenURL *string    `json:"teamkatalogenURL,omitempty"`
	TeamID           *uuid.UUID `json:"teamID,omitempty"`
	CreatedBy        string     `json:"createdBy"`
	Created          time.Time  `json:"created"`
	LastModified     time.Time  `json:"LastModified"`
}

type PublicMetabaseDocument struct {
	ID               uuid.UUID
	Name             string
	Description      *string
	Group            string
	PublicDocumentID uuid.UUID
	MetabaseID       int
	CreatedBy        string
	Created          time.Time
	LastModified     time.Time
	Keywords         []string
	TeamkatalogenURL *string
	TeamID           *uuid.UUID
}

type MetabaseDocument struct {
	ID           int    `json:"id"`
	CollectionID *int   `json:"collection_id"`
	Name         string `json:"name"`
	Archived     bool   `json:"archived"`
}
