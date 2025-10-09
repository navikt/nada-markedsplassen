package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MetabaseDashboardStorage interface {
	CreateMetabaseDashboard(ctx context.Context, mbDashboard *NewPublicMetabaseDashboard) (*PublicMetabaseDashboard, error)
	GetMetabaseDashboard(ctx context.Context, id uuid.UUID) (*PublicMetabaseDashboard, error)
	DeleteMetabaseDashboard(ctx context.Context, id uuid.UUID) error
}

type MetabaseDashboardsService interface {
	CreateMetabaseDashboard(
		ctx context.Context,
		user *User,
		input PublicMetabaseDashboardInput,
	) (*PublicMetabaseDashboardOutput, error)
	DeleteMetabaseDashboard(
		ctx context.Context,
		user *User,
		id uuid.UUID,
	) error
}

type PublicMetabaseDashboardInput struct {
	Description      *string    `json:"description,omitempty"`
	Link             string     `json:"link"`
	Keywords         []string   `json:"keywords"`
	Group            string     `json:"group"`
	TeamkatalogenURL *string    `json:"teamkatalogenURL,omitempty"`
	ProductAreaID    *uuid.UUID `json:"productAreaID,omitempty"`
	TeamID           *uuid.UUID `json:"teamID,omitempty"`
}

type NewPublicMetabaseDashboard struct {
	Input             *PublicMetabaseDashboardInput
	CreatorEmail      string
	Name              string
	PublicDashboardID uuid.UUID
	MetabaseID        int32
}

type PublicMetabaseDashboardOutput struct {
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

type PublicMetabaseDashboard struct {
	ID                uuid.UUID
	Name              string
	Description       *string
	Group             string
	PublicDashboardID uuid.UUID
	MetabaseID        int
	CreatedBy         string
	Created           time.Time
	LastModified      time.Time
	Keywords          []string
	TeamkatalogenURL  *string
	TeamID            *uuid.UUID
}
