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

func (a *artifactKeeperAPI) ListTokens(ctx context.Context) ([]service.ArtifactKeeperTokenMetadata, error) {
	tokens, err := a.ops.ListTokens(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]service.ArtifactKeeperTokenMetadata, 0, len(tokens))
	for _, token := range tokens {
		result = append(result, service.ArtifactKeeperTokenMetadata{ID: token.ID, Name: token.Name})
	}
	return result, nil
}

func (a *artifactKeeperAPI) CreateToken(ctx context.Context, request service.ArtifactKeeperCreateTokenRequest) (*service.ArtifactKeeperCreatedToken, error) {
	token, err := a.ops.CreateToken(ctx, artifactkeeper.CreateTokenRequest{
		Name:          request.Name,
		ExpiresInDays: request.ExpiresInDays,
		Scopes:        request.Scopes,
		RepoSelector:  artifactkeeper.RepositorySelector{MatchPattern: request.RepositoryMatch},
	})
	if err != nil {
		return nil, err
	}
	return &service.ArtifactKeeperCreatedToken{ID: token.ID, Token: token.Token, Name: token.Name}, nil
}

func (a *artifactKeeperAPI) DeleteToken(ctx context.Context, id string) error {
	return a.ops.DeleteToken(ctx, id)
}
