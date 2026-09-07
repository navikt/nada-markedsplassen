package core

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
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
	enabled      bool
	accessLabel  string
	api          service.ArtifactKeeperAPI
	totalTimeout time.Duration
}

func NewArtifactRegistryCredentialService(enabled bool, accessLabel string, totalTimeout time.Duration, api service.ArtifactKeeperAPI) service.ArtifactRegistryCredentialService {
	return &artifactRegistryCredentialService{
		enabled: enabled, accessLabel: accessLabel, api: api, totalTimeout: totalTimeout,
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
	created, err := s.api.CreateToken(ctx, service.ArtifactKeeperCreateTokenRequest{
		Name: name, ExpiresInDays: 1, Scopes: []string{"read:artifacts"}, MatchLabels: map[string]string{s.accessLabel: "true"},
	})
	if err != nil {
		artifactRegistryOutcomes.WithLabelValues("create_failed").Inc()
		return nil, err
	}
	if created == nil || created.Name != name || created.Token == "" {
		artifactRegistryOutcomes.WithLabelValues("invalid_token_response").Inc()
		return nil, errors.New("Artifact Keeper returned an invalid token")
	}

	artifactRegistryOutcomes.WithLabelValues("prepared").Inc()
	return &service.ArtifactRegistryCredential{
		Environment: map[string]string{
			"ARTIFACT_REGISTRY_TOKEN": created.Token,
		},
	}, nil
}
