package core

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.TokenService = &tokenService{}

type tokenService struct {
	tokenStorage service.TokenStorage
}

func (s *tokenService) ValidateToken(ctx context.Context, token uuid.UUID) (bool, error) {
	const op errs.Op = "tokenService.ValidateToken"

	tokens, err := s.GetNadaTokens(ctx)
	if err != nil {
		return false, errs.E(op, err)
	}

	_, hasKey := tokens[token]

	return hasKey, nil
}

func (s *tokenService) GetNadaTokens(ctx context.Context) (map[uuid.UUID]string, error) {
	const op errs.Op = "tokenService.GetNadaTokens"

	tokens, err := s.tokenStorage.GetNadaTokens(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return tokens, nil
}

func (s *tokenService) GetTeamEmailFromNadaToken(ctx context.Context, token uuid.UUID) (string, error) {
	const op errs.Op = "tokenService.GetTeamEmailFromNadaToken"

	teamEmail, err := s.tokenStorage.GetTeamEmailFromNadaToken(ctx, token)
	if err != nil {
		if errs.KindIs(errs.NotExist, err) {
			return "", errs.E(errs.InvalidRequest, op, fmt.Errorf("token not found"))
		}

		return "", errs.E(op, err)
	}

	return teamEmail, nil
}

func (s *tokenService) RotateNadaToken(ctx context.Context, user *service.User, team string) error {
	const op errs.Op = "tokenService.RotateNadaToken"

	if team == "" {
		return errs.E(errs.InvalidRequest, service.CodeTeamMissing, op, fmt.Errorf("no team provided"))
	}

	groupEmail, err := s.tokenStorage.GetGroupEmailFromTeamSlug(ctx, team)
	if err != nil {
		return errs.E(op, err)
	}

	if err := ensureUserInGroup(user, groupEmail); err != nil {
		return errs.E(op, err)
	}

	err = s.tokenStorage.RotateNadaToken(ctx, team)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func NewTokenService(tokenStorage service.TokenStorage) service.TokenService {
	return &tokenService{
		tokenStorage: tokenStorage,
	}
}
