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
	metabaseDashboardStorage service.MetabaseDashboardStorage
	metabaseAPI              service.MetabaseAPI

	host string
}

func NewMetabaseDashboardsService(
	storage service.MetabaseDashboardStorage,
	metabaseAPI service.MetabaseAPI,
	host string,
) *metabaseDashboardsService {
	return &metabaseDashboardsService{metabaseDashboardStorage: storage, metabaseAPI: metabaseAPI, host: host}
}

func (s *metabaseDashboardsService) DeleteMetabaseDashboard(
	ctx context.Context,
	user *service.User,
	id uuid.UUID,
) error {
	const op errs.Op = "metabaseDasboardsService.DeleteMetabaseDashboard"

	publicDashboard, err := s.metabaseDashboardStorage.GetMetabaseDashboard(ctx, id)
	if err != nil {
		return errs.E(op, errs.UserName(user.Email), err)
	}

	err = s.metabaseAPI.DeletePublicDashboardLink(ctx, publicDashboard.MetabaseID)
	if err != nil {
		return errs.E(op, errs.UserName(user.Email), err)
	}

	err = s.metabaseDashboardStorage.DeleteMetabaseDashboard(ctx, id)
	if err != nil {
		return errs.E(op, errs.UserName(user.Email), err)
	}

	return nil
}

func (s *metabaseDashboardsService) CreateMetabaseDashboard(
	ctx context.Context,
	user *service.User,
	input service.PublicMetabaseDashboardInput,
) (*service.PublicMetabaseDashboardOutput, error) {
	const op errs.Op = "metabaseDasboardsService.CreateMetabaseDashboard"
	const permissionWrite = "write"

	if input.Keywords == nil {
		input.Keywords = []string{}
	}

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

	publicDasboardUUID, err := s.metabaseAPI.CreatePublicDashboardLink(ctx, dashboardID)
	if err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	dashboardIDInt, err := strconv.Atoi(dashboardID)
	if err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	mbd, err := s.metabaseDashboardStorage.CreateMetabaseDashboard(ctx, &service.NewPublicMetabaseDashboard{
		Input:             &input,
		CreatorEmail:      user.Email,
		Name:              dashboard.Name,
		PublicDashboardID: publicDasboardUUID,
		MetabaseID:        int32(dashboardIDInt),
	})
	if err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	return &service.PublicMetabaseDashboardOutput{
		ID:               mbd.ID,
		Name:             mbd.Name,
		Description:      mbd.Description,
		Link:             fmt.Sprintf("%s/public/dashboard/%s", s.host, mbd.PublicDashboardID),
		Keywords:         mbd.Keywords,
		Group:            mbd.Group,
		TeamkatalogenURL: mbd.TeamkatalogenURL,
		ProductAreaID:    input.ProductAreaID,
		TeamID:           input.TeamID,
		CreatedBy:        mbd.CreatedBy,
		Created:          mbd.Created,
		LastModified:     mbd.LastModified,
	}, nil
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
