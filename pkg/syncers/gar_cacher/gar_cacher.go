package gar_cacher

import (
	"context"
	"fmt"

	"github.com/navikt/nada-backend/pkg/cache"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/syncers"
	"github.com/rs/zerolog"
)

var _ syncers.Runner = &Runner{}

type Runner struct {
	garAPI   service.ArtifactRegistryAPI
	garCache cache.Cacher

	project    string
	location   string
	repository string
}

func New(project, location, repository string, garAPI service.ArtifactRegistryAPI, garCache cache.Cacher) *Runner {
	return &Runner{
		garAPI:     garAPI,
		garCache:   garCache,
		project:    project,
		location:   location,
		repository: repository,
	}
}

func (r *Runner) Name() string {
	return "GARCacher"
}

func (r *Runner) RunOnce(ctx context.Context, _ zerolog.Logger) error {
	containerImages, err := r.garAPI.ListContainerImagesWithTag(ctx, &service.ContainerRepositoryIdentifier{
		Project:    r.project,
		Location:   r.location,
		Repository: r.repository,
	}, "latest")
	if err != nil {
		return fmt.Errorf("listing GAR container images with tag: %w", err)
	}

	key := fmt.Sprintf("artifactregistry:%s:%s:%s", r.project, r.location, r.repository)

	r.garCache.Set(key, containerImages)

	return nil
}
