package service

import (
	"context"

	"github.com/google/uuid"
)

type MetabaseDashboardsService interface {
	CreateMetabaseDashboard(
		ctx context.Context,
		user *User,
		input NewInsightProduct,
	) (*InsightProduct, error)
	DeleteMetabaseDashboard(
		ctx context.Context,
		user *User,
		id uuid.UUID,
	) (error)
}

type NewMetabaseDashboard struct{
	URL string `json:"url"`
	Group string `json:"group"`
	Description *string `json:"description"`
}

