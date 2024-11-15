package artifactregistry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	artv1 "cloud.google.com/go/artifactregistry/apiv1"
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"cloud.google.com/go/iam/apiv1/iampb"

	"github.com/containers/image/v5/docker"
	dockeref "github.com/containers/image/v5/docker/reference"
	imgtypes "github.com/containers/image/v5/types"
	"golang.org/x/exp/maps"
	"golang.org/x/oauth2"
	auth "golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var ErrNotFound = errors.New("not found")

type Operations interface {
	ListContainerImagesWithTag(ctx context.Context, id *ContainerRepositoryIdentifier, tag string) ([]*ContainerImage, error)
	AddArtifactRegistryPolicyBinding(ctx context.Context, id *ContainerRepositoryIdentifier, binding *Binding) error
	RemoveArtifactRegistryPolicyBinding(ctx context.Context, id *ContainerRepositoryIdentifier, binding *Binding) error
	GetContainerImageManifest(ctx context.Context, imageURI string) (*Manifest, error)
	ListContainerImageAttachments(ctx context.Context, img *ContainerImage) ([]*Attachment, error)
	DownloadAttachmentFile(ctx context.Context, attachment *Attachment) (*File, error)
}

type Attachment struct {
	Name string

	// Required. The target the attachment is for, can be a Version, Package or
	// Repository. E.g.
	// `projects/p1/locations/us-central1/repositories/repo1/packages/p1/versions/v1`.
	Target string

	// Type of attachment.
	// E.g. `application/vnd.spdx+json`
	Type string

	// The namespace this attachment belongs to.
	// E.g. If an attachment is created by artifact analysis, namespace is set
	// to `artifactanalysis.googleapis.com`.
	AttachmentNamespace string

	// Optional. User annotations. These attributes can only be set and used by
	// the user, and not by Artifact Registry. See
	// https://google.aip.dev/128#annotations for more details such as format and
	// size limitations.
	Annotations map[string]string

	// Required. The files that belong to this attachment.
	// If the file ID part contains slashes, they are escaped. E.g.
	// `projects/p1/locations/us-central1/repositories/repo1/files/sha:<sha-of-file>`.
	Files []string

	// Output only. The name of the OCI version that this attachment created. Only
	// populated for Docker attachments. E.g.
	// `projects/p1/locations/us-central1/repositories/repo1/packages/p1/versions/v1`.
	OciVersionName string
}

type Manifest struct {
	Labels map[string]string
}

type Binding struct {
	Role    string
	Members []string
}

type ContainerImage struct {
	ID ContainerRepositoryIdentifier

	Name   string
	URI    string
	Tag    string
	Digest string
}

func (c *ContainerImage) Target() string {
	return fmt.Sprintf("projects/%s/locations/%s/repositories/%s/packages/%s/versions/%s", c.ID.Project, c.ID.Location, c.ID.Repository, c.Name, c.Digest)
}

type ContainerRepositoryIdentifier struct {
	Project    string
	Location   string
	Repository string
}

func (c *ContainerRepositoryIdentifier) Parent() string {
	return fmt.Sprintf("projects/%s/locations/%s/repositories/%s", c.Project, c.Location, c.Repository)
}

type File struct {
	Name string
	Data []byte
}

type Client struct {
	apiEndpoint string
	disableAuth bool
}

func (c *Client) DownloadAttachmentFile(ctx context.Context, attachment *Attachment) (*File, error) {
	var err error

	token := ""
	if !c.disableAuth {
		token, err = c.getToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting token: %w", err)
		}
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("https://artifactregistry.googleapis.com/v1/%s:download?alt=media", attachment.Files[0])

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request for attachment %s: %w", attachment.Name, err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("downloading attachment %s: %s", attachment.Name, resp.Status)
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading attachment %s: %w", attachment.Name, err)
	}

	return &File{
		Name: attachment.Files[0],
		Data: data,
	}, nil
}

func (c *Client) ListContainerImageAttachments(ctx context.Context, img *ContainerImage) ([]*Attachment, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	it := client.ListAttachments(ctx, &artifactregistrypb.ListAttachmentsRequest{
		Parent: img.ID.Parent(),
		Filter: fmt.Sprintf(`target="%s"`, img.Target()),
	})

	var attachments []*Attachment

	for {
		a, err := it.Next()

		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("iterating over attachments %s: %w", img.Name, err)
		}

		attachments = append(attachments, &Attachment{
			Name:                a.Name,
			Target:              a.Target,
			Type:                a.Type,
			AttachmentNamespace: a.AttachmentNamespace,
			Annotations:         a.Annotations,
			Files:               a.Files,
			OciVersionName:      a.OciVersionName,
		})
	}

	return attachments, nil
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

		uri, digest, didSplit := strings.Cut(image.Uri, "@")
		// Should never happen
		if !didSplit {
			return nil, fmt.Errorf("splitting image URI: %w", err)
		}

		if !strings.HasPrefix(digest, "sha256:") {
			return nil, fmt.Errorf("invalid digest %s", digest)
		}

		name, _, _ := strings.Cut(image.Name, "@")
		name = path.Base(name)

		images = append(images, &ContainerImage{
			ID:     *id,
			Name:   name,
			URI:    fmt.Sprintf("%s:%s", uri, tag),
			Digest: digest,
			Tag:    tag,
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

func New(apiEndpoint string, disableAuth bool) *Client {
	return &Client{
		apiEndpoint: apiEndpoint,
		disableAuth: disableAuth,
	}
}
