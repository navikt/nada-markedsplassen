package service

import "context"

type EnsureProxyRuleWithURLList struct {
	// Project is the gcp project id
	Project string

	// Location is the gcp region
	Location string

	// Slug is the name of the url list
	Slug string
}

type SecureWebProxyAPI interface {
	EnsureURLList(ctx context.Context, opts *URLListEnsureOpts) error
	GetURLList(ctx context.Context, id *URLListIdentifier) ([]string, error)
	CreateURLList(ctx context.Context, opts *URLListCreateOpts) error
	UpdateURLList(ctx context.Context, opts *URLListUpdateOpts) error
	DeleteURLList(ctx context.Context, id *URLListIdentifier) error

	EnsureSecurityPolicyRule(ctx context.Context, opts *PolicyRuleEnsureOpts) error
	GetSecurityPolicyRule(ctx context.Context, id *PolicyRuleIdentifier) (*GatewaySecurityPolicyRule, error)
	CreateSecurityPolicyRule(ctx context.Context, opts *PolicyRuleCreateOpts) error
	UpdateSecurityPolicyRule(ctx context.Context, opts *PolicyRuleUpdateOpts) error
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

type PolicyRuleEnsureOpts struct {
	ID   *PolicyRuleIdentifier
	Rule *GatewaySecurityPolicyRule
}

type PolicyRuleCreateOpts struct {
	ID   *PolicyRuleIdentifier
	Rule *GatewaySecurityPolicyRule
}

type PolicyRuleUpdateOpts struct {
	ID   *PolicyRuleIdentifier
	Rule *GatewaySecurityPolicyRule
}

type URLListEnsureOpts struct {
	ID          *URLListIdentifier
	Description string
	URLS        []string
}

type URLListCreateOpts struct {
	ID          *URLListIdentifier
	Description string
	URLS        []string
}

type URLListUpdateOpts struct {
	ID          *URLListIdentifier
	Description string
	URLS        []string
}
