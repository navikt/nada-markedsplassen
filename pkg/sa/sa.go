package sa

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/exp/maps"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

const (
	DeletedPrefix = "deleted:"
)

var ErrNotFound = errors.New("not found")

type Operations interface {
	GetServiceAccount(ctx context.Context, name string) (*ServiceAccount, error)
	CreateServiceAccount(ctx context.Context, sa *ServiceAccountRequest) (*ServiceAccount, error)
	DeleteServiceAccount(ctx context.Context, name string) error
	ListServiceAccounts(ctx context.Context, project string) ([]*ServiceAccount, error)
	AddServiceAccountPolicyBinding(ctx context.Context, project, saEmail string, binding *Binding) error
	RemoveServiceAccountPolicyBinding(ctx context.Context, project, email string, binding *Binding) error
	CreateServiceAccountKey(ctx context.Context, name string) (*ServiceAccountKeyWithPrivateKeyData, error)
	DeleteServiceAccountKey(ctx context.Context, name string) error
	ListServiceAccountKeys(ctx context.Context, name string) ([]*ServiceAccountKey, error)
}

type ServiceAccountKey struct {
	Name           string
	KeyAlgorithm   string
	KeyOrigin      string
	KeyType        string
	ValidAfterTime string
}

type ServiceAccountKeyWithPrivateKeyData struct {
	*ServiceAccountKey
	PrivateKeyData []byte
}

type Binding struct {
	Role    string
	Members []string
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

type ServiceAccount struct {
	Description string
	DisplayName string
	Email       string
	Name        string
	ProjectId   string
	UniqueId    string
}

var _ Operations = &Client{}

type Client struct {
	endpoint    string
	disableAuth bool
}

func (c *Client) RemoveServiceAccountPolicyBinding(ctx context.Context, project string, saEmail string, remove *Binding) error {
	service, err := c.iamService(ctx)
	if err != nil {
		return err
	}

	policy, err := service.Projects.ServiceAccounts.GetIamPolicy(ServiceAccountNameFromEmail(project, saEmail)).Do()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return fmt.Errorf("project %s: %w", project, ErrNotFound)
		}

		return fmt.Errorf("getting project %s policy: %w", project, err)
	}

	var bindings []*iam.Binding

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
			bindings = append(bindings, &iam.Binding{
				Role:    existing.Role,
				Members: existingMembers,
			})
		}
	}

	policy.Bindings = bindings

	_, err = service.Projects.ServiceAccounts.SetIamPolicy(ServiceAccountNameFromEmail(project, saEmail), &iam.SetIamPolicyRequest{
		Policy: policy,
	}).Do()
	if err != nil {
		return fmt.Errorf("setting project %s policy: %w", project, err)
	}

	return nil
}

func (c *Client) CreateServiceAccountKey(ctx context.Context, name string) (*ServiceAccountKeyWithPrivateKeyData, error) {
	service, err := c.iamService(ctx)
	if err != nil {
		return nil, err
	}

	key, err := service.Projects.ServiceAccounts.Keys.Create(name, &iam.CreateServiceAccountKeyRequest{}).Do()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return nil, fmt.Errorf("service account %s: %w", name, ErrNotFound)
		}

		return nil, fmt.Errorf("creating service account key %s: %w", name, err)
	}

	keyMatter, err := base64.StdEncoding.DecodeString(key.PrivateKeyData)
	if err != nil {
		return nil, fmt.Errorf("decoding private key data: %w", err)
	}

	return &ServiceAccountKeyWithPrivateKeyData{
		ServiceAccountKey: &ServiceAccountKey{
			Name:         key.Name,
			KeyAlgorithm: key.KeyAlgorithm,
			KeyOrigin:    key.KeyOrigin,
			KeyType:      key.KeyType,
		},
		PrivateKeyData: keyMatter,
	}, nil
}

func (c *Client) DeleteServiceAccountKey(ctx context.Context, name string) error {
	service, err := c.iamService(ctx)
	if err != nil {
		return err
	}

	_, err = service.Projects.ServiceAccounts.Keys.Delete(name).Do()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return fmt.Errorf("service account key %s: %w", name, ErrNotFound)
		}

		return fmt.Errorf("deleting service account key %s: %w", name, err)
	}

	return nil
}

func (c *Client) ListServiceAccountKeys(ctx context.Context, name string) ([]*ServiceAccountKey, error) {
	service, err := c.iamService(ctx)
	if err != nil {
		return nil, err
	}

	keys, err := service.Projects.ServiceAccounts.Keys.List(name).Do()
	if err != nil {
		return nil, fmt.Errorf("listing service account keys %s: %w", name, err)
	}

	result := make([]*ServiceAccountKey, len(keys.Keys))
	for i, key := range keys.Keys {
		result[i] = &ServiceAccountKey{
			Name:           key.Name,
			KeyAlgorithm:   key.KeyAlgorithm,
			KeyOrigin:      key.KeyOrigin,
			KeyType:        key.KeyType,
			ValidAfterTime: key.ValidAfterTime,
		}
	}

	return result, nil
}

func (c *Client) AddServiceAccountPolicyBinding(ctx context.Context, project, saEmail string, binding *Binding) error {
	service, err := c.iamService(ctx)
	if err != nil {
		return err
	}

	policy, err := service.Projects.ServiceAccounts.GetIamPolicy(ServiceAccountNameFromEmail(project, saEmail)).Do()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return fmt.Errorf("service account %s: %w", saEmail, ErrNotFound)
		}

		return fmt.Errorf("getting policy for %s.%s: %w", project, saEmail, err)
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
		policy.Bindings = append(policy.Bindings, &iam.Binding{
			Role:    binding.Role,
			Members: binding.Members,
		})
	}

	_, err = service.Projects.ServiceAccounts.SetIamPolicy(ServiceAccountNameFromEmail(project, saEmail), &iam.SetIamPolicyRequest{
		Policy: policy,
	}).Do()
	if err != nil {
		return fmt.Errorf("setting policy for %s.%s: %w", project, saEmail, err)
	}

	return nil
}

func (c *Client) ListServiceAccounts(ctx context.Context, project string) ([]*ServiceAccount, error) {
	service, err := c.iamService(ctx)
	if err != nil {
		return nil, err
	}

	raw, err := service.Projects.ServiceAccounts.List("projects/" + project).Do()
	if err != nil {
		return nil, fmt.Errorf("listing service accounts: %w", err)
	}

	accounts := make([]*ServiceAccount, len(raw.Accounts))
	for i, account := range raw.Accounts {
		accounts[i] = &ServiceAccount{
			Description: account.Description,
			DisplayName: account.DisplayName,
			Email:       account.Email,
			Name:        account.Name,
			ProjectId:   account.ProjectId,
			UniqueId:    account.UniqueId,
		}
	}

	return accounts, nil
}

func (c *Client) DeleteServiceAccount(ctx context.Context, name string) error {
	service, err := c.iamService(ctx)
	if err != nil {
		return err
	}

	_, err = service.Projects.ServiceAccounts.Delete(name).Do()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return fmt.Errorf("service account %s: %w", name, ErrNotFound)
		}

		return fmt.Errorf("deleting service account: %w", err)
	}

	return nil
}

func (c *Client) GetServiceAccount(ctx context.Context, name string) (*ServiceAccount, error) {
	service, err := c.iamService(ctx)
	if err != nil {
		return nil, err
	}

	account, err := service.Projects.ServiceAccounts.Get(name).Do()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return nil, fmt.Errorf("service account %s: %w", name, ErrNotFound)
		}

		return nil, fmt.Errorf("getting service account: %w", err)
	}

	return &ServiceAccount{
		Description: account.Description,
		DisplayName: account.DisplayName,
		Email:       account.Email,
		Name:        account.Name,
		ProjectId:   account.ProjectId,
		UniqueId:    account.UniqueId,
	}, nil
}

func (c *Client) CreateServiceAccount(ctx context.Context, sa *ServiceAccountRequest) (*ServiceAccount, error) {
	if err := sa.Validate(); err != nil {
		return nil, fmt.Errorf("validating service account request: %w", err)
	}

	service, err := c.iamService(ctx)
	if err != nil {
		return nil, err
	}

	request := &iam.CreateServiceAccountRequest{
		AccountId: sa.AccountID,
		ServiceAccount: &iam.ServiceAccount{
			Description: sa.Description,
			DisplayName: sa.DisplayName,
		},
	}

	account, err := service.Projects.ServiceAccounts.Create("projects/"+sa.ProjectID, request).Do()
	if err != nil {
		return nil, fmt.Errorf("creating service account: %w", err)
	}

	return &ServiceAccount{
		Description: account.Description,
		DisplayName: account.DisplayName,
		Email:       account.Email,
		Name:        account.Name,
		ProjectId:   account.ProjectId,
		UniqueId:    account.UniqueId,
	}, nil
}

func (c *Client) iamService(ctx context.Context) (*iam.Service, error) {
	var opts []option.ClientOption

	if c.disableAuth {
		opts = append(opts, option.WithoutAuthentication())
	}

	if len(c.endpoint) > 0 {
		opts = append(opts, option.WithEndpoint(c.endpoint))
	}

	service, err := iam.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating iam service: %w", err)
	}

	return service, nil
}

func NewClient(endpoint string, disableAuth bool) *Client {
	return &Client{
		endpoint:    endpoint,
		disableAuth: disableAuth,
	}
}

func ServiceAccountNameFromAccountID(project, accountID string) string {
	return "projects/" + project + "/serviceAccounts/" + accountID + "@" + project + ".iam.gserviceaccount.com"
}

func ServiceAccountNameFromEmail(project, email string) string {
	return "projects/" + project + "/serviceAccounts/" + email
}

func ServiceAccountKeyName(project, accountID, keyID string) string {
	return "projects/" + project + "/serviceAccounts/" + accountID + "@" + project + ".iam.gserviceaccount.com/keys/" + keyID
}
