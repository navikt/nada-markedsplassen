package gcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/sa"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.ServiceAccountAPI = &serviceAccountAPI{}

type serviceAccountAPI struct {
	ops sa.Operations
}

func (a *serviceAccountAPI) ListServiceAccounts(ctx context.Context, gcpProject string) ([]*service.ServiceAccount, error) {
	const op errs.Op = "serviceAccountAPI.ListServiceAccounts"

	raw, err := a.ops.ListServiceAccounts(ctx, gcpProject)
	if err != nil {
		return nil, errs.E(errs.IO, op, err)
	}

	var accounts []*service.ServiceAccount

	for _, r := range raw {
		account := &service.ServiceAccount{
			ServiceAccountMeta: r,
		}

		keys, err := a.ops.ListServiceAccountKeys(ctx, r.Name)
		if err != nil {
			return nil, errs.E(errs.IO, op, fmt.Errorf("listing service account keys '%s': %w", r.Name, err))
		}

		for _, key := range keys {
			account.Keys = append(account.Keys, &service.ServiceAccountKey{
				Name:         key.Name,
				KeyAlgorithm: key.KeyAlgorithm,
				KeyOrigin:    key.KeyOrigin,
				KeyType:      key.KeyType,
			})
		}

		bindings, err := a.ops.ListProjectServiceAccountPolicyBindings(ctx, r.ProjectId, r.Email)
		if err != nil {
			return nil, errs.E(errs.IO, op, fmt.Errorf("listing project service account policy bindings '%s': %w", r.Email, err))
		}

		for _, binding := range bindings {
			account.Bindings = append(account.Bindings, &service.Binding{
				Role:    binding.Role,
				Members: binding.Members,
			})
		}
	}

	return accounts, nil
}

func (a *serviceAccountAPI) DeleteServiceAccount(ctx context.Context, project, email string) error {
	const op errs.Op = "serviceAccountAPI.DeleteServiceAccount"

	name := sa.ServiceAccountNameFromEmail(project, email)

	err := a.ops.DeleteServiceAccount(ctx, name)
	if err != nil {
		if errors.Is(err, sa.ErrNotFound) {
			return nil
		}

		return errs.E(errs.IO, op, fmt.Errorf("deleting service account '%s': %w", name, err))
	}

	return nil
}

func (a *serviceAccountAPI) DeleteServiceAccountAndBindings(ctx context.Context, project, email string) error {
	const op errs.Op = "serviceAccountAPI.DeleteServiceAccountAndBinding"

	err := a.ops.RemoveProjectServiceAccountPolicyBinding(ctx, project, email)
	if err != nil {
		return errs.E(errs.IO, op, err)
	}

	err = a.DeleteServiceAccount(ctx, project, email)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (a *serviceAccountAPI) EnsureServiceAccountWithBindings(ctx context.Context, sa *service.ServiceAccountRequestWithBindings) (*service.ServiceAccountMeta, error) {
	const op errs.Op = "serviceAccountAPI.EnsureServiceAccount"

	accountMeta, err := a.ensureServiceAccountExists(ctx, &service.ServiceAccountRequest{
		ProjectID:   sa.ProjectID,
		AccountID:   sa.AccountID,
		DisplayName: sa.DisplayName,
		Description: sa.Description,
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	for _, binding := range sa.Bindings {
		err := a.ensureServiceAccountProjectBinding(ctx, sa.ProjectID, binding)
		if err != nil {
			return nil, errs.E(op, err)
		}
	}

	return accountMeta, nil
}

func (a *serviceAccountAPI) EnsureServiceAccountWithKeyAndBinding(ctx context.Context, req *service.ServiceAccountRequestWithBinding) (*service.ServiceAccountWithPrivateKey, error) {
	const op errs.Op = "serviceAccountAPI.EnsureServiceAccountWithKeyAndBinding"

	accountMeta, err := a.ensureServiceAccountExists(ctx, &req.ServiceAccountRequest)
	if err != nil {
		return nil, errs.E(op, err)
	}

	if req.Binding != nil {
		err = a.ensureServiceAccountProjectBinding(ctx, req.ProjectID, req.Binding)
		if err != nil {
			return nil, errs.E(op, err)
		}
	}

	key, err := a.ensureServiceAccountKey(ctx, accountMeta.Name)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &service.ServiceAccountWithPrivateKey{
		ServiceAccountMeta: accountMeta,
		Key:                key,
	}, nil
}

func (a *serviceAccountAPI) ensureServiceAccountKey(ctx context.Context, name string) (*service.ServiceAccountKeyWithPrivateKeyData, error) {
	const op errs.Op = "serviceAccountAPI.ensureServiceAccountKey"

	keys, err := a.ops.ListServiceAccountKeys(ctx, name)
	if err != nil {
		return nil, errs.E(errs.IO, op, fmt.Errorf("listing service account keys '%s': %w", name, err))
	}

	for _, key := range keys {
		if key.KeyType == "USER_MANAGED" {
			err := a.ops.DeleteServiceAccountKey(ctx, key.Name)
			if err != nil {
				return nil, errs.E(errs.IO, op, fmt.Errorf("deleting service account key '%s': %w", key.Name, err))
			}
		}
	}

	key, err := a.ops.CreateServiceAccountKey(ctx, name)
	if err != nil {
		return nil, errs.E(errs.IO, op, fmt.Errorf("creating service account key '%s': %w", name, err))
	}

	return &service.ServiceAccountKeyWithPrivateKeyData{
		ServiceAccountKey: &service.ServiceAccountKey{
			Name:         key.Name,
			KeyAlgorithm: key.KeyAlgorithm,
			KeyOrigin:    key.KeyOrigin,
			KeyType:      key.KeyType,
		},
		PrivateKeyData: key.PrivateKeyData,
	}, nil
}

func (a *serviceAccountAPI) ensureServiceAccountProjectBinding(ctx context.Context, project string, binding *service.Binding) error {
	const op errs.Op = "serviceAccountAPI.ensureServiceAccountProjectBinding"

	err := a.ops.AddProjectServiceAccountPolicyBinding(ctx, project, &sa.Binding{
		Role:    binding.Role,
		Members: binding.Members,
	})
	if err != nil {
		return errs.E(errs.IO, op, fmt.Errorf("adding project service account policy binding '%s': %w", project, err))
	}

	return nil
}

func (a *serviceAccountAPI) ensureServiceAccountExists(ctx context.Context, req *service.ServiceAccountRequest) (*service.ServiceAccountMeta, error) {
	const op errs.Op = "serviceAccountAPI.ensureServiceAccountExists"

	account, err := a.ops.GetServiceAccount(ctx, sa.ServiceAccountNameFromAccountID(req.ProjectID, req.AccountID))
	if err == nil {
		return account, nil
	}

	if !errors.Is(err, sa.ErrNotFound) {
		return nil, errs.E(errs.IO, op, fmt.Errorf("getting service account '%s': %w", sa.ServiceAccountNameFromAccountID(req.ProjectID, req.AccountID), err))
	}

	request := &sa.ServiceAccountRequest{
		ProjectID:   req.ProjectID,
		AccountID:   req.AccountID,
		DisplayName: req.DisplayName,
		Description: req.Description,
	}

	account, err = a.ops.CreateServiceAccount(ctx, request)
	if err != nil {
		return nil, errs.E(errs.IO, op, fmt.Errorf("creating service account '%s': %w", sa.ServiceAccountNameFromAccountID(req.ProjectID, req.AccountID), err))
	}

	return account, nil
}

func NewServiceAccountAPI(ops sa.Operations) *serviceAccountAPI {
	return &serviceAccountAPI{
		ops: ops,
	}
}
