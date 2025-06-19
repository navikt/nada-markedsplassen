package postgres

import (
	"context"
	"fmt"

	"github.com/navikt/nada-backend/pkg/cache"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.ArtifactRegistryAPI = &artifactRegistryCache{}

type artifactRegistryCache struct {
	api   service.ArtifactRegistryAPI
	cache cache.Cacher
}

func (a *artifactRegistryCache) ListContainerImagesWithTag(ctx context.Context, id *service.ContainerRepositoryIdentifier, tag string) ([]*service.ContainerImage, error) {
	const op errs.Op = "artifactRegistryCache.ListContainerImagesWithTag"

	key := fmt.Sprintf("artifactregistry:%s:%s:%s", id.Project, id.Location, id.Repository)

	containerImages := []*service.ContainerImage{}
	valid := a.cache.Get(key, &containerImages)
	if valid {
		return containerImages, nil
	}

	containerImages, err := a.api.ListContainerImagesWithTag(ctx, id, tag)
	if err != nil {
		return nil, errs.E(op, err)
	}

	a.cache.Set(key, containerImages)

	return containerImages, nil
}

func (c *artifactRegistryCache) AddArtifactRegistryPolicyBinding(ctx context.Context, id *service.ContainerRepositoryIdentifier, binding *service.Binding) error {
	const op errs.Op = "artifactRegistryCache.AddArtifactRegistryPolicyBinding"

	err := c.api.AddArtifactRegistryPolicyBinding(ctx, id, binding)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (c *artifactRegistryCache) RemoveArtifactRegistryPolicyBinding(ctx context.Context, id *service.ContainerRepositoryIdentifier, binding *service.Binding) error {
	const op errs.Op = "artifactRegistryCache.RemoveArtifactRegistryPolicyBinding"

	err := c.api.RemoveArtifactRegistryPolicyBinding(ctx, id, binding)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (c *artifactRegistryCache) GetContainerImage(ctx context.Context, id *service.ContainerImageIdentifier) (*service.ContainerImage, error) {
	const op errs.Op = "artifactRegistryCache.GetContainerImage"

	containerImage, err := c.api.GetContainerImage(ctx, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return containerImage, nil
}

func (c *artifactRegistryCache) GetContainerImageVersion(ctx context.Context, id *service.ContainerRepositoryIdentifier, image string) (*service.ContainerImage, error) {
	const op errs.Op = "artifactRegistryCache.GetContainerImageVersion"

	containerImageVersion, err := c.api.GetContainerImageVersion(ctx, id, image)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return containerImageVersion, nil
}

func NewArtifactRegistryCache(api service.ArtifactRegistryAPI, cache cache.Cacher) *artifactRegistryCache {
	return &artifactRegistryCache{
		api:   api,
		cache: cache,
	}
}
