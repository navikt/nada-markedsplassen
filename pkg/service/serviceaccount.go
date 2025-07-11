package service

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	ServiceAccountUserRole = "roles/iam.serviceAccountUser"
)

type ServiceAccountAPI interface {
	// ListServiceAccounts returns a list of service accounts in the project
	// with their keys and role bindings.
	ListServiceAccounts(ctx context.Context, project string) ([]*ServiceAccount, error)

	// EnsureServiceAccount creates a service account in the project and adds the specified role bindings
	EnsureServiceAccount(ctx context.Context, sa *ServiceAccountRequest) (*ServiceAccountMeta, error)

	// EnsureServiceAccountWithKey creates a service account in the project, and generates a keypair
	EnsureServiceAccountWithKey(ctx context.Context, sa *ServiceAccountRequest) (*ServiceAccountWithPrivateKey, error)

	EnsureServiceAccountKey(ctx context.Context, name string) (*ServiceAccountKeyWithPrivateKeyData, error)

	// DeleteServiceAccount deletes the service account
	DeleteServiceAccount(ctx context.Context, project, email string) error

	// AddServiceAccountPolicyBinding adds a role binding to the service account
	AddServiceAccountPolicyBinding(ctx context.Context, project, saEmail string, binding *Binding) error

	// RemoveServiceAccountPolicyBinding removes a role binding from the service account
	RemoveServiceAccountPolicyBinding(ctx context.Context, project, email string, binding *Binding) error
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

func ServiceAccountEmailFromAccountID(project, accountID string) string {
	return accountID + "@" + project + ".iam.gserviceaccount.com"
}

func ServiceAccountNameFromEmail(project, email string) string {
	return "projects/" + project + "/serviceAccounts/" + email
}
