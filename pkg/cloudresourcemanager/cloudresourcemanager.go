package cloudresourcemanager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"golang.org/x/oauth2/google"

	"github.com/rs/zerolog"
	"golang.org/x/exp/maps"
	"google.golang.org/api/cloudresourcemanager/v1"
	crmv3 "google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const (
	DeletedPrefix = "deleted:"
)

var ErrNotFound = errors.New("not found")

type Operations interface {
	// AddProjectIAMPolicyBinding adds a binding to a project IAM policy.
	AddProjectIAMPolicyBinding(ctx context.Context, project string, binding *Binding) error

	// RemoveProjectIAMPolicyBindingMember removes a member from all project IAM policy bindings.
	RemoveProjectIAMPolicyBindingMember(ctx context.Context, project string, member string) error

	// RemoveProjectIAMPolicyBindingMemberForRole removes a member from a specific role in all project IAM policy bindings.
	RemoveProjectIAMPolicyBindingMemberForRole(ctx context.Context, project, role, member string) error

	// ListProjectIAMPolicyBindings returns all project IAM policy bindings for a specific member.
	ListProjectIAMPolicyBindings(ctx context.Context, project, member string) ([]*Binding, error)

	// UpdateProjectIAMPolicyBindingsMembers allows updating the members of all project IAM policy bindings.
	UpdateProjectIAMPolicyBindingsMembers(ctx context.Context, project string, fn UpdateProjectIAMPolicyBindingsMembersFn) error

	// CreateZonalTagBinding creates a new tag binding
	CreateZonalTagBinding(ctx context.Context, project, parentResource, tagNamespacedName string) error
}

type UpdateProjectIAMPolicyBindingsMembersFn func(role string, members []string) []string

type Binding struct {
	Role    string
	Members []string
}

var _ Operations = &Client{}

type Client struct {
	endpoint             string
	disableAuth          bool
	tagBindingHTTPClient *http.Client
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	tokenSource, err := google.DefaultTokenSource(ctx)
	if err != nil {
		return "", fmt.Errorf("getting default token source: %w", err)
	}

	token, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("getting token: %w", err)
	}

	return token.AccessToken, nil
}

// CreateZonalTagBinding creates a new tag binding
// Note: we cannot use the provided client to create the tag binding, as the client does not support adding
// an instance binding yet
func (c *Client) CreateZonalTagBinding(ctx context.Context, zone, parentResource, tagNamespacedName string) error {
	var err error

	token := ""
	if !c.disableAuth {
		token, err = c.getToken(ctx)
		if err != nil {
			return fmt.Errorf("getting token: %w", err)
		}
	}

	body := crmv3.TagBinding{
		Parent:                 parentResource,
		TagValueNamespacedName: tagNamespacedName,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshalling request body: %w", err)
	}

	url := fmt.Sprintf("https://%s-cloudresourcemanager.googleapis.com/v3/tagBindings", zone)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	_, err = c.tagBindingHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}

	return nil
}

func (c *Client) RemoveProjectIAMPolicyBindingMemberForRole(ctx context.Context, project, role, member string) error {
	service, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	policy, err := service.Projects.GetIamPolicy(project, &cloudresourcemanager.GetIamPolicyRequest{}).Do()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return fmt.Errorf("project %s: %w", project, ErrNotFound)
		}

		return fmt.Errorf("getting project %s policy: %w", project, err)
	}

	var bindings []*cloudresourcemanager.Binding

	for _, binding := range policy.Bindings {
		// Skip the binding if it's not the one we want to remove the member from.
		if binding.Role != role {
			bindings = append(bindings, binding)

			continue
		}

		var members []string

		// Skip the member if it's the one we want to remove.
		for _, m := range binding.Members {
			if m != member {
				members = append(members, m)
			}
		}

		if len(members) > 0 {
			bindings = append(bindings, &cloudresourcemanager.Binding{
				Role:    binding.Role,
				Members: members,
			})
		}
	}

	policy.Bindings = bindings

	_, err = service.Projects.SetIamPolicy(project, &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}).Do()
	if err != nil {
		return fmt.Errorf("setting project %s policy: %w", project, err)
	}

	return nil
}

func (c *Client) UpdateProjectIAMPolicyBindingsMembers(ctx context.Context, project string, fn UpdateProjectIAMPolicyBindingsMembersFn) error {
	service, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	policy, err := service.Projects.GetIamPolicy(project, &cloudresourcemanager.GetIamPolicyRequest{}).Do()
	if err != nil {
		return fmt.Errorf("getting project %s policy: %w", project, err)
	}

	for _, binding := range policy.Bindings {
		binding.Members = fn(binding.Role, binding.Members)
	}

	_, err = service.Projects.SetIamPolicy(project, &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}).Do()
	if err != nil {
		return fmt.Errorf("setting project %s policy: %w", project, err)
	}

	return nil
}

func (c *Client) RemoveProjectIAMPolicyBindingMember(ctx context.Context, project string, member string) error {
	service, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	policy, err := service.Projects.GetIamPolicy(project, &cloudresourcemanager.GetIamPolicyRequest{}).Do()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return fmt.Errorf("project %s: %w", project, ErrNotFound)
		}

		return fmt.Errorf("getting project %s policy: %w", project, err)
	}

	var bindings []*cloudresourcemanager.Binding

	for _, binding := range policy.Bindings {
		var members []string

		for _, m := range binding.Members {
			if m != member {
				members = append(members, m)
			}
		}

		if len(members) > 0 {
			bindings = append(bindings, &cloudresourcemanager.Binding{
				Role:    binding.Role,
				Members: members,
			})
		}
	}

	policy.Bindings = bindings

	_, err = service.Projects.SetIamPolicy(project, &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}).Do()
	if err != nil {
		return fmt.Errorf("setting project %s policy: %w", project, err)
	}

	return nil
}

func (c *Client) ListProjectIAMPolicyBindings(ctx context.Context, project, member string) ([]*Binding, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	policy, err := client.Projects.GetIamPolicy(project, &cloudresourcemanager.GetIamPolicyRequest{}).Do()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return nil, fmt.Errorf("project %s: %w", project, ErrNotFound)
		}

		return nil, fmt.Errorf("getting project %s policy: %w", project, err)
	}

	var bindings []*Binding

	for _, binding := range policy.Bindings {
		for _, m := range binding.Members {
			if m == member {
				bindings = append(bindings, &Binding{
					Role:    binding.Role,
					Members: binding.Members,
				})

				break
			}
		}
	}

	return bindings, nil
}

func (c *Client) AddProjectIAMPolicyBinding(ctx context.Context, project string, binding *Binding) error {
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	policy, err := client.Projects.GetIamPolicy(project, &cloudresourcemanager.GetIamPolicyRequest{}).Do()
	if err != nil {
		return fmt.Errorf("getting project %s policy: %w", project, err)
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
		policy.Bindings = append(policy.Bindings, &cloudresourcemanager.Binding{
			Role:    binding.Role,
			Members: binding.Members,
		})
	}

	_, err = client.Projects.SetIamPolicy(project, &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}).Do()
	if err != nil {
		return fmt.Errorf("setting project %s policy: %w", project, err)
	}

	return nil
}

func (c *Client) newClient(ctx context.Context) (*cloudresourcemanager.Service, error) {
	var opts []option.ClientOption

	if c.disableAuth {
		opts = append(opts, option.WithoutAuthentication())
	}

	if len(c.endpoint) > 0 {
		opts = append(opts, option.WithEndpoint(c.endpoint))
	}

	service, err := cloudresourcemanager.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating cloudresourcemanager service: %w", err)
	}

	return service, nil
}

func NewClient(endpoint string, disableAuth bool, tagBindingHTTPClient *http.Client) *Client {
	return &Client{
		endpoint:             endpoint,
		disableAuth:          disableAuth,
		tagBindingHTTPClient: tagBindingHTTPClient,
	}
}

func RemoveDeletedMembersWithRole(roles []string, log zerolog.Logger) UpdateProjectIAMPolicyBindingsMembersFn {
	return func(role string, members []string) []string {
		if !slices.Contains(roles, role) {
			log.Info().Str("role", role).Msg("Skipping role")

			return members
		}

		var keep, remove []string

		for _, member := range members {
			if strings.HasPrefix(member, DeletedPrefix) {
				remove = append(remove, member)

				continue
			}

			keep = append(keep, member)
		}

		log.Info().Str("role", role).Fields(map[string]interface{}{
			"removed_members": remove,
			"kept_members":    keep,
			"removed_count":   len(members) - len(keep),
		}).Msg("Removed deleted members")

		return keep
	}
}
