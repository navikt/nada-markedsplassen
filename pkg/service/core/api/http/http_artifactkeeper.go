package http

import (
	"context"

	"github.com/navikt/nada-backend/pkg/artifactkeeper"
	"github.com/navikt/nada-backend/pkg/service"
)

type artifactKeeperAPI struct {
	ops artifactkeeper.Operations
}

func NewArtifactKeeperAPI(ops artifactkeeper.Operations) service.ArtifactKeeperAPI {
	return &artifactKeeperAPI{ops: ops}
}

func (a *artifactKeeperAPI) CreateToken(ctx context.Context, request service.ArtifactKeeperCreateTokenRequest) (*service.ArtifactKeeperCreatedToken, error) {
	token, err := a.ops.CreateToken(ctx, artifactkeeper.CreateTokenRequest{
		Name:          request.Name,
		ExpiresInDays: request.ExpiresInDays,
		Scopes:        request.Scopes,
		RepoSelector:  artifactkeeper.RepositorySelector{MatchLabels: request.MatchLabels},
	})
	if err != nil {
		return nil, err
	}
	return &service.ArtifactKeeperCreatedToken{ID: token.ID, Token: token.Token, Name: token.Name}, nil
}
