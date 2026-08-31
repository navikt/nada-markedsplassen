package service

import "context"

const ArtifactRegistryEnvPrefix = "ARTIFACT_REGISTRY_"

type ArtifactKeeperAPI interface {
	ListTokens(ctx context.Context) ([]ArtifactKeeperTokenMetadata, error)
	CreateToken(ctx context.Context, request ArtifactKeeperCreateTokenRequest) (*ArtifactKeeperCreatedToken, error)
	DeleteToken(ctx context.Context, id string) error
}

type ArtifactKeeperCreateTokenRequest struct {
	Name            string
	ExpiresInDays   int
	Scopes          []string
	RepositoryMatch string
}

type ArtifactKeeperCreatedToken struct {
	ID    string
	Token string
	Name  string
}

type ArtifactKeeperTokenMetadata struct {
	ID   string
	Name string
}

type ArtifactRegistryCredential struct {
	TokenID          string
	Environment      map[string]string
	PreviousTokenIDs []string
}

type ArtifactRegistryCredentialService interface {
	Enabled() bool
	Prepare(ctx context.Context, workstationID string) (*ArtifactRegistryCredential, error)
	DeleteTokens(ctx context.Context, tokenIDs []string)
}

type DisabledArtifactRegistryCredentialService struct{}

func (DisabledArtifactRegistryCredentialService) Enabled() bool { return false }
func (DisabledArtifactRegistryCredentialService) Prepare(context.Context, string) (*ArtifactRegistryCredential, error) {
	return nil, nil
}
func (DisabledArtifactRegistryCredentialService) DeleteTokens(context.Context, []string) {}
