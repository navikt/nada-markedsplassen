package service

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ServiceAccountAPI interface {
	// ListServiceAccounts returns a list of service accounts in the project
	// with their keys and role bindings.
	ListServiceAccounts(ctx context.Context, project string) ([]*ServiceAccount, error)

	// EnsureServiceAccount creates a service account in the project and adds the specified role bindings
	EnsureServiceAccount(ctx context.Context, sa *ServiceAccountRequest) (*ServiceAccountMeta, error)

	// EnsureServiceAccountWithKey creates a service account in the project, and generates a keypair
	EnsureServiceAccountWithKey(ctx context.Context, sa *ServiceAccountRequest) (*ServiceAccountWithPrivateKey, error)

	// DeleteServiceAccount deletes the service account
	DeleteServiceAccount(ctx context.Context, project, email string) error
}

type ServiceAccountRequest struct {
	ProjectID   string
	AccountID   string
	DisplayName string
	Description string
}

func (s ServiceAccountRequest) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.ProjectID, validation.Required),
		validation.Field(&s.AccountID, validation.Required),
		validation.Field(&s.DisplayName, validation.Required),
		validation.Field(&s.Description, validation.Required),
	)
}

type ServiceAccountMeta struct {
	Description string
	DisplayName string
	Email       string
	Name        string
	ProjectId   string
	UniqueId    string
}

type ServiceAccount struct {
	*ServiceAccountMeta
	Keys []*ServiceAccountKey
}

type ServiceAccountWithPrivateKey struct {
	*ServiceAccountMeta
	Key *ServiceAccountKeyWithPrivateKeyData
}

type ServiceAccountKey struct {
	Name         string
	KeyAlgorithm string
	KeyOrigin    string
	KeyType      string
}

type ServiceAccountKeyWithPrivateKeyData struct {
	*ServiceAccountKey
	PrivateKeyData []byte
}
