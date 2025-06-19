package service

import (
	"context"
)

const (
	ArtifactRegistryReaderRole = "roles/artifactregistry.reader"

	ArtifactTypeKnastIndex           = "application/vnd.knast.docs.index"
	ArtifactTypeKnastAnnotations     = "application/vnd.knast.annotations"
	ArtifactTypeKnastContainerConfig = "application/vnd.knast.container.config"
)

type ArtifactRegistryAPI interface {
	ListContainerImagesWithTag(ctx context.Context, id *ContainerRepositoryIdentifier, tag string) ([]*ContainerImage, error)
	AddArtifactRegistryPolicyBinding(ctx context.Context, id *ContainerRepositoryIdentifier, binding *Binding) error
	RemoveArtifactRegistryPolicyBinding(ctx context.Context, id *ContainerRepositoryIdentifier, binding *Binding) error
	GetContainerImage(ctx context.Context, id *ContainerImageIdentifier) (*ContainerImage, error)
	GetContainerImageVersion(ctx context.Context, id *ContainerRepositoryIdentifier, image, tag string) (*ContainerImage, error)
}

type ContainerRepositoryIdentifier struct {
	Project    string
	Location   string
	Repository string
}

type ContainerImageIdentifier struct {
	RepositoryID *ContainerRepositoryIdentifier

	Name   string
	Tag    string
	URI    string
	Digest string
}

type ContainerImage struct {
	Name            string
	URI             string
	Manifest        *Manifest
	Documentation   string
	ContainerConfig *ContainerConfig
}

type Manifest struct {
	Labels map[string]string
}

type Annotations struct {
	Source      string `yaml:"source"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

type ReadinessCheck struct {
	Path string `yaml:"path"`
	Port int    `yaml:"port"`
}

type ContainerConfig struct {
	ExtraReadinessChecks []*ReadinessCheck `yaml:"extra_readiness_checks"`
}
