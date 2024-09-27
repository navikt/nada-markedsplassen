package securewebproxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/networksecurity/v1"
	"google.golang.org/api/option"
)

var (
	ErrNotExist = errors.New("not exist")
	ErrExist    = errors.New("exist")
)

var _ Operations = &Client{}

type Operations interface {
	GetURLList(ctx context.Context, id *URLListIdentifier) ([]string, error)
	CreateURLList(ctx context.Context, opts *URLListCreateOpts) error
	DeleteURLList(ctx context.Context, id *URLListIdentifier) error

	GetSecurityPolicyRule(ctx context.Context, id *PolicyRuleIdentifier) (*GatewaySecurityPolicyRule, error)
	CreateSecurityPolicyRule(ctx context.Context, opts *PolicyRuleCreateOpts) error
	DeleteSecurityPolicyRule(ctx context.Context, id *PolicyRuleIdentifier) error
}

type URLListIdentifier struct {
	// Project is the gcp project id
	Project string

	// Location is the gcp region
	Location string

	// Slug is the name of the url list
	Slug string
}

func (i *URLListIdentifier) Parent() string {
	return fmt.Sprintf("projects/%s/locations/%s", i.Project, i.Location)
}

func (i *URLListIdentifier) FullyQualifiedName() string {
	return fmt.Sprintf("projects/%s/locations/%s/urlLists/%s", i.Project, i.Location, i.Slug)
}

type URLListCreateOpts struct {
	ID          *URLListIdentifier
	Description string
	URLS        []string
}

type PolicyRuleIdentifier struct {
	// Project is the gcp project id
	Project string

	// Location is the gcp region
	Location string

	// Policy is the name of the policy the rule is part of
	Policy string

	// Slug is the name of the policy rule
	Slug string
}

func (p *PolicyRuleIdentifier) Parent() string {
	return fmt.Sprintf("projects/%s/locations/%s/gatewaySecurityPolicies/%s", p.Project, p.Location, p.Policy)
}

func (p *PolicyRuleIdentifier) FullyQualifiedName() string {
	return fmt.Sprintf("projects/%s/locations/%s/gatewaySecurityPolicies/%s/rules/%s", p.Project, p.Location, p.Policy, p.Slug)
}

type GatewaySecurityPolicyRule struct {
	// ApplicationMatcher: Optional. CEL expression for matching on L7/application
	// level criteria.
	ApplicationMatcher string
	// BasicProfile: Required. Profile which tells what the primitive action should
	// be.
	//
	// Possible values:
	//   "BASIC_PROFILE_UNSPECIFIED" - If there is not a mentioned action for the
	// target.
	//   "ALLOW" - Allow the matched traffic.
	//   "DENY" - Deny the matched traffic.
	BasicProfile string
	// CreateTime: Output only. Time when the rule was created.
	CreateTime string
	// Description: Optional. Free-text description of the resource.
	Description string
	// Enabled: Required. Whether the rule is enforced.
	Enabled bool
	// Name: Required. Immutable. Name of the resource. ame is the full resource
	// name so
	// projects/{project}/locations/{location}/gatewaySecurityPolicies/{gateway_secu
	// rity_policy}/rules/{rule} rule should match the pattern: (^a-z
	// ([a-z0-9-]{0,61}[a-z0-9])?$).
	Name string
	// Priority: Required. Priority of the rule. Lower number corresponds to higher
	// precedence.
	Priority int64
	// SessionMatcher: Required. CEL expression for matching on session criteria.
	SessionMatcher string
	// TlsInspectionEnabled: Optional. Flag to enable TLS inspection of traffic
	// matching on , can only be true if the parent GatewaySecurityPolicy
	// references a TLSInspectionConfig.
	TlsInspectionEnabled bool
	// UpdateTime: Output only. Time when the rule was updated.
	UpdateTime string
}

type PolicyRuleCreateOpts struct {
	ID   *PolicyRuleIdentifier
	Rule *GatewaySecurityPolicyRule
}

type Client struct {
	apiEndpoint string
	disableAuth bool
}

func (c *Client) GetURLList(ctx context.Context, id *URLListIdentifier) ([]string, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	list, err := client.Projects.Locations.UrlLists.Get(id.FullyQualifiedName()).Do()
	if err != nil {
		var gapierr *googleapi.Error
		if errors.As(err, &gapierr) && gapierr.Code == http.StatusNotFound {
			return nil, ErrNotExist
		}

		return nil, err
	}

	return list.Values, nil
}

func (c *Client) CreateURLList(ctx context.Context, opts *URLListCreateOpts) error {
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.Projects.Locations.UrlLists.Create(opts.ID.Parent(), &networksecurity.UrlList{
		Name:        opts.ID.FullyQualifiedName(),
		Description: opts.Description,
		Values:      opts.URLS,
	}).Do()
	if err != nil {
		var gapierr *googleapi.Error
		if errors.As(err, &gapierr) && gapierr.Code == http.StatusConflict {
			return ErrExist
		}

		return err
	}

	return nil
}

func (c *Client) DeleteURLList(ctx context.Context, id *URLListIdentifier) error {
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.Projects.Locations.UrlLists.Delete(id.FullyQualifiedName()).Do()
	if err != nil {
		var gapierr *googleapi.Error
		if errors.As(err, &gapierr) && gapierr.Code == http.StatusNotFound {
			return nil
		}

		return err
	}

	return nil
}

func (c *Client) GetSecurityPolicyRule(ctx context.Context, id *PolicyRuleIdentifier) (*GatewaySecurityPolicyRule, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	raw, err := client.Projects.Locations.GatewaySecurityPolicies.Rules.Get(id.FullyQualifiedName()).Do()
	if err != nil {
		var gapierr *googleapi.Error
		if errors.As(err, &gapierr) && gapierr.Code == http.StatusNotFound {
			return nil, ErrNotExist
		}

		return nil, err
	}

	return &GatewaySecurityPolicyRule{
		ApplicationMatcher:   raw.ApplicationMatcher,
		BasicProfile:         raw.BasicProfile,
		CreateTime:           raw.CreateTime,
		Description:          raw.Description,
		Enabled:              raw.Enabled,
		Name:                 raw.Name,
		Priority:             raw.Priority,
		SessionMatcher:       raw.SessionMatcher,
		TlsInspectionEnabled: raw.TlsInspectionEnabled,
		UpdateTime:           raw.UpdateTime,
	}, nil
}

func (c *Client) CreateSecurityPolicyRule(ctx context.Context, opts *PolicyRuleCreateOpts) error {
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.Projects.Locations.GatewaySecurityPolicies.Rules.Create(opts.ID.Parent(), &networksecurity.GatewaySecurityPolicyRule{
		Name:                 opts.ID.FullyQualifiedName(),
		ApplicationMatcher:   opts.Rule.ApplicationMatcher,
		BasicProfile:         opts.Rule.BasicProfile,
		Description:          opts.Rule.Description,
		Enabled:              opts.Rule.Enabled,
		Priority:             opts.Rule.Priority,
		SessionMatcher:       opts.Rule.SessionMatcher,
		TlsInspectionEnabled: opts.Rule.TlsInspectionEnabled,
	}).Do()
	if err != nil {
		var gapierr *googleapi.Error
		if errors.As(err, &gapierr) && gapierr.Code == http.StatusConflict {
			return ErrExist
		}

		return err
	}

	return nil
}

func (c *Client) DeleteSecurityPolicyRule(ctx context.Context, id *PolicyRuleIdentifier) error {
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.Projects.Locations.GatewaySecurityPolicies.Rules.Delete(id.FullyQualifiedName()).Do()
	if err != nil {
		var gapierr *googleapi.Error
		if errors.As(err, &gapierr) && gapierr.Code == http.StatusNotFound {
			return nil
		}

		return err
	}

	return nil
}

func (c *Client) newClient(ctx context.Context) (*networksecurity.Service, error) {
	var options []option.ClientOption

	if c.disableAuth {
		options = append(options, option.WithoutAuthentication())
	}

	if c.apiEndpoint != "" {
		options = append(options, option.WithEndpoint(c.apiEndpoint))
	}

	client, err := networksecurity.NewService(ctx, options...)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func New(apiEndpoint string, disableAuth bool) *Client {
	return &Client{
		apiEndpoint: apiEndpoint,
		disableAuth: disableAuth,
	}
}
