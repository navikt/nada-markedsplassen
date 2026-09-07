package service

import "context"

const ArtifactRegistryEnvPrefix = "ARTIFACT_REGISTRY_"

type ArtifactKeeperAPI interface {
	CreateToken(ctx context.Context, request ArtifactKeeperCreateTokenRequest) (*ArtifactKeeperCreatedToken, error)
}

type ArtifactKeeperCreateTokenRequest struct {
	Name          string
	ExpiresInDays int
	Scopes        []string
	MatchLabels   map[string]string
}

type ArtifactKeeperCreatedToken struct {
	ID    string
	Token string
	Name  string
}

type ArtifactRegistryCredential struct {
	Environment map[string]string
}

type ArtifactRegistryCredentialService interface {
	Enabled() bool
	Prepare(ctx context.Context, workstationID string) (*ArtifactRegistryCredential, error)
}

type DisabledArtifactRegistryCredentialService struct{}

func (DisabledArtifactRegistryCredentialService) Enabled() bool { return false }
func (DisabledArtifactRegistryCredentialService) Prepare(context.Context, string) (*ArtifactRegistryCredential, error) {
	return nil, nil
}
