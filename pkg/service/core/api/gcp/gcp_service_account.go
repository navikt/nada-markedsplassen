package gcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/sa"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.ServiceAccountAPI = &serviceAccountAPI{}

type serviceAccountAPI struct {
	ops sa.Operations
}

func (a *serviceAccountAPI) AddServiceAccountPolicyBinding(ctx context.Context, project, saEmail string, binding *service.Binding) error {
	const op errs.Op = "serviceAccountAPI.AddServiceAccountPolicyBinding"

	err := a.ops.AddServiceAccountPolicyBinding(ctx, project, saEmail, &sa.Binding{
		Role:    binding.Role,
		Members: binding.Members,
	})
	if err != nil {
		if errors.Is(err, sa.ErrNotFound) {
			return errs.E(errs.NotExist, service.CodeGCPServiceAccount, op, err, service.ParamServiceAccount)
		}

		return errs.E(errs.IO, service.CodeGCPServiceAccount, op, fmt.Errorf("adding role binding '%s' to service account '%s': %w", binding.Role, saEmail, err))
	}

	return nil
}

func (a *serviceAccountAPI) RemoveServiceAccountPolicyBinding(ctx context.Context, project, email string, binding *service.Binding) error {
	const op errs.Op = "serviceAccountAPI.RemoveServiceAccountPolicyBinding"

	err := a.ops.RemoveServiceAccountPolicyBinding(ctx, project, email, &sa.Binding{
		Role:    binding.Role,
		Members: binding.Members,
	})
	if err != nil {
		if errors.Is(err, sa.ErrNotFound) {
			return errs.E(errs.NotExist, service.CodeGCPServiceAccount, op, err, service.ParamServiceAccount)
		}

		return errs.E(errs.IO, op, fmt.Errorf("removing role binding '%s' from service account '%s': %w", binding.Role, email, err))
	}

	return nil
}

func (a *serviceAccountAPI) ListServiceAccounts(ctx context.Context, gcpProject string) ([]*service.ServiceAccount, error) {
	const op errs.Op = "serviceAccountAPI.ListServiceAccounts"

	raw, err := a.ops.ListServiceAccounts(ctx, gcpProject)
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeGCPServiceAccount, op, err)
	}

	var accounts []*service.ServiceAccount

	for _, r := range raw {
		account := &service.ServiceAccount{
			ServiceAccountMeta: &service.ServiceAccountMeta{
				Description: r.Description,
				DisplayName: r.DisplayName,
				Email:       r.Email,
				Name:        r.Name,
				ProjectId:   r.ProjectId,
				UniqueId:    r.UniqueId,
			},
		}

		keys, err := a.ops.ListServiceAccountKeys(ctx, r.Name)
		if err != nil {
			return nil, errs.E(errs.IO, service.CodeGCPServiceAccount, op, fmt.Errorf("listing service account keys '%s': %w", r.Name, err))
		}

		for _, key := range keys {
			account.Keys = append(account.Keys, &service.ServiceAccountKey{
				Name:         key.Name,
				KeyAlgorithm: key.KeyAlgorithm,
				KeyOrigin:    key.KeyOrigin,
				KeyType:      key.KeyType,
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

		return errs.E(errs.IO, service.CodeGCPServiceAccount, op, fmt.Errorf("deleting service account '%s': %w", name, err))
	}

	return nil
}

func (a *serviceAccountAPI) EnsureServiceAccount(ctx context.Context, sa *service.ServiceAccountRequest) (*service.ServiceAccountMeta, error) {
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

	return accountMeta, nil
}

func (a *serviceAccountAPI) EnsureServiceAccountWithKey(ctx context.Context, req *service.ServiceAccountRequest) (*service.ServiceAccountWithPrivateKey, error) {
	const op errs.Op = "serviceAccountAPI.EnsureServiceAccountWithKey"

	accountMeta, err := a.ensureServiceAccountExists(ctx, req)
	if err != nil {
		return nil, errs.E(op, err)
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
		return nil, errs.E(errs.IO, service.CodeGCPServiceAccount, op, fmt.Errorf("listing service account keys '%s': %w", name, err))
	}

	for _, key := range keys {
		if key.KeyType == "USER_MANAGED" {
			err := a.ops.DeleteServiceAccountKey(ctx, key.Name)
			if err != nil {
				return nil, errs.E(errs.IO, service.CodeGCPServiceAccount, op, fmt.Errorf("deleting service account key '%s': %w", key.Name, err))
			}
		}
	}

	key, err := a.ops.CreateServiceAccountKey(ctx, name)
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeGCPServiceAccount, op, fmt.Errorf("creating service account key '%s': %w", name, err))
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

func (a *serviceAccountAPI) ensureServiceAccountExists(ctx context.Context, req *service.ServiceAccountRequest) (*service.ServiceAccountMeta, error) {
	const op errs.Op = "serviceAccountAPI.ensureServiceAccountExists"

	account, err := a.ops.GetServiceAccount(ctx, sa.ServiceAccountNameFromAccountID(req.ProjectID, req.AccountID))
	if err == nil {
		return &service.ServiceAccountMeta{
			Description: account.Description,
			DisplayName: account.DisplayName,
			Email:       account.Email,
			Name:        account.Name,
			ProjectId:   account.ProjectId,
			UniqueId:    account.UniqueId,
		}, nil
	}

	if !errors.Is(err, sa.ErrNotFound) {
		return nil, errs.E(errs.IO, service.CodeGCPServiceAccount, op, fmt.Errorf("getting service account '%s': %w", sa.ServiceAccountNameFromAccountID(req.ProjectID, req.AccountID), err))
	}

	request := &sa.ServiceAccountRequest{
		ProjectID:   req.ProjectID,
		AccountID:   req.AccountID,
		DisplayName: req.DisplayName,
		Description: req.Description,
	}

	account, err = a.ops.CreateServiceAccount(ctx, request)
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeGCPServiceAccount, op, fmt.Errorf("creating service account '%s': %w", sa.ServiceAccountNameFromAccountID(req.ProjectID, req.AccountID), err))
	}

	// Ensure service account is ready
	time.Sleep(time.Minute)

	return &service.ServiceAccountMeta{
		Description: account.Description,
		DisplayName: account.DisplayName,
		Email:       account.Email,
		Name:        account.Name,
		ProjectId:   account.ProjectId,
		UniqueId:    account.UniqueId,
	}, nil
}

func NewServiceAccountAPI(ops sa.Operations) *serviceAccountAPI {
	return &serviceAccountAPI{
		ops: ops,
	}
}
