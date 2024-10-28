package service

import "context"

type ArtifactRegistryAPI interface {
	ListContainerImagesWithTag(ctx context.Context, id *ContainerRepositoryIdentifier, tag string) ([]*ContainerImage, error)
}

type ContainerRepositoryIdentifier struct {
	Project    string
	Location   string
	Repository string
}

type ContainerImage struct {
	Name string
	URI  string
}
