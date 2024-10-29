package artifactregistry

import (
	artv1 "cloud.google.com/go/artifactregistry/apiv1"
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"cloud.google.com/go/iam/apiv1/iampb"
	"context"
	"errors"
	"fmt"
	"github.com/containers/image/v5/docker"
	dockeref "github.com/containers/image/v5/docker/reference"
	imgtypes "github.com/containers/image/v5/types"
	"golang.org/x/exp/maps"
	"golang.org/x/oauth2"
	auth "golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"net/http"
	"strings"
)

var (
	ErrNotFound = errors.New("not found")
)

type Operations interface {
	ListContainerImagesWithTag(ctx context.Context, id *ContainerRepositoryIdentifier, tag string) ([]*ContainerImage, error)
	AddArtifactRegistryPolicyBinding(ctx context.Context, id *ContainerRepositoryIdentifier, binding *Binding) error
	RemoveArtifactRegistryPolicyBinding(ctx context.Context, id *ContainerRepositoryIdentifier, binding *Binding) error
	GetContainerImageManifest(ctx context.Context, imageURI string) (*Manifest, error)
}

type Manifest struct {
	Labels map[string]string
}

type Binding struct {
	Role    string
	Members []string
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

func (c *Client) GetContainerImageManifest(ctx context.Context, imageURI string) (*Manifest, error) {
	var err error

	token := ""
	if !c.disableAuth {
		token, err = c.getToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting token: %w", err)
		}
	}

	namedRef, err := dockeref.ParseDockerRef(imageURI)
	if err != nil {
		return nil, fmt.Errorf("parsing image URI %s: %w", imageURI, err)
	}

	ref, err := docker.NewReference(namedRef)
	if err != nil {
		return nil, fmt.Errorf("creating reference %s: %w", imageURI, err)
	}

	img, err := ref.NewImage(ctx, &imgtypes.SystemContext{
		DockerBearerRegistryToken: token,
	})
	if err != nil {
		return nil, fmt.Errorf("creating image context %s: %w", imageURI, err)
	}
	defer img.Close()

	ins, err := img.Inspect(ctx)
	if err != nil {
		return nil, fmt.Errorf("inspecting image %s: %w", imageURI, err)
	}

	return &Manifest{
		Labels: ins.Labels,
	}, nil
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	var token *oauth2.Token

	scopes := []string{
		"https://www.googleapis.com/auth/cloud-platform",
	}

	credentials, err := auth.FindDefaultCredentials(ctx, scopes...)
	if err != nil {
		return "", fmt.Errorf("finding default credentials: %w", err)
	}

	token, err = credentials.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("getting token: %w", err)
	}

	return token.AccessToken, nil
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
			return nil, fmt.Errorf("iterating over images %s: %w", id.Parent(), err)
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

func (c *Client) RemoveArtifactRegistryPolicyBinding(ctx context.Context, id *ContainerRepositoryIdentifier, remove *Binding) error {
	service, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	policy, err := service.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: id.Parent(),
		Options:  nil,
	})
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return fmt.Errorf("policy %s: %w", id.Parent(), ErrNotFound)
		}

		return fmt.Errorf("getting policy %s: %w", id.Parent(), err)
	}

	var bindings []*iampb.Binding

	for _, existing := range policy.Bindings {
		if existing.Role != remove.Role {
			bindings = append(bindings, existing)

			continue
		}

		var existingMembers []string

		for _, existingMember := range existing.Members {
			notFound := true
			for _, removeMember := range remove.Members {
				if existingMember == removeMember {
					notFound = false
					break
				}
			}

			if notFound {
				existingMembers = append(existingMembers, existingMember)
			}
		}

		if len(existingMembers) > 0 {
			bindings = append(bindings, &iampb.Binding{
				Role:    existing.Role,
				Members: existingMembers,
			})
		}
	}

	policy.Bindings = bindings

	_, err = service.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: id.Parent(),
		Policy:   policy,
	})
	if err != nil {
		return fmt.Errorf("setting policy %s: %w", id.Parent(), err)
	}

	return nil
}

func (c *Client) AddArtifactRegistryPolicyBinding(ctx context.Context, id *ContainerRepositoryIdentifier, binding *Binding) error {
	service, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	policy, err := service.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: id.Parent(),
	})
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return fmt.Errorf("policy %s: %w", id.Parent(), ErrNotFound)
		}

		return fmt.Errorf("getting policy for %s: %w", id.Parent(), err)
	}

	uniqueMembers := make(map[string]struct{})
	for _, member := range binding.Members {
		uniqueMembers[member] = struct{}{}
	}

	found := false

	for _, b := range policy.Bindings {
		if b.Role == binding.Role {
			for _, member := range b.Members {
				uniqueMembers[member] = struct{}{}
			}

			b.Members = maps.Keys(uniqueMembers)
			found = true
			break
		}
	}

	if !found {
		policy.Bindings = append(policy.Bindings, &iampb.Binding{
			Role:    binding.Role,
			Members: binding.Members,
		})
	}

	_, err = service.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: id.Parent(),
		Policy:   policy,
	})
	if err != nil {
		return fmt.Errorf("setting policy %s: %w", id.Parent(), err)
	}

	return nil
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
