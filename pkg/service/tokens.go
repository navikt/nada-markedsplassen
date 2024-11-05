package service

import (
	"context"

	"github.com/google/uuid"
)

type TokenStorage interface {
	GetNadaTokensForTeams(ctx context.Context, teams []string) ([]NadaToken, error)
	GetTeamEmailFromNadaToken(ctx context.Context, token uuid.UUID) (string, error)
	GetNadaTokens(ctx context.Context) (map[uuid.UUID]string, error)
	GetNadaTokenFromGroupEmail(ctx context.Context, groupEmail string) (uuid.UUID, error)
	RotateNadaToken(ctx context.Context, team string) error
}

type TokenService interface {
	RotateNadaToken(ctx context.Context, user *User, team string) error
	GetTeamEmailFromNadaToken(ctx context.Context, token uuid.UUID) (string, error)
	GetNadaTokens(ctx context.Context) (map[uuid.UUID]string, error)
	ValidateToken(ctx context.Context, token uuid.UUID) (bool, error)
}

type NadaToken struct {
	Team  string    `json:"team"`
	Token uuid.UUID `json:"token"`
}
