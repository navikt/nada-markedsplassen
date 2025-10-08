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
) error {
	const op errs.Op = "metabaseDasboardsService.DeleteMetabaseDashboard"

	insightProduct, err := s.insightProductStorage.GetInsightProduct(ctx, id)
	if err != nil {
		return errs.E(op, errs.UserName(user.Email), err)
	}

	linkParts := strings.Split(insightProduct.Link, "/")
	publicDashboardUUID := linkParts[len(linkParts)-1]

	publicDashboards, err := s.metabaseAPI.GetPublicMetabaseDashboards(ctx)
	if err != nil {
		return errs.E(op, errs.UserName(user.Email), err)
	}

	dashboardID, err := findMetabaseDashboardID(publicDashboardUUID, publicDashboards)
	if err != nil {
		return errs.E(op, errs.UserName(user.Email), err)
	}

	err = s.metabaseAPI.DeletePublicDashboardLink(ctx, dashboardID)
	if err != nil {
		return errs.E(op, errs.UserName(user.Email), err)
	}

	err = s.insightProductStorage.DeleteInsightProduct(ctx, id)
	if err != nil {
		return errs.E(op, errs.UserName(user.Email), err)
	}

	return nil
}

func findMetabaseDashboardID(publicDashboardUUID string, publicDashboards []service.PublicMetabaseDashboardResponse) (int, error) {
	for _, dashboard := range publicDashboards {
		if dashboard.PublicUUID == publicDashboardUUID {
			return dashboard.ID, nil
		}
	}
	return -1, fmt.Errorf("public dashboard does not exist %s", publicDashboardUUID)
}

func (s *metabaseDashboardsService) CreateMetabaseDashboard(
	ctx context.Context,
	user *service.User,
	input service.NewPublicMetabaseDashboard,
) (*service.InsightProduct, error) {
	const op errs.Op = "metabaseDasboardsService.CreateMetabaseDashboard"
	const permissionWrite = "write"

	urlParts := strings.Split(input.Link, "/")
	dashboardID := strings.Split(urlParts[len(urlParts)-1], "-")[0]

	dashboard, err := s.metabaseAPI.GetDashboard(ctx, dashboardID)
	if err != nil {
		return nil, err
	}

	dashboardIDStr := strconv.Itoa(dashboard.CollectionID)

	permissions, err := s.metabaseAPI.GetCollectionPermissions(ctx, dashboardIDStr)
	if err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	groupIDs := []string{}

	for groupID, collectionPermission := range permissions.Groups {
		if collectionPermission[dashboardIDStr] == permissionWrite {
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

func (s *metabaseDashboardsService) checkEditPrivileges(ctx context.Context, userEmail string, groupIDs []string) error {
	const op = "metabaseDashboardsService.hasEditPrivileges"

	for _, id := range groupIDs {
		id, err := strconv.Atoi(id)
		if err != nil {
			return errs.E(op, errs.UserName(userEmail), err)
		}

		if id == service.MetabaseAllUsersGroupID { // Metabase all users group
			return nil
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
