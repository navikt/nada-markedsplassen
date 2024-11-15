package service

import "context"

const (
	ArtifactRegistryReaderRole = "roles/artifactregistry.reader"

	ArtifactTypeKnastIndex       = "application/vnd.knast.docs.index"
	ArtifactTypeKnastAnnotations = "application/vnd.knast.annotations"
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
	Name          string
	URI           string
	Manifest      *Manifest
	Documentation string
}

type Manifest struct {
	Labels map[string]string
}

type Annotations struct {
	Source      string `yaml:"source"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}
