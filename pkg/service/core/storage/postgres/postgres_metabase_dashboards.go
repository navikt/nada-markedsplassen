package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/database/gensql"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.MetabaseDashboardStorage = &metabaseDashboardStorage{}

type metabaseDashboardStorage struct {
	db *database.Repo
}

func (m *metabaseDashboardStorage) CreateMetabaseDashboard(ctx context.Context, mbDashboard *service.NewPublicMetabaseDashboard) (*service.PublicMetabaseDashboard, error) {
	dashboard, err := m.db.Querier.CreatePublicDashboard(ctx, gensql.CreatePublicDashboardParams{
		Name:              mbDashboard.Name,
		Description:       ptrToNullString(mbDashboard.Input.Description),
		OwnerGroup:        mbDashboard.Input.Group,
		CreatedBy:         mbDashboard.CreatorEmail,
		PublicDashboardID: mbDashboard.PublicDashboardID,
		MetabaseID:        mbDashboard.MetabaseID,
		Keywords:          mbDashboard.Input.Keywords,
		TeamkatalogenUrl:  ptrToNullString(mbDashboard.Input.TeamkatalogenURL),
		TeamID:            uuidPtrToNullUUID(mbDashboard.Input.TeamID),
	})
	if err != nil {
		return nil, err
	}

	return metabaseDashboardFromSQL(&dashboard), nil
}

func (m *metabaseDashboardStorage) GetMetabaseDashboard(ctx context.Context, id uuid.UUID) (*service.PublicMetabaseDashboard, error) {
	dashboard, err := m.db.Querier.GetPublicDashboard(ctx, id)
	if err != nil {
		return nil, err
	}

	return metabaseDashboardFromSQL(&dashboard), nil
}

func (m *metabaseDashboardStorage) GetMetabaseDashboardForGroups(ctx context.Context, groups []string) ([]*service.PublicMetabaseDashboard, error) {
	dashboards, err := m.db.Querier.GetPublicDashboardsForGroups(ctx, groups)
	if err != nil {
		return nil, err
	}

	publicDashboards := []*service.PublicMetabaseDashboard{}
	for _, dashboard := range dashboards {
		publicDashboards = append(publicDashboards, metabaseDashboardFromSQL(&dashboard))
	}

	return publicDashboards, nil
}

func (m *metabaseDashboardStorage) DeleteMetabaseDashboard(ctx context.Context, id uuid.UUID) error {
	return m.db.Querier.DeletePublicDashboard(ctx, id)
}

func metabaseDashboardFromSQL(dashboard *gensql.MetabaseDashboard) *service.PublicMetabaseDashboard {
	return &service.PublicMetabaseDashboard{
		ID:                dashboard.ID,
		PublicDashboardID: dashboard.PublicDashboardID,
		MetabaseID:        int(dashboard.MetabaseID),
		Name:              dashboard.Name,
		Description:       nullStringToPtr(dashboard.Description),
		Keywords:          dashboard.Keywords,
		Group:             dashboard.Group,
		TeamkatalogenURL:  nullStringToPtr(dashboard.TeamkatalogenUrl),
		CreatedBy:         dashboard.CreatedBy,
		Created:           dashboard.Created,
		LastModified:      dashboard.LastModified,
		TeamID:            nullUUIDToUUIDPtr(dashboard.TeamID),
	}
}

func NewMetabaseDashboardStorage(db *database.Repo) *metabaseDashboardStorage {
	return &metabaseDashboardStorage{
		db: db,
	}
}
