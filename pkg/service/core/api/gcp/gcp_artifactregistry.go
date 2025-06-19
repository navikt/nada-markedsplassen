package gcp

import (
	"context"
	"strings"

	"github.com/navikt/nada-backend/pkg/artifactregistry"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

var _ service.ArtifactRegistryAPI = &artifactRegistryAPI{}

type artifactRegistryAPI struct {
	ops artifactregistry.Operations
	log zerolog.Logger
}

func (a *artifactRegistryAPI) AddArtifactRegistryPolicyBinding(ctx context.Context, id *service.ContainerRepositoryIdentifier, binding *service.Binding) error {
	const op errs.Op = "gcp.AddArtifactRegistryPolicyBinding"

	err := a.ops.AddArtifactRegistryPolicyBinding(ctx, &artifactregistry.ContainerRepositoryIdentifier{
		Project:    id.Project,
		Location:   id.Location,
		Repository: id.Repository,
	}, &artifactregistry.Binding{
		Role:    binding.Role,
		Members: binding.Members,
	})
	if err != nil {
		return errs.E(errs.IO, service.CodeGCPArtifactRegistry, op, err)
	}

	return nil
}

func (a *artifactRegistryAPI) RemoveArtifactRegistryPolicyBinding(ctx context.Context, id *service.ContainerRepositoryIdentifier, binding *service.Binding) error {
	const op errs.Op = "gcp.RemoveArtifactRegistryPolicyBinding"

	err := a.ops.RemoveArtifactRegistryPolicyBinding(ctx, &artifactregistry.ContainerRepositoryIdentifier{
		Project:    id.Project,
		Location:   id.Location,
		Repository: id.Repository,
	}, &artifactregistry.Binding{
		Role:    binding.Role,
		Members: binding.Members,
	})
	if err != nil {
		return errs.E(errs.IO, service.CodeGCPArtifactRegistry, op, err)
	}

	return nil
}

func (a *artifactRegistryAPI) ListContainerImagesWithTag(ctx context.Context, id *service.ContainerRepositoryIdentifier, tag string) ([]*service.ContainerImage, error) {
	const op errs.Op = "gcp.ListContainerImagesWithTag"

	raw, err := a.ops.ListContainerImagesWithTag(ctx, &artifactregistry.ContainerRepositoryIdentifier{
		Project:    id.Project,
		Location:   id.Location,
		Repository: id.Repository,
	}, tag)
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeGCPArtifactRegistry, op, err)
	}

	var images []*service.ContainerImage
	for _, image := range raw {
		i, err := a.GetContainerImage(ctx, &service.ContainerImageIdentifier{
			RepositoryID: id,
			Name:         image.ImageVersion.Name,
			URI:          image.ImageVersion.URI,
			Tag:          image.ImageVersion.Tag,
			Digest:       image.ImageVersion.Digest,
		})
		if err != nil {
			return nil, errs.E(errs.IO, service.CodeGCPArtifactRegistry, op, err)
		}
		images = append(images, i)
	}

	return images, nil
}

func (a *artifactRegistryAPI) GetContainerImageVersion(ctx context.Context, id *service.ContainerRepositoryIdentifier, image, tag string) (*service.ContainerImage, error) {
	const op errs.Op = "gcp.GetContainerImageTags"

	parts := strings.Split(image, ":")
	image = parts[0]
	tag = parts[1]

	imageVersion, err := a.ops.GetContainerImageVersion(ctx, &artifactregistry.ContainerRepositoryIdentifier{
		Project:    id.Project,
		Location:   id.Location,
		Repository: id.Repository,
	}, image, tag)
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeGCPArtifactRegistry, op, err)
	}

	containerImage, err := a.GetContainerImage(ctx, &service.ContainerImageIdentifier{
		RepositoryID: id,
		Name:         imageVersion.Name,
		URI:          imageVersion.URI,
		Tag:          imageVersion.Tag,
		Digest:       imageVersion.Digest,
	})
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeGCPArtifactRegistry, op, err)
	}

	return containerImage, nil
}

func (a *artifactRegistryAPI) GetContainerImage(ctx context.Context, id *service.ContainerImageIdentifier) (*service.ContainerImage, error) {
	const op errs.Op = "gcp.GetContainerImage"

	image := &artifactregistry.ContainerImage{
		ID: artifactregistry.ContainerRepositoryIdentifier(*id.RepositoryID),
		ImageVersion: artifactregistry.ContainerImageVersion{
			Name:   id.Name,
			URI:    id.URI,
			Tag:    id.Tag,
			Digest: id.Digest,
		},
	}

	labels := map[string]string{}
	docs := ""
	var containerConfig *service.ContainerConfig

	attachments, err := a.ops.ListContainerImageAttachments(ctx, image)
	if err != nil {
		a.log.Error().Err(err).Msgf("failed to get attachments for image %s", image.ImageVersion.URI)
	}

	for _, at := range attachments {
		file, err := a.ops.DownloadAttachmentFile(ctx, at)
		if err != nil {
			a.log.Error().Err(err).Msgf("failed to get attachment %s for image %s", at, image.ImageVersion.URI)
			continue
		}

		if at.Type == service.ArtifactTypeKnastAnnotations {
			var annotations service.Annotations

			err := yaml.Unmarshal(file.Data, &annotations)
			if err != nil {
				a.log.Error().Err(err).Msgf("failed to unmarshal annotations for image %s", image.ImageVersion.URI)
				continue
			}

			labels["org.opencontainers.image.source"] = annotations.Source
			labels["org.opencontainers.image.title"] = annotations.Title
			labels["org.opencontainers.image.description"] = annotations.Description
		}

		if at.Type == service.ArtifactTypeKnastIndex {
			docs = string(file.Data)
		}

		if at.Type == service.ArtifactTypeKnastContainerConfig {
			var config service.ContainerConfig

			err := yaml.Unmarshal(file.Data, &config)
			if err != nil {
				a.log.Error().Err(err).Msgf("failed to unmarshal container config for image %s", image.ImageVersion.URI)
				continue
			}

			containerConfig = &config
		}
	}

	return &service.ContainerImage{
		Name: image.ImageVersion.Name,
		URI:  image.ImageVersion.URI,
		Manifest: &service.Manifest{
			Labels: labels,
		},
		Documentation:   docs,
		ContainerConfig: containerConfig,
	}, nil
}

func NewArtifactRegistryAPI(ops artifactregistry.Operations, log zerolog.Logger) service.ArtifactRegistryAPI {
	return &artifactRegistryAPI{
		ops: ops,
		log: log,
	}
}
