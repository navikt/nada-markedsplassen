package core

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.MetabaseDashboardsService = &metabaseDashboardsService{}

type metabaseDashboardsService struct {
	insightProductStorage service.InsightProductStorage
	metabaseAPI           service.MetabaseAPI
}

func NewMetabaseDashboardsService(
	storage service.InsightProductStorage,
	metabaseAPI service.MetabaseAPI,
) *metabaseDashboardsService {
	return &metabaseDashboardsService{insightProductStorage: storage, metabaseAPI: metabaseAPI}
}

func (s *metabaseDashboardsService) DeleteMetabaseDashboard(
	ctx context.Context,
	user *service.User,
	id uuid.UUID,
) (error) {
	return nil
}

func (s *metabaseDashboardsService) CreateMetabaseDashboard(
	ctx context.Context,
	user *service.User,
	input service.NewInsightProduct,
) (*service.InsightProduct, error) {
	const op errs.Op = "metabaseDasboardsService.CreateMetabaseDashboard"
	const permissionWrite = "write"

	urlParts := strings.Split(input.Link, "/")
	dashboardID := strings.Split(urlParts[len(urlParts) - 1], "-")[0]

	dashboard, err := s.metabaseAPI.GetDashboard(ctx, dashboardID)
	permissions, err := s.metabaseAPI.GetCollectionPermissions(ctx, dashboard.CollectionID)

	groupIDs := []string{}

	for groupID, collectionPermission := range permissions.Groups {
		if collectionPermission[dashboard.CollectionID] == permissionWrite {
			groupIDs = append(groupIDs, groupID)
		}
	}

	if err := s.checkEditPrivileges(ctx, user.Email, groupIDs); err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	publicDasboardURL, err := s.metabaseAPI.CreatePublicDashboardLink(ctx, dashboardID)
	if err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	input.Link = publicDasboardURL

	ip, err := s.insightProductStorage.CreateInsightProduct(ctx, user.Email, input)

	if err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	return ip, nil
}


func (s *metabaseDashboardsService) checkEditPrivileges(ctx context.Context, userEmail string, groupIDs []string) (error) {
	const op = "metabaseDashboardsService.hasEditPrivileges"

	for _, id := range groupIDs {
		id, err := strconv.Atoi(id)
		if err != nil {
			return errs.E(op, errs.UserName(userEmail), err)
		}

		group, err := s.metabaseAPI.GetPermissionGroup(ctx, id)
		if err != nil {
			return errs.E(op, errs.UserName(userEmail), err)
		}

		for _, user := range group {
			if user.Email == userEmail {
				return nil
			}
		}
	}

	return errs.E(op, service.CodeInsufficientPrivileges, fmt.Errorf("user does not have edit privileges to metabase dashboard"))
}
