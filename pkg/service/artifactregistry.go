package service

import "context"

const (
	ArtifactRegistryReaderRole = "roles/artifactregistry.reader"
)

type ArtifactRegistryAPI interface {
	ListContainerImagesWithTag(ctx context.Context, id *ContainerRepositoryIdentifier, tag string) ([]*ContainerImage, error)
	AddArtifactRegistryPolicyBinding(ctx context.Context, id *ContainerRepositoryIdentifier, binding *Binding) error
	RemoveArtifactRegistryPolicyBinding(ctx context.Context, id *ContainerRepositoryIdentifier, binding *Binding) error
}

type ContainerRepositoryIdentifier struct {
	Project    string
	Location   string
	Repository string
}

type ContainerImage struct {
	Name     string
	URI      string
	Manifest *Manifest
}

type Manifest struct {
	Labels map[string]string
}
