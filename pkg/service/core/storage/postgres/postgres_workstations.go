package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/database/gensql"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.WorkstationsStorage = &workstationsStorage{}

type workstationsStorage struct {
	db *database.Repo
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

	err := s.db.Querier.CreateWorkstationsURLListChange(ctx, gensql.CreateWorkstationsURLListChangeParams{
		NavIdent:             navIdent,
		UrlList:              strings.Join(input.URLAllowList, "\n"),
		DisableGlobalUrlList: input.DisableGlobalAllowList,
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

func (s *workstationsStorage) GetLastWorkstationsURLList(ctx context.Context, navIdent string) (*service.WorkstationURLList, error) {
	const op errs.Op = "workstationsStorage.GetLastWorkstationsURLListChange"

	raw, err := s.db.Querier.GetLastWorkstationsURLListChange(ctx, navIdent)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &service.WorkstationURLList{
				URLAllowList:           []string{},
				DisableGlobalAllowList: false,
			}, nil
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return &service.WorkstationURLList{
		URLAllowList:           strings.Split(raw.UrlList, "\n"),
		DisableGlobalAllowList: raw.DisableGlobalUrlList,
	}, nil
}

func NewWorkstationsStorage(repo *database.Repo) *workstationsStorage {
	return &workstationsStorage{
		db: repo,
	}
}
