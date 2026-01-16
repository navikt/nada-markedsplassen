package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/database/gensql"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.WorkstationsStorage = &workstationsStorage{}

type workstationsStorage struct {
	db *database.Repo
}

func (s *workstationsStorage) CreateWorkstationsActivity(ctx context.Context, navIdent, instanceID string, activity service.WorkstationActionType) error {
	const op errs.Op = "workstationsStorage.CreateWorkstationsActivity"

	err := s.db.Querier.CreateWorkstationsActivityHistory(ctx, gensql.CreateWorkstationsActivityHistoryParams{
		NavIdent: navIdent,
		InstanceID: sql.NullString{
			String: instanceID,
			Valid:  instanceID != "",
		},
		Action: string(activity),
	})
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *workstationsStorage) CreateWorkstationsConfigChange(ctx context.Context, navIdent string, config json.RawMessage) error {
	const op errs.Op = "workstationsStorage.CreateWorkstationsConfigChange"

	err := s.db.Querier.CreateWorkstationsConfigChange(ctx, gensql.CreateWorkstationsConfigChangeParams{
		NavIdent:          navIdent,
		WorkstationConfig: config,
	})
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *workstationsStorage) CreateWorkstationsOnpremAllowListChange(ctx context.Context, navIdent string, hosts []string) error {
	const op errs.Op = "workstationsStorage.CreateWorkstationsOnpremAllowlistChange"

	err := s.db.Querier.CreateWorkstationsOnpremAllowlistChange(ctx, gensql.CreateWorkstationsOnpremAllowlistChangeParams{
		NavIdent: navIdent,
		Hosts:    hosts,
	})
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *workstationsStorage) CreateWorkstationsURLListChange(ctx context.Context, navIdent string, input *service.WorkstationURLList) error {
	const op errs.Op = "workstationsStorage.CreateWorkstationsURLListChange"

	var filteredURLs []string
	for _, url := range input.URLAllowList {
		if len(url) > 0 {
			filteredURLs = append(filteredURLs, url)
		}
	}

	err := s.db.Querier.CreateWorkstationsURLListChange(ctx, gensql.CreateWorkstationsURLListChangeParams{
		NavIdent: navIdent,
		UrlList:  strings.Join(filteredURLs, "\n"),
	})
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *workstationsStorage) GetLastWorkstationsOnpremAllowList(ctx context.Context, navIdent string) ([]string, error) {
	const op errs.Op = "workstationsStorage.GetLastWorkstationsOnpremAllowlistChange"

	raw, err := s.db.Querier.GetLastWorkstationsOnpremAllowlistChange(ctx, navIdent)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []string{}, nil
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return raw.Hosts, nil
}

func (s *workstationsStorage) GetWorkstationURLListForIdent(ctx context.Context, slug string) (*service.WorkstationURLListForIdent, error) {
	const op errs.Op = "workstationsStorage.GetWorkstationURLListForIdent"

	settings, err := s.db.Querier.GetWorkstationURLListUserSettings(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err = s.db.Querier.UpdateWorkstationURLListUserSettings(ctx, gensql.UpdateWorkstationURLListUserSettingsParams{
				NavIdent:               slug,
				DisableGlobalAllowList: false,
			})
			if err != nil {
				return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
			}
			settings = gensql.WorkstationsUrlListUserSetting{
				NavIdent:               slug,
				DisableGlobalAllowList: false,
			}
		}
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	raw, err := s.db.Querier.GetWorkstationURLListForIdent(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &service.WorkstationURLListForIdent{
				NavIdent:               slug,
				DisableGlobalAllowList: settings.DisableGlobalAllowList,
			}, nil
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	urllist := []*service.WorkstationURLListItem{}
	for _, item := range raw {
		urllist = append(urllist, &service.WorkstationURLListItem{
			ID:          item.ID,
			URL:         item.Url,
			Description: item.Description,
			ExpiresAt:   item.ExpiresAt,
			CreatedAt:   item.CreatedAt,
			Duration:    item.Duration,
			Selected:    item.Selected,
		})
	}

	return &service.WorkstationURLListForIdent{
		NavIdent:               slug,
		Items:                  urllist,
		DisableGlobalAllowList: settings.DisableGlobalAllowList,
	}, nil
}

func (s *workstationsStorage) CreateWorkstationURLListItemForIdent(ctx context.Context, slug string, item *service.WorkstationURLListItem) (*service.WorkstationURLListItem, error) {
	const op errs.Op = "workstationsStorage.CreateWorkstationURLListItemForIdent"

	raw, err := s.db.Querier.CreateWorkstationURLListItemForIdent(ctx, gensql.CreateWorkstationURLListItemForIdentParams{
		NavIdent:    slug,
		Url:         item.URL,
		Description: item.Description,
		Duration:    item.Duration,
	})
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return &service.WorkstationURLListItem{
		ID:          raw.ID,
		URL:         raw.Url,
		Description: raw.Description,
		ExpiresAt:   raw.ExpiresAt,
		CreatedAt:   raw.CreatedAt,
		Duration:    raw.Duration,
		Selected:    raw.Selected,
	}, nil
}

func (s *workstationsStorage) UpdateWorkstationURLListItemForIdent(ctx context.Context, item *service.WorkstationURLListItem) (*service.WorkstationURLListItem, error) {
	const op errs.Op = "workstationsStorage.UpdateWorkstationURLListItemForIdent"

	raw, err := s.db.Querier.UpdateWorkstationURLListItemForIdent(ctx, gensql.UpdateWorkstationURLListItemForIdentParams{
		ID:          item.ID,
		Url:         item.URL,
		Description: item.Description,
		Duration:    item.Duration,
		Selected:    item.Selected,
	})
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return &service.WorkstationURLListItem{
		ID:          raw.ID,
		URL:         raw.Url,
		Description: raw.Description,
		ExpiresAt:   raw.ExpiresAt,
		CreatedAt:   raw.CreatedAt,
		Duration:    raw.Duration,
		Selected:    raw.Selected,
	}, nil
}

func (s *workstationsStorage) DeleteWorkstationURLListItemForIdent(ctx context.Context, itemID uuid.UUID) error {
	const op errs.Op = "workstationsStorage.DeleteWorkstationURLListItemForIdent"

	err := s.db.Querier.DeleteWorkstationURLListItemForIdent(ctx, itemID)
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *workstationsStorage) GetWorkstationActiveURLListForIdent(ctx context.Context, navIdent string) (*service.WorkstationActiveURLListForIdent, error) {
	const op errs.Op = "workstationStorage.GetWorkstationActiveURLListForIdent"

	raw, err := s.db.Querier.GetWorkstationActiveURLListForIdent(ctx, navIdent)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &service.WorkstationActiveURLListForIdent{
				Slug:    navIdent,
				URLList: []string{},
			}, nil
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return &service.WorkstationActiveURLListForIdent{
		Slug:                 navIdent,
		URLList:              raw.UrlListItems,
		DisableGlobalURLList: raw.DisableGlobalUrlList,
	}, nil
}

func (s *workstationsStorage) GetWorkstationURLListSettingsForIdent(ctx context.Context, navIdent string) (*service.WorkstationURLListSettings, error) {
	const op errs.Op = "workstationsStorage.GetWorkstationURLListSettingsForIdent"

	raw, err := s.db.Querier.GetWorkstationURLListUserSettings(ctx, navIdent)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &service.WorkstationURLListSettings{
				DisableGlobalAllowList: false,
			}, nil
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return &service.WorkstationURLListSettings{
		DisableGlobalAllowList: raw.DisableGlobalAllowList,
	}, nil
}

func (s *workstationsStorage) GetWorkstationURLListUsers(ctx context.Context) ([]*service.WorkstationURLListUser, error) {
	const op errs.Op = "workstationsStorage.GetWorkstationURLListUsers"

	raw, err := s.db.Querier.GetWorkstationURLListUsers(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*service.WorkstationURLListUser{}, nil
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	users := make([]*service.WorkstationURLListUser, len(raw))
	for i, user := range raw {
		users[i] = &service.WorkstationURLListUser{
			NavIdent: user,
		}
	}

	return users, nil
}

func (s *workstationsStorage) UpdateWorkstationURLListSettingsForIdent(ctx context.Context, navIdent string, opts *service.WorkstationURLListSettingsOpts) error {
	const op errs.Op = "workstationsStorage.UpdateWorkstationURLListSettings"

	_, err := s.db.Querier.UpdateWorkstationURLListUserSettings(ctx, gensql.UpdateWorkstationURLListUserSettingsParams{
		NavIdent:               navIdent,
		DisableGlobalAllowList: opts.DisableGlobalURLList,
	})
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *workstationsStorage) ActivateWorkstationURLListForIdent(ctx context.Context, urlListItemIDs []uuid.UUID) error {
	const op errs.Op = "workstationsStorage.ActivateWorkstationURLListForIdent"

	err := s.db.Querier.UpdateWorkstationURLListItemsExpiresAtForIdent(ctx, urlListItemIDs)
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *workstationsStorage) ExpireWorkstationURLListItemsForIdent(ctx context.Context, urlListItemIDs []uuid.UUID) error {
	const op errs.Op = "workstationsStorage.ExpireWorkstationURLListItemForIdent"

	err := s.db.Querier.ExpireWorkstationURLListItemsForIdent(ctx, urlListItemIDs)
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *workstationsStorage) GetLatestWorkstationURLListHistoryEntry(ctx context.Context, navIdent string) (*service.WorkstationURLListHistoryEntry, error) {
	const op errs.Op = "workstationsStorage.GetLatestWorkstationURLListHistoryEntry"

	latest, err := s.db.Querier.GetLatestWorkstationURLListHistoryEntry(ctx, navIdent)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return &service.WorkstationURLListHistoryEntry{
		ID:                   latest.ID,
		NavIdent:             latest.NavIdent,
		URLList:              latest.UrlList,
		DisableGlobalURLList: latest.DisableGlobalUrlList,
	}, nil
}

func NewWorkstationsStorage(repo *database.Repo) *workstationsStorage {
	return &workstationsStorage{
		db: repo,
	}
}
