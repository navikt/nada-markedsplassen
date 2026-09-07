package core

import (
	"context"
	"testing"
	"time"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/stretchr/testify/require"
)

type artifactKeeperStub struct {
	created          service.ArtifactKeeperCreatedToken
	createRequest    service.ArtifactKeeperCreateTokenRequest
	returnNilCreated bool
}

func (s *artifactKeeperStub) CreateToken(_ context.Context, request service.ArtifactKeeperCreateTokenRequest) (*service.ArtifactKeeperCreatedToken, error) {
	s.createRequest = request
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
	sut := NewArtifactRegistryCredentialService(true, "knast-default", time.Second, stub)

	credential, err := sut.Prepare(context.Background(), "test-knast")
	require.Error(t, err)
	require.Nil(t, credential)
}

func TestArtifactRegistryCredentialServiceRejectsInvalidCreatedToken(t *testing.T) {
	t.Parallel()
	stub := &artifactKeeperStub{created: service.ArtifactKeeperCreatedToken{ID: "invalid", Name: "knast:test-knast"}}
	sut := NewArtifactRegistryCredentialService(true, "knast-default", time.Second, stub)

	credential, err := sut.Prepare(context.Background(), "test-knast")
	require.Error(t, err)
	require.Nil(t, credential)
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
			sut := NewArtifactRegistryCredentialService(true, "knast-default", time.Second, stub)

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
func TestArtifactRegistryCredentialServiceCreatesToken(t *testing.T) {
	t.Parallel()
	stub := &artifactKeeperStub{
		created: service.ArtifactKeeperCreatedToken{ID: "new", Name: "knast:test-knast", Token: "marker-secret"},
	}
	sut := NewArtifactRegistryCredentialService(true, "knast-default", time.Second, stub)

	credential, err := sut.Prepare(context.Background(), "test-knast")
	require.NoError(t, err)
	require.Equal(t, "marker-secret", credential.Environment["ARTIFACT_REGISTRY_TOKEN"])
	require.Len(t, credential.Environment, 1)
	require.Equal(t, service.ArtifactKeeperCreateTokenRequest{
		Name:          "knast:test-knast",
		ExpiresInDays: 1,
		Scopes:        []string{"read:artifacts"},
		MatchLabels:   map[string]string{"knast-default": "true"},
	}, stub.createRequest)
}

func TestWithoutArtifactRegistryEnv(t *testing.T) {
	t.Parallel()
	got := withoutArtifactRegistryEnv(map[string]string{
		"KEEP": "value", "ARTIFACT_REGISTRY_TOKEN": "marker-secret", "ARTIFACT_REGISTRY_USERNAME": "__token__",
	})
	require.Equal(t, map[string]string{"KEEP": "value"}, got)
}
