package artifactregistry

import (
	artv1 "cloud.google.com/go/artifactregistry/apiv1"
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"strings"
)

type Operations interface {
	ListContainerImagesWithTag(ctx context.Context, id *ContainerRepositoryIdentifier, tag string) ([]*ContainerImage, error)
}

type ContainerImage struct {
	Name string
	URI  string
}

type ContainerRepositoryIdentifier struct {
	Project    string
	Location   string
	Repository string
}

func (c *ContainerRepositoryIdentifier) Parent() string {
	return fmt.Sprintf("projects/%s/locations/%s/repositories/%s", c.Project, c.Location, c.Repository)
}

func (c *Client) ListContainerImagesWithTag(ctx context.Context, id *ContainerRepositoryIdentifier, tag string) ([]*ContainerImage, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	it := client.ListDockerImages(ctx, &artifactregistrypb.ListDockerImagesRequest{
		Parent: id.Parent(),
	})

	var images []*ContainerImage

	for {
		image, err := it.Next()

		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("iterating over docker images: %w", err)
		}

		hasTag := false
		for _, t := range image.Tags {
			if t == tag {
				hasTag = true
				break
			}
		}

		if !hasTag {
			continue
		}

		uri, _, didSplit := strings.Cut(image.Uri, "@")
		// Should never happen
		if !didSplit {
			return nil, fmt.Errorf("splitting image URI: %w", err)
		}

		images = append(images, &ContainerImage{
			Name: image.Name,
			URI:  fmt.Sprintf("%s:%s", uri, tag),
		})
	}

	return images, nil
}

func (c *Client) newClient(ctx context.Context) (*artv1.Client, error) {
	var opts []option.ClientOption

	if c.disableAuth {
		opts = append(opts, option.WithoutAuthentication())
	}

	if c.apiEndpoint != "" {
		opts = append(opts, option.WithEndpoint(c.apiEndpoint))
	}

	client, err := artv1.NewRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating artifact registry client: %w", err)
	}

	return client, nil
}

type Client struct {
	apiEndpoint string
	disableAuth bool
}

func New(apiEndpoint string, disableAuth bool) *Client {
	return &Client{
		apiEndpoint: apiEndpoint,
		disableAuth: disableAuth,
	}
}
