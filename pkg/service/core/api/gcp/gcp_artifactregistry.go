package gcp

import (
	"context"

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
		labels := map[string]string{}
		docs := ""

		// Fetch the manifest to get the labels, for now, we ignore errors
		manifest, err := a.ops.GetContainerImageManifest(ctx, image.URI)
		if err != nil {
			a.log.Error().Err(err).Msgf("failed to get manifest for image %s", image.URI)
		}

		if err == nil {
			labels = manifest.Labels
		}

		attachments, err := a.ops.ListContainerImageAttachments(ctx, image)
		if err != nil {
			a.log.Error().Err(err).Msgf("failed to get attachments for image %s", image.URI)
		}

		for _, at := range attachments {
			file, err := a.ops.DownloadAttachmentFile(ctx, at)
			if err != nil {
				a.log.Error().Err(err).Msgf("failed to get attachment %s for image %s", at, image.URI)
				continue
			}

			if at.Type == service.ArtifactTypeKnastAnnotations {
				var annotations service.Annotations

				err := yaml.Unmarshal(file.Data, &annotations)
				if err != nil {
					a.log.Error().Err(err).Msgf("failed to unmarshal annotations for image %s", image.URI)
					continue
				}

				labels["org.opencontainers.image.source"] = annotations.Source
				labels["org.opencontainers.image.title"] = annotations.Title
				labels["org.opencontainers.image.description"] = annotations.Description
			}

			if at.Type == service.ArtifactTypeKnastIndex {
				docs = string(file.Data)
			}
		}

		images = append(images, &service.ContainerImage{
			Name: image.Name,
			URI:  image.URI,
			Manifest: &service.Manifest{
				Labels: labels,
			},
			Documentation: docs,
		})
	}

	return images, nil
}

func NewArtifactRegistryAPI(ops artifactregistry.Operations, log zerolog.Logger) service.ArtifactRegistryAPI {
	return &artifactRegistryAPI{
		ops: ops,
		log: log,
	}
}
