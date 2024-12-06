package sa_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/navikt/nada-backend/pkg/sa"
	"github.com/navikt/nada-backend/pkg/sa/emulator"
	"github.com/rs/zerolog"
	"google.golang.org/api/iam/v1"

	"github.com/stretchr/testify/assert"
)

func TestClient_CreateServiceAccount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		request   *sa.ServiceAccountRequest
		fn        func(*emulator.Emulator)
		expect    *sa.ServiceAccount
		expectErr string
	}{
		{
			name: "Create valid account",
			request: &sa.ServiceAccountRequest{
				ProjectID:   "test-project",
				AccountID:   "test-account",
				DisplayName: "Test Account",
				Description: "Test Description",
			},
			expect: &sa.ServiceAccount{
				Name:        sa.ServiceAccountNameFromAccountID("test-project", "test-account"),
				Email:       "test-account@test-project.iam.gserviceaccount.com",
				DisplayName: "Test Account",
				Description: "Test Description",
				ProjectId:   "test-project",
				UniqueId:    "1",
			},
		},
		{
			name: "Create account with missing project",
			request: &sa.ServiceAccountRequest{
				AccountID:   "test-account",
				DisplayName: "Test Account",
				Description: "Test Description",
			},
			expect:    nil,
			expectErr: "validating service account request: ProjectID: cannot be blank.",
		},
		{
			name: "Create account with failure",
			request: &sa.ServiceAccountRequest{
				ProjectID:   "test-project",
				AccountID:   "test-account",
				DisplayName: "Test Account",
				Description: "Test Description",
			},
			fn: func(em *emulator.Emulator) {
				em.SetError(fmt.Errorf("oops"))
			},
			expect:    nil,
			expectErr: "creating service account: googleapi: got HTTP response code 500 with body: oops\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.New(zerolog.NewConsoleWriter())
			em := emulator.New(log)
			url := em.Run()

			client := sa.NewClient(url, true)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			got, err := client.CreateServiceAccount(ctx, tc.request)

			if len(tc.expectErr) > 0 {
				require.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expect, got)
			}
		})
	}
}

func TestClient_GetServiceAccount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		project   string
		id        string
		fn        func(*emulator.Emulator)
		expect    *sa.ServiceAccount
		expectErr string
	}{
		{
			name:    "Get valid account",
			project: "test-project",
			id:      "test-account",
			fn: func(em *emulator.Emulator) {
				em.SetServiceAccount(sa.ServiceAccountNameFromAccountID("test-project", "test-account"), &iam.ServiceAccount{
					Name:        sa.ServiceAccountNameFromAccountID("test-project", "test-account"),
					Email:       "test-account@test-project.iam.gserviceaccount.com",
					DisplayName: "Test Account",
					Description: "Test Description",
					ProjectId:   "test-project",
					UniqueId:    "1",
				})
			},
			expect: &sa.ServiceAccount{
				Name:        sa.ServiceAccountNameFromAccountID("test-project", "test-account"),
				Email:       "test-account@test-project.iam.gserviceaccount.com",
				DisplayName: "Test Account",
				Description: "Test Description",
				ProjectId:   "test-project",
				UniqueId:    "1",
			},
		},
		{
			name:    "Get account with failure",
			project: "test-project",
			id:      "test-account",
			fn: func(em *emulator.Emulator) {
				em.SetError(fmt.Errorf("oops"))
			},
			expectErr: "getting service account: googleapi: got HTTP response code 500 with body: oops\n",
		},
		{
			name:      "Get account not found",
			project:   "test-project",
			id:        "test-account",
			expectErr: "service account projects/test-project/serviceAccounts/test-account@test-project.iam.gserviceaccount.com: not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.New(zerolog.NewConsoleWriter())
			em := emulator.New(log)

			url := em.Run()

			client := sa.NewClient(url, true)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			got, err := client.GetServiceAccount(ctx, sa.ServiceAccountNameFromAccountID(tc.project, tc.id))

			if len(tc.expectErr) > 0 {
				require.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expect, got)
			}
		})
	}
}

func TestClient_DeleteServiceAccount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		project   string
		id        string
		fn        func(*emulator.Emulator)
		expectErr string
	}{
		{
			name:    "Delete valid account",
			project: "test-project",
			id:      "test-account",
			fn: func(em *emulator.Emulator) {
				em.SetServiceAccount(sa.ServiceAccountNameFromAccountID("test-project", "test-account"), &iam.ServiceAccount{
					Name: sa.ServiceAccountNameFromAccountID("test-project", "test-account"),
				})
			},
		},
		{
			name:      "Delete account not found",
			project:   "test-project",
			id:        "test-account",
			expectErr: "service account projects/test-project/serviceAccounts/test-account@test-project.iam.gserviceaccount.com: not found",
		},
		{
			name:    "Delete account with failure",
			project: "test-project",
			id:      "test-account",
			fn: func(em *emulator.Emulator) {
				em.SetError(fmt.Errorf("oops"))
			},
			expectErr: "deleting service account: googleapi: got HTTP response code 500 with body: oops\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.New(zerolog.NewConsoleWriter())
			em := emulator.New(log)

			url := em.Run()

			client := sa.NewClient(url, true)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			err := client.DeleteServiceAccount(ctx, sa.ServiceAccountNameFromAccountID(tc.project, tc.id))

			if len(tc.expectErr) > 0 {
				require.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_ListServiceAccounts(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		project   string
		fn        func(*emulator.Emulator)
		expect    []*sa.ServiceAccount
		expectErr string
	}{
		{
			name:    "List accounts with no accounts",
			project: "test-project",
			expect:  make([]*sa.ServiceAccount, 0),
		},
		{
			name:    "List valid accounts",
			project: "test-project",
			fn: func(em *emulator.Emulator) {
				em.SetServiceAccount(sa.ServiceAccountNameFromAccountID("test-project", "test-account"), &iam.ServiceAccount{
					Name:        sa.ServiceAccountNameFromAccountID("test-project", "test-account"),
					Email:       "test-account@test-project.iam.gserviceaccount.com",
					DisplayName: "Test Account",
					Description: "Test Description",
					ProjectId:   "test-project",
					UniqueId:    "1234567890",
				})
				em.SetServiceAccount(sa.ServiceAccountNameFromAccountID("test-project", "test-account2"), &iam.ServiceAccount{
					Name:        sa.ServiceAccountNameFromAccountID("test-project", "test-account2"),
					Email:       "test-account2@test-project.iam.gserviceaccount.com",
					DisplayName: "Test Account 2",
					Description: "Test Description 2",
					ProjectId:   "test-project",
					UniqueId:    "1234567891",
				})
			},
			expect: []*sa.ServiceAccount{
				{
					Name:        sa.ServiceAccountNameFromAccountID("test-project", "test-account2"),
					Email:       "test-account2@test-project.iam.gserviceaccount.com",
					DisplayName: "Test Account 2",
					Description: "Test Description 2",
					ProjectId:   "test-project",
					UniqueId:    "1234567891",
				},
				{
					Name:        sa.ServiceAccountNameFromAccountID("test-project", "test-account"),
					Email:       "test-account@test-project.iam.gserviceaccount.com",
					DisplayName: "Test Account",
					Description: "Test Description",
					ProjectId:   "test-project",
					UniqueId:    "1234567890",
				},
			},
		},
		{
			name:    "List accounts with failure",
			project: "test-project",
			fn: func(em *emulator.Emulator) {
				em.SetError(fmt.Errorf("oops"))
			},
			expectErr: "listing service accounts: googleapi: got HTTP response code 500 with body: oops\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.New(zerolog.NewConsoleWriter())
			em := emulator.New(log)

			url := em.Run()

			client := sa.NewClient(url, true)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			got, err := client.ListServiceAccounts(ctx, tc.project)

			if len(tc.expectErr) > 0 {
				require.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expect, got)
			}
		})
	}
}

func TestClient_CreateServiceAccountKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		project   string
		account   string
		fn        func(*emulator.Emulator)
		expectErr string
		expect    *sa.ServiceAccountKeyWithPrivateKeyData
	}{
		{
			name:    "Create valid key",
			project: "test-project",
			account: "test-account",
			fn: func(em *emulator.Emulator) {
				em.SetServiceAccount(sa.ServiceAccountNameFromAccountID("test-project", "test-account"), &iam.ServiceAccount{
					Name:  sa.ServiceAccountNameFromAccountID("test-project", "test-account"),
					Email: "test-account@test-project.iam.gserviceaccount.com",
				})
			},
			expect: &sa.ServiceAccountKeyWithPrivateKeyData{
				ServiceAccountKey: &sa.ServiceAccountKey{
					Name:         "projects/test-project/serviceAccounts/test-account@test-project.iam.gserviceaccount.com/keys/1",
					KeyAlgorithm: "KEY_ALG_RSA_2048",
					KeyOrigin:    "GOOGLE_PROVIDED",
					KeyType:      "USER_MANAGED",
				},
			},
		},
		{
			name:    "Create key with failure",
			project: "test-project",
			account: "test-account",
			fn: func(em *emulator.Emulator) {
				em.SetError(fmt.Errorf("oops"))
			},
			expectErr: "creating service account key projects/test-project/serviceAccounts/test-account@test-project.iam.gserviceaccount.com: creating service account key projects/test-project/serviceAccounts/test-account@test-project.iam.gserviceaccount.com: googleapi: got HTTP response code 500 with body: oops\n",
		},
		{
			name:      "Create key with missing account",
			project:   "test-project",
			account:   "test-account",
			expectErr: "creating service account key projects/test-project/serviceAccounts/test-account@test-project.iam.gserviceaccount.com: creating service account key projects/test-project/serviceAccounts/test-account@test-project.iam.gserviceaccount.com: not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.New(zerolog.NewConsoleWriter())
			em := emulator.New(log)

			url := em.Run()

			client := sa.NewClient(url, true)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			got, err := client.CreateServiceAccountKey(ctx, sa.ServiceAccountNameFromAccountID(tc.project, tc.account))

			if len(tc.expectErr) > 0 {
				require.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			} else {
				require.NoError(t, err)
				diff := cmp.Diff(tc.expect, got, cmpopts.IgnoreFields(sa.ServiceAccountKeyWithPrivateKeyData{}, "PrivateKeyData"))
				assert.Empty(t, diff)
			}
		})
	}
}

func TestClient_DeleteServiceAccountKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		project   string
		account   string
		key       string
		fn        func(*emulator.Emulator)
		expectErr string
	}{
		{
			name:    "Delete valid key",
			project: "test-project",
			account: "test-account",
			key:     "1",
			fn: func(em *emulator.Emulator) {
				em.SetServiceAccountKeys(sa.ServiceAccountNameFromAccountID("test-project", "test-account"), []*iam.ServiceAccountKey{
					{
						Name: sa.ServiceAccountKeyName("test-project", "test-account", "1"),
					},
				})
			},
		},
		{
			name:    "Delete key with failure",
			project: "test-project",
			account: "test-account",
			key:     "1",
			fn: func(em *emulator.Emulator) {
				em.SetError(fmt.Errorf("oops"))
			},
			expectErr: "deleting service account key projects/test-project/serviceAccounts/test-account@test-project.iam.gserviceaccount.com/keys/1: googleapi: got HTTP response code 500 with body: oops\n",
		},
		{
			name:      "Delete key with missing account",
			project:   "test-project",
			account:   "test-account",
			key:       "1",
			expectErr: "service account key projects/test-project/serviceAccounts/test-account@test-project.iam.gserviceaccount.com/keys/1: not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.New(zerolog.NewConsoleWriter())
			em := emulator.New(log)

			url := em.Run()

			client := sa.NewClient(url, true)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			err := client.DeleteServiceAccountKey(ctx, sa.ServiceAccountKeyName(tc.project, tc.account, tc.key))

			if len(tc.expectErr) > 0 {
				require.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_ListServiceAccountKeys(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		project   string
		account   string
		fn        func(*emulator.Emulator)
		expect    []*sa.ServiceAccountKey
		expectErr string
	}{
		{
			name:    "List keys with no keys",
			project: "test-project",
			account: "test-account",
			expect:  make([]*sa.ServiceAccountKey, 0),
		},
		{
			name:    "List valid keys",
			project: "test-project",
			account: "test-account",
			fn: func(em *emulator.Emulator) {
				em.SetServiceAccountKeys(sa.ServiceAccountNameFromAccountID("test-project", "test-account"), []*iam.ServiceAccountKey{
					{
						Name:         sa.ServiceAccountKeyName("test-project", "test-account", "1"),
						KeyAlgorithm: "KEY_ALG_RSA_2048",
						KeyOrigin:    "GOOGLE_PROVIDED",
						KeyType:      "USER_MANAGED",
					},
					{
						Name:         sa.ServiceAccountKeyName("test-project", "test-account", "2"),
						KeyAlgorithm: "KEY_ALG_RSA_2048",
						KeyOrigin:    "GOOGLE_PROVIDED",
						KeyType:      "SYSTEM_MANAGED",
					},
				})
			},
			expect: []*sa.ServiceAccountKey{
				{
					Name:         sa.ServiceAccountKeyName("test-project", "test-account", "1"),
					KeyAlgorithm: "KEY_ALG_RSA_2048",
					KeyOrigin:    "GOOGLE_PROVIDED",
					KeyType:      "USER_MANAGED",
				},
				{
					Name:         sa.ServiceAccountKeyName("test-project", "test-account", "2"),
					KeyAlgorithm: "KEY_ALG_RSA_2048",
					KeyOrigin:    "GOOGLE_PROVIDED",
					KeyType:      "SYSTEM_MANAGED",
				},
			},
		},
		{
			name:    "List keys with failure",
			project: "test-project",
			account: "test-account",
			fn: func(em *emulator.Emulator) {
				em.SetError(fmt.Errorf("oops"))
			},
			expectErr: "listing service account keys projects/test-project/serviceAccounts/test-account@test-project.iam.gserviceaccount.com: listing service account keys projects/test-project/serviceAccounts/test-account@test-project.iam.gserviceaccount.com: googleapi: got HTTP response code 500 with body: oops\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.New(zerolog.NewConsoleWriter())
			em := emulator.New(log)

			url := em.Run()

			client := sa.NewClient(url, true)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			got, err := client.ListServiceAccountKeys(ctx, sa.ServiceAccountNameFromAccountID(tc.project, tc.account))

			if len(tc.expectErr) > 0 {
				require.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expect, got)
			}
		})
	}
}

func TestClient_AddServiceAccountPolicyBinding(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		project   string
		saEmail   string
		binding   *sa.Binding
		fn        func(*emulator.Emulator)
		expect    any
		expectErr bool
	}{
		{
			name:    "Add valid policy binding, with no existing bindings",
			project: "test-project",
			saEmail: "test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("test-account@test-project.iam.gserviceaccount.com", &iam.Policy{})
			},
			binding: &sa.Binding{
				Role:    "roles/owner",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iam.Policy{
				"test-account@test-project.iam.gserviceaccount.com": {
					Bindings: []*iam.Binding{
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				},
			},
		},
		{
			name:    "Add valid policy binding, with existing bindings",
			project: "test-project",
			saEmail: "test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("test-account@test-project.iam.gserviceaccount.com", &iam.Policy{
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			binding: &sa.Binding{
				Role:    "roles/owner",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iam.Policy{
				"test-account@test-project.iam.gserviceaccount.com": {
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				},
			},
		},
		{
			name:    "Add valid policy binding, with duplicate binding",
			project: "test-project",
			saEmail: "test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("test-account@test-project.iam.gserviceaccount.com", &iam.Policy{
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			binding: &sa.Binding{
				Role:    "roles/editor",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iam.Policy{
				"test-account@test-project.iam.gserviceaccount.com": {
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
					},
				},
			},
		},
		{
			name:    "Add valid policy binding, with duplicat bindings, and other",
			project: "test-project",
			saEmail: "test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("test-account@test-project.iam.gserviceaccount.com", &iam.Policy{
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			binding: &sa.Binding{
				Role:    "roles/owner",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iam.Policy{
				"test-account@test-project.iam.gserviceaccount.com": {
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				},
			},
		},
		{
			name:    "Add valid policy binding, with same role, new member",
			project: "test-project",
			saEmail: "test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("test-account@test-project.iam.gserviceaccount.com", &iam.Policy{
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			binding: &sa.Binding{
				Role:    "roles/owner",
				Members: []string{"user:nais@nav.no"},
			},
			expect: map[string]*iam.Policy{
				"test-account@test-project.iam.gserviceaccount.com": {
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
						{
							Role:    "roles/owner",
							Members: []string{"user:nais@nav.no", "user:nada@nav.no"},
						},
					},
				},
			},
		},
		{
			name:    "Add valid policy binding, with same role, one new member",
			project: "test-project",
			saEmail: "test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("test-account@test-project.iam.gserviceaccount.com", &iam.Policy{
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			binding: &sa.Binding{
				Role:    "roles/owner",
				Members: []string{"user:nada@nav.no", "user:nais@nav.no"},
			},
			expect: map[string]*iam.Policy{
				"test-account@test-project.iam.gserviceaccount.com": {
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no", "user:nais@nav.no"},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.New(zerolog.NewConsoleWriter())
			em := emulator.New(log)

			url := em.Run()

			client := sa.NewClient(url, true)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			err := client.AddServiceAccountPolicyBinding(ctx, tc.project, tc.saEmail, tc.binding)

			if tc.expectErr {
				require.Error(t, err)
				assert.Equal(t, tc.expect, err.Error())
			} else {
				require.NoError(t, err)
				got := em.GetPolicies()

				for _, policy := range em.GetPolicies() {
					for _, binding := range policy.Bindings {
						sort.Strings(binding.Members)
					}
				}

				for _, policy := range tc.expect.(map[string]*iam.Policy) {
					for _, binding := range policy.Bindings {
						sort.Strings(binding.Members)
					}
				}

				assert.Equal(t, tc.expect, got)
			}
		})
	}
}

func TestClient_RemoveServiceAccountPolicyBinding(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		project   string
		saEmail   string
		binding   *sa.Binding
		fn        func(*emulator.Emulator)
		expect    any
		expectErr bool
	}{
		{
			name:    "Remove policy binding, with no existing bindings",
			project: "test-project",
			saEmail: "test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("test-account@test-project.iam.gserviceaccount.com", &iam.Policy{})
			},
			binding: &sa.Binding{
				Role:    "roles/owner",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iam.Policy{
				"test-account@test-project.iam.gserviceaccount.com": {},
			},
		},
		{
			name:    "Remove policy binding, with existing bindings",
			project: "test-project",
			saEmail: "test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("test-account@test-project.iam.gserviceaccount.com", &iam.Policy{
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			binding: &sa.Binding{
				Role:    "roles/editor",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iam.Policy{
				"test-account@test-project.iam.gserviceaccount.com": {},
			},
		},
		{
			name:    "Remove valid policy binding, with other binding",
			project: "test-project",
			saEmail: "test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("test-account@test-project.iam.gserviceaccount.com", &iam.Policy{
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			binding: &sa.Binding{
				Role:    "roles/editor",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iam.Policy{
				"test-account@test-project.iam.gserviceaccount.com": {
					Bindings: []*iam.Binding{
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				},
			},
		},
		{
			name:    "Remove policy binding, only one member",
			project: "test-project",
			saEmail: "test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("test-account@test-project.iam.gserviceaccount.com", &iam.Policy{
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no", "user:nais@nav.no"},
						},
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			binding: &sa.Binding{
				Role:    "roles/editor",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iam.Policy{
				"test-account@test-project.iam.gserviceaccount.com": {
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nais@nav.no"},
						},
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				},
			},
		},
		{
			name:    "Remove policy binding, multiple member",
			project: "test-project",
			saEmail: "test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("test-account@test-project.iam.gserviceaccount.com", &iam.Policy{
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no", "user:nais@nav.no", "user:bob@nav.no"},
						},
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			binding: &sa.Binding{
				Role:    "roles/editor",
				Members: []string{"user:nada@nav.no", "user:nais@nav.no"},
			},
			expect: map[string]*iam.Policy{
				"test-account@test-project.iam.gserviceaccount.com": {
					Bindings: []*iam.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:bob@nav.no"},
						},
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.New(zerolog.NewConsoleWriter())
			em := emulator.New(log)

			url := em.Run()

			client := sa.NewClient(url, true)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			err := client.RemoveServiceAccountPolicyBinding(ctx, tc.project, tc.saEmail, tc.binding)

			if tc.expectErr {
				require.Error(t, err)
				assert.Equal(t, tc.expect, err.Error())
			} else {
				require.NoError(t, err)
				got := em.GetPolicies()
				assert.Equal(t, tc.expect, got)
			}
		})
	}
}
