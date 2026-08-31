package core

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

var workstationIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

var artifactRegistryOutcomes = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "nada_backend",
	Subsystem: "workstation_artifact_registry",
	Name:      "outcomes_total",
	Help:      "Workstation artifact registry outcomes.",
}, []string{"outcome"})

func ArtifactRegistryCredentialCollectors() []prometheus.Collector {
	return []prometheus.Collector{artifactRegistryOutcomes}
}

type artifactRegistryCredentialService struct {
	enabled        bool
	repositoryName string
	registryURL    string
	api            service.ArtifactKeeperAPI
	log            zerolog.Logger
	totalTimeout   time.Duration
}

func NewArtifactRegistryCredentialService(enabled bool, repositoryName, registryURL string, totalTimeout time.Duration, api service.ArtifactKeeperAPI, log zerolog.Logger) service.ArtifactRegistryCredentialService {
	return &artifactRegistryCredentialService{
		enabled: enabled, repositoryName: repositoryName, registryURL: registryURL, api: api, log: log, totalTimeout: totalTimeout,
	}
}

func (s *artifactRegistryCredentialService) Enabled() bool {
	return s.enabled
}

func (s *artifactRegistryCredentialService) Prepare(ctx context.Context, workstationID string) (*service.ArtifactRegistryCredential, error) {
	if !s.enabled {
		return nil, errors.New("artifact registry integration is disabled")
	}
	if s.api == nil {
		return nil, errors.New("Artifact Keeper API is not configured")
	}
	if !workstationIDPattern.MatchString(workstationID) {
		artifactRegistryOutcomes.WithLabelValues("invalid_workstation_id").Inc()
		return nil, errors.New("invalid workstation ID for artifact registry credential")
	}
	ctx, cancel := context.WithTimeout(ctx, s.totalTimeout)
	defer cancel()

	name := "knast:" + workstationID
	var previousTokenIDs []string
	tokens, err := s.api.ListTokens(ctx)
	if err != nil {
		s.log.Warn().Err(err).Msg("failed to list old Artifact Keeper tokens")
		artifactRegistryOutcomes.WithLabelValues("list_failed").Inc()
	} else {
		for _, token := range tokens {
			if token.Name == name {
				previousTokenIDs = append(previousTokenIDs, token.ID)
			}
		}
	}

	created, err := s.api.CreateToken(ctx, service.ArtifactKeeperCreateTokenRequest{
		Name: name, ExpiresInDays: 1, Scopes: []string{"read:artifacts"}, RepositoryMatch: s.repositoryName,
	})
	if err != nil {
		artifactRegistryOutcomes.WithLabelValues("create_failed").Inc()
		return nil, err
	}
	if created == nil || created.Name != name || created.ID == "" || created.Token == "" {
		artifactRegistryOutcomes.WithLabelValues("invalid_token_response").Inc()
		if created != nil && created.ID != "" {
			_ = s.api.DeleteToken(ctx, created.ID)
		}
		return nil, errors.New("Artifact Keeper returned an invalid token")
	}

	repositories, err := json.Marshal([]map[string]string{{"name": s.repositoryName, "url": s.registryURL}})
	if err != nil {
		return nil, errors.New("encoding artifact registry repositories")
	}
	artifactRegistryOutcomes.WithLabelValues("prepared").Inc()
	return &service.ArtifactRegistryCredential{
		TokenID:          created.ID,
		PreviousTokenIDs: previousTokenIDs,
		Environment: map[string]string{
			"ARTIFACT_REGISTRY_REPOSITORIES": string(repositories),
			"ARTIFACT_REGISTRY_USERNAME":     "__token__",
			"ARTIFACT_REGISTRY_TOKEN":        created.Token,
		},
	}, nil
}

func (s *artifactRegistryCredentialService) DeleteTokens(ctx context.Context, tokenIDs []string) {
	for _, id := range tokenIDs {
		if err := s.api.DeleteToken(ctx, id); err != nil {
			s.log.Warn().Err(err).Msg("failed to delete Artifact Keeper token")
			artifactRegistryOutcomes.WithLabelValues("cleanup_failed").Inc()
			continue
		}
		artifactRegistryOutcomes.WithLabelValues("cleanup_succeeded").Inc()
	}
}
