package core

import (
	"context"
	"testing"
	"time"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

type artifactKeeperStub struct {
	listed           []service.ArtifactKeeperTokenMetadata
	created          service.ArtifactKeeperCreatedToken
	deleted          []string
	returnNilCreated bool
}

func (s *artifactKeeperStub) ListTokens(context.Context) ([]service.ArtifactKeeperTokenMetadata, error) {
	return s.listed, nil
}
func (s *artifactKeeperStub) CreateToken(_ context.Context, request service.ArtifactKeeperCreateTokenRequest) (*service.ArtifactKeeperCreatedToken, error) {
	if s.returnNilCreated {
		return nil, nil
	}
	if request.Name != s.created.Name {
		return nil, context.Canceled
	}
	return &s.created, nil
}

func TestArtifactRegistryCredentialServiceRejectsNilCreateResponse(t *testing.T) {
	t.Parallel()
	stub := &artifactKeeperStub{returnNilCreated: true}
	sut := NewArtifactRegistryCredentialService(true, "knast-pypi", "https://registry.example/pypi/knast-pypi/simple/", time.Second, stub, zerolog.Nop())

	credential, err := sut.Prepare(context.Background(), "test-knast")
	require.Error(t, err)
	require.Nil(t, credential)
}

func TestArtifactRegistryCredentialServiceDeletesInvalidCreatedToken(t *testing.T) {
	t.Parallel()
	stub := &artifactKeeperStub{created: service.ArtifactKeeperCreatedToken{ID: "invalid", Name: "knast:test-knast"}}
	sut := NewArtifactRegistryCredentialService(true, "knast-pypi", "https://registry.example/pypi/knast-pypi/simple/", time.Second, stub, zerolog.Nop())

	credential, err := sut.Prepare(context.Background(), "test-knast")
	require.Error(t, err)
	require.Nil(t, credential)
	require.Equal(t, []string{"invalid"}, stub.deleted)
}

func TestArtifactRegistryCredentialServiceValidatesWorkstationID(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		workstationID string
		valid         bool
	}{
		{name: "Nav-ident", workstationID: "p123456", valid: true},
		{name: "letters and hyphen", workstationID: "test-knast", valid: true},
		{name: "single letter", workstationID: "a", valid: true},
		{name: "starts with digit", workstationID: "1test", valid: false},
		{name: "uppercase", workstationID: "P123456", valid: false},
		{name: "colon", workstationID: "test:knast", valid: false},
		{name: "slash", workstationID: "test/knast", valid: false},
		{name: "wildcard", workstationID: "test*", valid: false},
		{name: "empty", workstationID: "", valid: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			stub := &artifactKeeperStub{created: service.ArtifactKeeperCreatedToken{
				ID: "new", Name: "knast:" + tc.workstationID, Token: "marker-secret",
			}}
			sut := NewArtifactRegistryCredentialService(true, "knast-pypi", "https://registry.example/pypi/knast-pypi/simple/", time.Second, stub, zerolog.Nop())

			credential, err := sut.Prepare(context.Background(), tc.workstationID)
			if tc.valid {
				require.NoError(t, err)
				require.NotNil(t, credential)
			} else {
				require.Error(t, err)
				require.Nil(t, credential)
			}
		})
	}
}
func (s *artifactKeeperStub) DeleteToken(_ context.Context, id string) error {
	s.deleted = append(s.deleted, id)
	return nil
}

func TestArtifactRegistryCredentialServiceSnapshotsOldTokens(t *testing.T) {
	t.Parallel()
	stub := &artifactKeeperStub{
		listed:  []service.ArtifactKeeperTokenMetadata{{ID: "old-1", Name: "knast:test-knast"}, {ID: "other", Name: "knast:other"}, {ID: "old-2", Name: "knast:test-knast"}},
		created: service.ArtifactKeeperCreatedToken{ID: "new", Name: "knast:test-knast", Token: "marker-secret"},
	}
	sut := NewArtifactRegistryCredentialService(true, "knast-pypi", "https://registry.example/pypi/knast-pypi/simple/", time.Second, stub, zerolog.Nop())

	credential, err := sut.Prepare(context.Background(), "test-knast")
	require.NoError(t, err)
	require.Equal(t, []string{"old-1", "old-2"}, credential.PreviousTokenIDs)
	require.Equal(t, "marker-secret", credential.Environment["ARTIFACT_REGISTRY_TOKEN"])
	require.Equal(t, "__token__", credential.Environment["ARTIFACT_REGISTRY_USERNAME"])
	require.JSONEq(t, `[{"name":"knast-pypi","url":"https://registry.example/pypi/knast-pypi/simple/"}]`, credential.Environment["ARTIFACT_REGISTRY_REPOSITORIES"])

	sut.DeleteTokens(context.Background(), credential.PreviousTokenIDs)
	require.Equal(t, []string{"old-1", "old-2"}, stub.deleted)
}

func TestWithoutArtifactRegistryEnv(t *testing.T) {
	t.Parallel()
	got := withoutArtifactRegistryEnv(map[string]string{
		"KEEP": "value", "ARTIFACT_REGISTRY_TOKEN": "marker-secret", "ARTIFACT_REGISTRY_USERNAME": "__token__",
	})
	require.Equal(t, map[string]string{"KEEP": "value"}, got)
}
