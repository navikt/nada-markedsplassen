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

	if err = ensureUserInGroup(user, publicDashboard.Group); err != nil {
		return errs.E(errs.Unauthorized, op, errs.UserName(user.Email), fmt.Errorf("unauthorized"))
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

	dashboardIDInt, err := parseMetabaseDashboardID(input.Link, s.host)
	if err != nil {
		return nil, errs.E(op, err)
	}
	dashboardID := strconv.Itoa(int(dashboardIDInt))

	if input.Keywords == nil {
		input.Keywords = []string{}
	}

	dashboard, err := s.metabaseAPI.GetDashboard(ctx, dashboardID)
	if err != nil {
		return nil, err
	}

	if err = checkCollectionWritePermissions(ctx, s.metabaseAPI, user, dashboard.CollectionID); err != nil {
		return nil, errs.E(errs.Unauthorized, op, errs.UserName(user.Email), err)
	}

	publicDasboardUUID, err := s.metabaseAPI.CreatePublicDashboardLink(ctx, dashboardID)
	if err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	mbd, err := s.metabaseDashboardStorage.CreateMetabaseDashboard(ctx, &service.NewPublicMetabaseDashboard{
		Input:             &input,
		CreatorEmail:      user.Email,
		Name:              dashboard.Name,
		PublicDashboardID: publicDasboardUUID,
		MetabaseID:        dashboardIDInt,
	})
	if err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	return PublicMetabaseDashboardToOutput(mbd, s.host), nil
}

func parseMetabaseDashboardID(link, host string) (int32, error) {
	const op errs.Op = "parseMetabaseDashboardID"
	return parseMetabaseResourceID(link, host, "dashboard", service.CodeMetabaseDashboardInvalidLink, op)
}

func (s *metabaseDashboardsService) GetMetabaseDashboard(ctx context.Context, id uuid.UUID) (*service.PublicMetabaseDashboardOutput, error) {
	const op errs.Op = "metabaseDasboardsService.GetMetabaseDashboard"

	publicDashboard, err := s.metabaseDashboardStorage.GetMetabaseDashboard(ctx, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return PublicMetabaseDashboardToOutput(publicDashboard, s.host), nil
}

func (s *metabaseDashboardsService) UpdateMetabaseDashboard(ctx context.Context, user *service.User, id uuid.UUID, input service.PublicMetabaseDashboardEditInput) (*service.PublicMetabaseDashboardOutput, error) {
	const op errs.Op = "metabaseDasboardsService.UpdateMetabaseDashboard"

	if input.Keywords == nil {
		input.Keywords = []string{}
	}

	publicDashboard, err := s.metabaseDashboardStorage.GetMetabaseDashboard(ctx, id)
	if err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	if err = ensureUserInGroup(user, publicDashboard.Group); err != nil {
		return nil, errs.E(errs.Unauthorized, op, errs.UserName(user.Email), fmt.Errorf("unauthorized"))
	}

	mbd, err := s.metabaseDashboardStorage.UpdateMetabaseDashboard(ctx, &service.EditPublicMetabaseDashboard{
		ID:    id,
		Input: &input,
		Name:  publicDashboard.Name,
	})
	if err != nil {
		return nil, errs.E(op, errs.UserName(user.Email), err)
	}

	return PublicMetabaseDashboardToOutput(mbd, s.host), nil
}

func PublicMetabaseDashboardToOutput(dashboard *service.PublicMetabaseDashboard, host string) *service.PublicMetabaseDashboardOutput {
	return &service.PublicMetabaseDashboardOutput{
		ID:               dashboard.ID,
		Name:             dashboard.Name,
		Description:      dashboard.Description,
		Link:             fmt.Sprintf("%s/public/dashboard/%s", host, dashboard.PublicDashboardID),
		MetabaseLink:     fmt.Sprintf("%s/dashboard/%d", host, dashboard.MetabaseID),
		Keywords:         dashboard.Keywords,
		Group:            dashboard.Group,
		TeamkatalogenURL: dashboard.TeamkatalogenURL,
		TeamID:           dashboard.TeamID,
		CreatedBy:        dashboard.CreatedBy,
		Created:          dashboard.Created,
		LastModified:     dashboard.LastModified,
	}
}

// checkCollectionWritePermissions verifies that the user has write access to the
// Metabase collection, and all its ancestors, before a resource is made public.
func checkCollectionWritePermissions(ctx context.Context, api service.MetabaseAPI, user *service.User, collectionID int) error {
	const permissionWrite = "write"
	const rootCollectionID = 0

	if collectionID == rootCollectionID {
		return nil
	}

	collection, err := api.GetCollection(ctx, collectionID)
	if err != nil {
		return err
	}

	collectionsIDs := []string{strconv.Itoa(collection.ID)}
	collectionsIDs = append(collectionsIDs, strings.Split(collection.Location, "/")...)

	permissions, err := api.GetCollectionPermissions(ctx)
	if err != nil {
		return err
	}

	for _, id := range collectionsIDs {
		if id != "" {
			groupIDs := []string{}
			for groupID, collectionPermission := range permissions.Groups {
				if collectionPermission[id] == permissionWrite {
					groupIDs = append(groupIDs, groupID)
				}
			}

			if err := checkEditPrivileges(ctx, api, user.Email, groupIDs); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkEditPrivileges(ctx context.Context, api service.MetabaseAPI, userEmail string, groupIDs []string) error {
	const op = "checkEditPrivileges"

	for _, id := range groupIDs {
		id, err := strconv.Atoi(id)
		if err != nil {
			return errs.E(op, errs.UserName(userEmail), err)
		}

		if id == service.MetabaseAllUsersGroupID { // Metabase all users group
			return nil
		}

		group, err := api.GetPermissionGroup(ctx, id)
		if err != nil {
			return errs.E(op, errs.UserName(userEmail), err)
		}

		for _, user := range group {
			if user.Email == userEmail {
				return nil
			}
		}
	}

	return errs.E(op, service.CodeInsufficientPrivileges, fmt.Errorf("user does not have edit privileges to metabase collection"))
}
