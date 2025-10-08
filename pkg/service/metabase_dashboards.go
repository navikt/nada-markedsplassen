package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MetabaseDashboardsService interface {
	CreateMetabaseDashboard(
		ctx context.Context,
		user *User,
		input NewPublicMetabaseDashboard,
	) (*InsightProduct, error)
	DeleteMetabaseDashboard(
		ctx context.Context,
		user *User,
		id uuid.UUID,
	) error
}

type NewPublicMetabaseDashboard struct {
	Description      *string    `json:"description,omitempty"`
	Link             string     `json:"link"`
	Keywords         []string   `json:"keywords"`
	Group            string     `json:"group"`
	TeamkatalogenURL *string    `json:"teamkatalogenURL,omitempty"`
	ProductAreaID    *uuid.UUID `json:"productAreaID,omitempty"`
	TeamID           *uuid.UUID `json:"teamID,omitempty"`
}

type PublicMetabaseDashboard struct {
	ID               uuid.UUID  `json:"id"`
	Name             string     `json:"name"`
	Description      *string    `json:"description,omitempty"`
	Link             string     `json:"link"`
	Keywords         []string   `json:"keywords"`
	Group            string     `json:"group"`
	TeamkatalogenURL *string    `json:"teamkatalogenURL,omitempty"`
	ProductAreaID    *uuid.UUID `json:"productAreaID,omitempty"`
	TeamID           *uuid.UUID `json:"teamID,omitempty"`
	CreatedBy        string     `json:"createdBy"`
	Created          time.Time  `json:"created"`
	LastModified     time.Time  `json:"LastModified"`
}
