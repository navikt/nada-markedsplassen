package postgres

import (
	"context"

	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/database/gensql"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

const (
	defaultLimit  = 24
	defaultOffset = 0
)

var _ service.SearchStorage = &searchStorage{}

type searchStorage struct {
	db *database.Repo
}

func (s *searchStorage) Search(ctx context.Context, query *service.SearchOptions) ([]*service.SearchResultRaw, error) {
	const op errs.Op = "searchStorage.Search"

	res, err := s.db.Querier.Search(ctx, gensql.SearchParams{
		Query:   query.Text,
		Keyword: query.Keywords,
		Grp:     query.Groups,
		TeamID:  query.TeamIDs,
		Types:   query.Types,
		Lim:     int32(ptrToIntDefault(query.Limit, defaultLimit)),
		Offs:    int32(ptrToIntDefault(query.Offset, defaultOffset)),
	})
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	results := make([]*service.SearchResultRaw, 0, len(res))
	for _, r := range res {
		results = append(results, &service.SearchResultRaw{
			ElementID:   r.ElementID,
			ElementType: r.ElementType,
			Rank:        r.Rank,
			Excerpt:     r.Excerpt,
		})
	}

	return results, nil
}

func (s *searchStorage) SearchDatasets(ctx context.Context, keyword string) ([]*service.Dataset, error) {
	const op errs.Op = "searchStorage.SearchDatasets"

	res, err := s.db.Querier.SearchDatasets(ctx, keyword)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	results := make([]*service.Dataset, 0, len(res))
	for _, r := range res {
		results = append(results, &service.Dataset{
			ID:                       r.ID,
			Name:                     r.Name,
			DataproductID:            r.DataproductID,
			Description:              nullStringToPtr(r.Description),
			Created:                  r.Created,
			LastModified:             r.LastModified,
			Slug:                     r.Slug,
			Repo:                     nullStringToPtr(r.Repo),
			Pii:                      service.PiiLevel(r.Pii),
			Keywords:                 r.Keywords,
			AnonymisationDescription: nullStringToPtr(r.AnonymisationDescription),
			TargetUser:               nullStringToPtr(r.TargetUser),
		})
	}

	return results, nil
}

func NewSearchStorage(db *database.Repo) *searchStorage {
	return &searchStorage{
		db: db,
	}
}
