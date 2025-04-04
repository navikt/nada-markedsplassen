package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.TokenStorage = &tokenStorage{}

type tokenStorage struct {
	db *database.Repo
}

func (s *tokenStorage) GetNadaTokenFromGroupEmail(ctx context.Context, groupEmail string) (uuid.UUID, error) {
	const op errs.Op = "tokenStorage.GetNadaTokenFromGroupEmail"

	token, err := s.db.Querier.GetNadaTokenFromGroupEmail(ctx, groupEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.UUID{}, errs.E(errs.NotExist, service.CodeDatabase, op, err, service.ParamGroupEmail)
		}

		return uuid.UUID{}, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return token, nil
}

func (s *tokenStorage) RotateNadaToken(ctx context.Context, team string) error {
	const op errs.Op = "tokenStorage.RotateNadaToken"

	err := s.db.Querier.RotateNadaToken(ctx, team)
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *tokenStorage) GetNadaTokensForTeams(ctx context.Context, teams []string) ([]service.NadaToken, error) {
	const op errs.Op = "tokenStorage.GetNadaTokensForTeams"

	rawTokens, err := s.db.Querier.GetNadaTokensForTeams(ctx, teams)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	tokens := make([]service.NadaToken, len(rawTokens))
	for i, t := range rawTokens {
		tokens[i] = service.NadaToken{
			Team:  t.Team,
			Token: t.Token,
		}
	}

	return tokens, nil
}

func (s *tokenStorage) GetTeamEmailFromNadaToken(ctx context.Context, token uuid.UUID) (string, error) {
	const op errs.Op = "tokenStorage.GetTeamEmailFromNadaToken"

	email, err := s.db.Querier.GetTeamEmailFromNadaToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errs.E(errs.NotExist, service.CodeDatabase, op, fmt.Errorf("token not found"), service.ParamToken)
		}

		return "", errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return email, nil
}

func (s *tokenStorage) GetNadaTokens(ctx context.Context) (map[uuid.UUID]string, error) {
	const op errs.Op = "tokenStorage.GetNadaTokens"

	rawTokens, err := s.db.Querier.GetNadaTokens(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	tokens := map[uuid.UUID]string{}
	for _, t := range rawTokens {
		tokens[t.Token] = t.Team
	}

	return tokens, nil
}

func (s *tokenStorage) GetGroupEmailFromTeamSlug(ctx context.Context, teamSlug string) (string, error) {
	const op errs.Op = "tokenStorage.GetGroupEmailFromTeamSlug"

	email, err := s.db.Querier.GetGroupEmailFromTeamSlug(ctx, teamSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errs.E(errs.NotExist, service.CodeDatabase, op, fmt.Errorf("team not found"), service.ParamTeam)
		}

		return "", errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return email, nil
}

func NewTokenStorage(db *database.Repo) *tokenStorage {
	return &tokenStorage{
		db: db,
	}
}
