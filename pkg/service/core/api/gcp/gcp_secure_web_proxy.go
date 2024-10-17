package gcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/securewebproxy"
	"github.com/navikt/nada-backend/pkg/service"
	"golang.org/x/exp/rand"
)

const (
	maxRetries = 10
)

var _ service.SecureWebProxyAPI = &secureWebProxyAPI{}

type secureWebProxyAPI struct {
	ops securewebproxy.Operations
}

func (a *secureWebProxyAPI) EnsureURLList(ctx context.Context, opts *service.URLListEnsureOpts) error {
	const op errs.Op = "secureWebProxyAPI.EnsureURLList"

	_, err := a.GetURLList(ctx, &service.URLListIdentifier{
		Project:  opts.ID.Project,
		Location: opts.ID.Location,
		Slug:     opts.ID.Slug,
	})
	if err != nil {
		if errs.KindIs(errs.NotExist, err) {
			err = a.CreateURLList(ctx, &service.URLListCreateOpts{
				ID:          opts.ID,
				Description: opts.Description,
				URLS:        opts.URLS,
			})
			if err != nil {
				return errs.E(op, err)
			}

			return nil
		}

		return errs.E(op, err)
	}

	err = a.UpdateURLList(ctx, &service.URLListUpdateOpts{
		ID:          opts.ID,
		Description: opts.Description,
		URLS:        opts.URLS,
	})
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (a *secureWebProxyAPI) GetURLList(ctx context.Context, id *service.URLListIdentifier) ([]string, error) {
	const op errs.Op = "secureWebProxyAPI.GetURLList"

	raw, err := a.ops.GetURLList(ctx, &securewebproxy.URLListIdentifier{
		Project:  id.Project,
		Location: id.Location,
		Slug:     id.Slug,
	})
	if err != nil {
		if errors.Is(err, securewebproxy.ErrNotExist) {
			return nil, errs.E(errs.NotExist, op, fmt.Errorf("urllist for slug %s.%s.%s not found: %w", id.Project, id.Location, id.Slug, err))
		}

		return nil, errs.E(errs.IO, op, fmt.Errorf("get urllist for slug %s.%s.%s: %w", id.Project, id.Location, id.Slug, err))
	}

	return raw, nil
}

func (a *secureWebProxyAPI) CreateURLList(ctx context.Context, opts *service.URLListCreateOpts) error {
	const op errs.Op = "secureWebProxyAPI.CreateURLList"

	err := a.ops.CreateURLList(ctx, &securewebproxy.URLListCreateOpts{
		ID: &securewebproxy.URLListIdentifier{
			Project:  opts.ID.Project,
			Location: opts.ID.Location,
			Slug:     opts.ID.Slug,
		},
		Description: opts.Description,
		URLS:        opts.URLS,
	})
	if err != nil {
		if errors.Is(err, securewebproxy.ErrExist) {
			return errs.E(errs.Exist, op, fmt.Errorf("urllist for slug %s.%s.%s already exist: %w", opts.ID.Project, opts.ID.Project, opts.ID.Slug, err))
		}

		return errs.E(errs.IO, op, fmt.Errorf("create urllist for slug %s.%s.%s: %w", opts.ID.Project, opts.ID.Project, opts.ID.Slug, err))
	}

	return nil
}

func (a *secureWebProxyAPI) UpdateURLList(ctx context.Context, opts *service.URLListUpdateOpts) error {
	const op errs.Op = "secureWebProxyAPI.UpdateURLList"

	err := a.ops.UpdateURLList(ctx, &securewebproxy.URLListUpdateOpts{
		ID: &securewebproxy.URLListIdentifier{
			Project:  opts.ID.Project,
			Location: opts.ID.Location,
			Slug:     opts.ID.Slug,
		},
		Description: opts.Description,
		URLS:        opts.URLS,
	})
	if err != nil {
		if errors.Is(err, securewebproxy.ErrNotExist) {
			return errs.E(errs.NotExist, op, fmt.Errorf("urllist for slug %s.%s.%s does not exist: %w", opts.ID.Project, opts.ID.Project, opts.ID.Slug, err))
		}

		return errs.E(errs.IO, op, fmt.Errorf("update urllist for slug %s.%s.%s: %w", opts.ID.Project, opts.ID.Project, opts.ID.Slug, err))
	}

	return nil
}

func (a *secureWebProxyAPI) DeleteURLList(ctx context.Context, id *service.URLListIdentifier) error {
	const op errs.Op = "secureWebProxyAPI.DeleteURLList"

	err := a.ops.DeleteURLList(ctx, &securewebproxy.URLListIdentifier{
		Project:  id.Project,
		Location: id.Location,
		Slug:     id.Slug,
	})
	if err != nil {
		if errors.Is(err, securewebproxy.ErrNotExist) {
			return nil
		}

		return errs.E(errs.IO, op, fmt.Errorf("delete urllist for slug %s.%s.%s: %w", id.Project, id.Project, id.Slug, err))
	}

	return nil
}

func (a *secureWebProxyAPI) EnsureSecurityPolicyRuleWithRandomPriority(ctx context.Context, opts *service.PolicyRuleEnsureNextAvailablePortOpts) error {
	const op errs.Op = "secureWebProxyAPI.EnsureSecurityPolicyRuleWithRandomPriority"

	existingRule, err := a.GetSecurityPolicyRule(ctx, &service.PolicyRuleIdentifier{
		Project:  opts.ID.Project,
		Location: opts.ID.Location,
		Policy:   opts.ID.Policy,
		Slug:     opts.Name,
	})
	if errs.KindIs(errs.NotExist, err) {
		// Find a random priority between the min and max range
		for i := 0; i < maxRetries; i++ {
			priority := rand.Intn(opts.PriorityMaxRange-opts.PriorityMinRange+1) + opts.PriorityMinRange

			err = a.CreateSecurityPolicyRule(ctx, &service.PolicyRuleCreateOpts{
				ID: &service.PolicyRuleIdentifier{
					Project:  opts.ID.Project,
					Location: opts.ID.Location,
					Policy:   opts.ID.Policy,
					Slug:     opts.Name,
				},
				Rule: &service.GatewaySecurityPolicyRule{
					ApplicationMatcher:   opts.ApplicationMatcher,
					BasicProfile:         opts.BasicProfile,
					Description:          opts.Description,
					Enabled:              opts.Enabled,
					Name:                 opts.Name,
					Priority:             int64(priority),
					SessionMatcher:       opts.SessionMatcher,
					TlsInspectionEnabled: opts.TlsInspectionEnabled,
				},
			})
			if err != nil {
				if !errs.KindIs(errs.Exist, err) {
					return errs.E(op, err)
				}

				// Rule with priority already exists, try again
				continue
			}
		}

		return nil
	}

	err = a.UpdateSecurityPolicyRule(ctx, &service.PolicyRuleUpdateOpts{
		ID: &service.PolicyRuleIdentifier{
			Project:  opts.ID.Project,
			Location: opts.ID.Location,
			Policy:   opts.ID.Policy,
			Slug:     opts.Name,
		},
		Rule: &service.GatewaySecurityPolicyRule{
			ApplicationMatcher:   opts.ApplicationMatcher,
			BasicProfile:         opts.BasicProfile,
			Description:          opts.Description,
			Enabled:              opts.Enabled,
			Name:                 opts.Name,
			Priority:             existingRule.Priority,
			SessionMatcher:       opts.SessionMatcher,
			TlsInspectionEnabled: opts.TlsInspectionEnabled,
		},
	})
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (a *secureWebProxyAPI) GetSecurityPolicyRule(ctx context.Context, id *service.PolicyRuleIdentifier) (*service.GatewaySecurityPolicyRule, error) {
	const op errs.Op = "secureWebProxyAPI.GetSecurityPolicyRule"

	raw, err := a.ops.GetSecurityPolicyRule(ctx, &securewebproxy.PolicyRuleIdentifier{
		Project:  id.Project,
		Location: id.Location,
		Slug:     id.Slug,
		Policy:   id.Policy,
	})
	if err != nil {
		if errors.Is(err, securewebproxy.ErrNotExist) {
			return nil, errs.E(errs.NotExist, op, fmt.Errorf("security policy for slug %s.%s.%s does not exist: %w", id.Project, id.Project, id.Slug, err))
		}

		return nil, errs.E(errs.IO, op, fmt.Errorf("getting security policy for slug %s.%s.%s: %w", id.Project, id.Project, id.Slug, err))
	}

	return &service.GatewaySecurityPolicyRule{
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

func (a *secureWebProxyAPI) CreateSecurityPolicyRule(ctx context.Context, opts *service.PolicyRuleCreateOpts) error {
	const op errs.Op = "secureWebProxyAPI.CreateSecurityPolicyRule"

	err := a.ops.CreateSecurityPolicyRule(ctx, &securewebproxy.PolicyRuleCreateOpts{
		ID: &securewebproxy.PolicyRuleIdentifier{
			Project:  opts.ID.Project,
			Location: opts.ID.Location,
			Policy:   opts.ID.Policy,
			Slug:     opts.ID.Slug,
		},
		Rule: &securewebproxy.GatewaySecurityPolicyRule{
			ApplicationMatcher:   opts.Rule.ApplicationMatcher,
			BasicProfile:         opts.Rule.BasicProfile,
			CreateTime:           opts.Rule.CreateTime,
			Description:          opts.Rule.Description,
			Enabled:              opts.Rule.Enabled,
			Name:                 opts.Rule.Name,
			Priority:             opts.Rule.Priority,
			SessionMatcher:       opts.Rule.SessionMatcher,
			TlsInspectionEnabled: opts.Rule.TlsInspectionEnabled,
			UpdateTime:           opts.Rule.UpdateTime,
		},
	})
	if err != nil {
		if errors.Is(err, securewebproxy.ErrExist) {
			return errs.E(errs.Exist, op, fmt.Errorf("security policy for slug %s.%s.%s already exist: %w", opts.ID.Project, opts.ID.Project, opts.ID.Slug, err))
		}

		return errs.E(errs.IO, op, fmt.Errorf("create security policy for slug %s.%s.%s: %w", opts.ID.Project, opts.ID.Project, opts.ID.Slug, err))
	}

	return nil
}

func (a *secureWebProxyAPI) UpdateSecurityPolicyRule(ctx context.Context, opts *service.PolicyRuleUpdateOpts) error {
	const op errs.Op = "secureWebProxyAPI.UpdateSecurityPolicyRule"

	err := a.ops.UpdateSecurityPolicyRule(ctx, &securewebproxy.PolicyRuleCreateOpts{
		ID: &securewebproxy.PolicyRuleIdentifier{
			Project:  opts.ID.Project,
			Location: opts.ID.Location,
			Policy:   opts.ID.Policy,
			Slug:     opts.ID.Slug,
		},
		Rule: &securewebproxy.GatewaySecurityPolicyRule{
			ApplicationMatcher:   opts.Rule.ApplicationMatcher,
			BasicProfile:         opts.Rule.BasicProfile,
			CreateTime:           opts.Rule.CreateTime,
			Description:          opts.Rule.Description,
			Enabled:              opts.Rule.Enabled,
			Name:                 opts.Rule.Name,
			Priority:             opts.Rule.Priority,
			SessionMatcher:       opts.Rule.SessionMatcher,
			TlsInspectionEnabled: opts.Rule.TlsInspectionEnabled,
			UpdateTime:           opts.Rule.UpdateTime,
		},
	})
	if err != nil {
		if errors.Is(err, securewebproxy.ErrNotExist) {
			return errs.E(errs.Exist, op, fmt.Errorf("security policy for slug %s.%s.%s does not exist: %w", opts.ID.Project, opts.ID.Project, opts.ID.Slug, err))
		}

		return errs.E(errs.IO, op, fmt.Errorf("update security policy for slug %s.%s.%s: %w", opts.ID.Project, opts.ID.Project, opts.ID.Slug, err))
	}

	return nil
}

func (a *secureWebProxyAPI) DeleteSecurityPolicyRule(ctx context.Context, id *service.PolicyRuleIdentifier) error {
	const op errs.Op = "secureWebProxyAPI.DeleteSecurityPolicyRule"

	err := a.ops.DeleteSecurityPolicyRule(ctx, &securewebproxy.PolicyRuleIdentifier{
		Project:  id.Project,
		Location: id.Location,
		Slug:     id.Slug,
		Policy:   id.Policy,
	})
	if err != nil {
		if errors.Is(err, securewebproxy.ErrNotExist) {
			return nil
		}

		return errs.E(errs.IO, op, fmt.Errorf("delete security policy for slug %s.%s.%s: %w", id.Project, id.Project, id.Slug, err))
	}

	return nil
}

func NewSecureWebProxyAPI(ops securewebproxy.Operations) *secureWebProxyAPI {
	return &secureWebProxyAPI{
		ops: ops,
	}
}
