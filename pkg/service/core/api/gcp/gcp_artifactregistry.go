package gcp

import (
	"context"
	"github.com/navikt/nada-backend/pkg/artifactregistry"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.ArtifactRegistryAPI = &artifactRegistryAPI{}

type artifactRegistryAPI struct {
	ops artifactregistry.Operations
}

func (a *artifactRegistryAPI) ListContainerImagesWithTag(ctx context.Context, id *service.ContainerRepositoryIdentifier, tag string) ([]*service.ContainerImage, error) {
	const op errs.Op = "gcp.ListContainerImagesWithTag"

	raw, err := a.ops.ListContainerImagesWithTag(ctx, &artifactregistry.ContainerRepositoryIdentifier{
		Project:    id.Project,
		Location:   id.Location,
		Repository: id.Repository,
	}, tag)
	if err != nil {
		return nil, errs.E(op, err)
	}

	var images []*service.ContainerImage
	for _, image := range raw {
		images = append(images, &service.ContainerImage{
			Name: image.Name,
			URI:  image.URI,
		})
	}

	return images, nil
}

func NewArtifactRegistryAPI(ops artifactregistry.Operations) service.ArtifactRegistryAPI {
	return &artifactRegistryAPI{
		ops: ops,
	}
}
