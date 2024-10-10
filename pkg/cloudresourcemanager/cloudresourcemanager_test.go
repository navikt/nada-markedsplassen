package cloudresourcemanager_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	crm "github.com/navikt/nada-backend/pkg/cloudresourcemanager"
	"github.com/navikt/nada-backend/pkg/cloudresourcemanager/emulator"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/cloudresourcemanager/v3"
)

func TestClient_AddProjectServiceAccountPolicyBinding(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		project   string
		binding   *crm.Binding
		fn        func(*emulator.Emulator)
		expectErr string
		expect    *cloudresourcemanager.Policy
	}{
		{
			name:    "Add valid policy binding",
			project: "test-project",
			binding: &crm.Binding{
				Role:    "roles/editor",
				Members: []string{"serviceAccount:test-account@test-project.iam.gserviceaccount.com"},
			},
			fn: func(em *emulator.Emulator) {
				em.SetPolicy("test-project", &cloudresourcemanager.Policy{
					Bindings: []*cloudresourcemanager.Binding{
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			expect: &cloudresourcemanager.Policy{
				Bindings: []*cloudresourcemanager.Binding{
					{
						Role: "roles/owner",
						Members: []string{
							"user:nada@nav.no",
						},
					},
					{
						Role: "roles/editor",
						Members: []string{
							"serviceAccount:test-account@test-project.iam.gserviceaccount.com",
						},
					},
				},
			},
		},
		{
			name:    "Add policy binding with failure",
			project: "test-project",
			binding: &crm.Binding{
				Role:    "roles/editor",
				Members: []string{"serviceAccount:test-account@test-project.iam.gserviceaccount.com"},
			},
			expectErr: "getting project test-project policy: googleapi: got HTTP response code 500 with body: oops\n",
			fn: func(em *emulator.Emulator) {
				em.SetError(fmt.Errorf("oops"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.New(zerolog.NewConsoleWriter())
			em := emulator.New(log)

			url := em.Run()

			client := crm.NewClient(url, true, nil)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			err := client.AddProjectIAMPolicyBinding(ctx, tc.project, tc.binding)

			if len(tc.expectErr) > 0 {
				require.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expect, em.GetPolicy(tc.project))
			}
		})
	}
}

func TestClient_ListProjectServiceAccountPolicyBindings(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		project   string
		member    string
		fn        func(*emulator.Emulator)
		expect    []*crm.Binding
		expectErr string
	}{
		{
			name:      "List bindings with no bindings",
			project:   "test-project",
			expectErr: "project test-project: not found",
		},
		{
			name:    "List bindings with failure",
			project: "test-project",
			member:  "serviceAccount:test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetError(fmt.Errorf("oops"))
			},
			expectErr: "getting project test-project policy: googleapi: got HTTP response code 500 with body: oops\n",
		},
		{
			name:    "List bindings with no matching bindings",
			project: "test-project",
			member:  "serviceAccount:test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetPolicy("test-project", &cloudresourcemanager.Policy{
					Bindings: []*cloudresourcemanager.Binding{
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			expect: []*crm.Binding(nil),
		},
		{
			name:    "List valid bindings",
			project: "test-project",
			member:  "serviceAccount:test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetPolicy("test-project", &cloudresourcemanager.Policy{
					Bindings: []*cloudresourcemanager.Binding{
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
						{
							Role:    "roles/editor",
							Members: []string{"serviceAccount:test-account@test-project.iam.gserviceaccount.com"},
						},
					},
				})
			},
			expect: []*crm.Binding{
				{
					Role:    "roles/editor",
					Members: []string{"serviceAccount:test-account@test-project.iam.gserviceaccount.com"},
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

			client := crm.NewClient(url, true, nil)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			got, err := client.ListProjectIAMPolicyBindings(ctx, tc.project, tc.member)

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

func TestClient_RemoveProjectServiceAccountPolicyBinding(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		project   string
		member    string
		fn        func(*emulator.Emulator)
		expect    *cloudresourcemanager.Policy
		expectErr string
	}{
		{
			name:    "Remove valid binding",
			project: "test-project",
			member:  "serviceAccount:test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetPolicy("test-project", &cloudresourcemanager.Policy{
					Bindings: []*cloudresourcemanager.Binding{
						{
							Role: "roles/owner",
							Members: []string{
								"user:nada@nav.no",
								"serviceAccount:test-account@test-project.iam.gserviceaccount.com",
							},
						},
						{
							Role:    "roles/editor",
							Members: []string{"serviceAccount:test-account@test-project.iam.gserviceaccount.com"},
						},
					},
				})
			},
			expect: &cloudresourcemanager.Policy{
				Bindings: []*cloudresourcemanager.Binding{
					{
						Role:    "roles/owner",
						Members: []string{"user:nada@nav.no"},
					},
				},
			},
		},
		{
			name:    "Remove binding with failure",
			project: "test-project",
			member:  "serviceAccount:test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetError(fmt.Errorf("oops"))
			},
			expectErr: "getting project test-project policy: googleapi: got HTTP response code 500 with body: oops\n",
		},
		{
			name:      "Remove binding with missing project",
			project:   "test-project",
			member:    "serviceAccount:test-account@test-project.iam.gserviceaccount.com",
			expectErr: "project test-project: not found",
		},
		{
			name:    "Remove binding with missing binding",
			project: "test-project",
			member:  "serviceAccount:test-account@test-project.iam.gserviceaccount.com",
			fn: func(em *emulator.Emulator) {
				em.SetPolicy("test-project", &cloudresourcemanager.Policy{
					Bindings: []*cloudresourcemanager.Binding{
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			expect: &cloudresourcemanager.Policy{
				Bindings: []*cloudresourcemanager.Binding{
					{
						Role:    "roles/owner",
						Members: []string{"user:nada@nav.no"},
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

			client := crm.NewClient(url, true, nil)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			err := client.RemoveProjectIAMPolicyBindingMember(ctx, tc.project, tc.member)

			if len(tc.expectErr) > 0 {
				require.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expect, em.GetPolicy(tc.project))
			}
		})
	}
}

func TestClient_UpdateProjectPolicyBindingsMembers(t *testing.T) {
	t.Parallel()

	log := zerolog.New(zerolog.NewConsoleWriter())

	testCases := []struct {
		name      string
		project   string
		callback  func(string, []string) []string
		fn        func(*emulator.Emulator)
		expect    *cloudresourcemanager.Policy
		expectErr string
	}{
		{
			name:    "Remove deleted member with role",
			project: "test-project",
			fn: func(em *emulator.Emulator) {
				em.SetPolicy("test-project", &cloudresourcemanager.Policy{
					Bindings: []*cloudresourcemanager.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"deleted:user:nada@nav.no"},
						},
					},
				})
			},
			callback: crm.RemoveDeletedMembersWithRole([]string{"roles/editor"}, log),
			expect: &cloudresourcemanager.Policy{
				Bindings: []*cloudresourcemanager.Binding{
					{
						Role: "roles/editor",
					},
				},
			},
		},
		{
			name:    "Only remove deleted members with role",
			project: "test-project",
			fn: func(em *emulator.Emulator) {
				em.SetPolicy("test-project", &cloudresourcemanager.Policy{
					Bindings: []*cloudresourcemanager.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"deleted:user:nada@nav.no"},
						},
						{
							Role:    "roles/owner",
							Members: []string{"deleted:serviceAccount:something"},
						},
						{
							Role:    "roles/viewer",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			callback: crm.RemoveDeletedMembersWithRole([]string{"roles/editor", "roles/owner", "roles/viewer"}, log),
			expect: &cloudresourcemanager.Policy{
				Bindings: []*cloudresourcemanager.Binding{
					{
						Role: "roles/editor",
					},
					{
						Role: "roles/owner",
					},
					{
						Role:    "roles/viewer",
						Members: []string{"user:nada@nav.no"},
					},
				},
			},
		},
		{
			name:    "Only evaluate the provided roles",
			project: "test-project",
			fn: func(em *emulator.Emulator) {
				em.SetPolicy("test-project", &cloudresourcemanager.Policy{
					Bindings: []*cloudresourcemanager.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"deleted:user:nada@nav.no"},
						},
						{
							Role:    "roles/owner",
							Members: []string{"deleted:serviceAccount:something"},
						},
						{
							Role:    "roles/viewer",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			callback: crm.RemoveDeletedMembersWithRole([]string{"roles/editor", "roles/viewer"}, log),
			expect: &cloudresourcemanager.Policy{
				Bindings: []*cloudresourcemanager.Binding{
					{
						Role: "roles/editor",
					},
					{
						Role:    "roles/owner",
						Members: []string{"deleted:serviceAccount:something"},
					},
					{
						Role:    "roles/viewer",
						Members: []string{"user:nada@nav.no"},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			em := emulator.New(log)

			url := em.Run()

			client := crm.NewClient(url, true, nil)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			err := client.UpdateProjectIAMPolicyBindingsMembers(ctx, tc.project, tc.callback)

			if len(tc.expectErr) > 0 {
				require.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expect, em.GetPolicy(tc.project))
			}
		})
	}
}

func TestClient_RemoveProjectIAMPolicyBindingMemberForRole(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		project   string
		member    string
		role      string
		fn        func(*emulator.Emulator)
		expect    *cloudresourcemanager.Policy
		expectErr string
	}{
		{
			name:    "Remove valid binding",
			project: "test-project",
			member:  "serviceAccount:test-account@test-project.iam.gserviceaccount.com",
			role:    "roles/owner",
			fn: func(em *emulator.Emulator) {
				em.SetPolicy("test-project", &cloudresourcemanager.Policy{
					Bindings: []*cloudresourcemanager.Binding{
						{
							Role: "roles/owner",
							Members: []string{
								"user:nada@nav.no",
								"serviceAccount:test-account@test-project.iam.gserviceaccount.com",
							},
						},
						{
							Role:    "roles/editor",
							Members: []string{"serviceAccount:test-account@test-project.iam.gserviceaccount.com"},
						},
					},
				})
			},
			expect: &cloudresourcemanager.Policy{
				Bindings: []*cloudresourcemanager.Binding{
					{
						Role:    "roles/owner",
						Members: []string{"user:nada@nav.no"},
					},
					{
						Role:    "roles/editor",
						Members: []string{"serviceAccount:test-account@test-project.iam.gserviceaccount.com"},
					},
				},
			},
		},
		{
			name:    "Remove binding with failure",
			project: "test-project",
			member:  "serviceAccount:test-account@test-project.iam.gserviceaccount.com",
			role:    "roles/editor",
			fn: func(em *emulator.Emulator) {
				em.SetError(fmt.Errorf("oops"))
			},
			expectErr: "getting project test-project policy: googleapi: got HTTP response code 500 with body: oops\n",
		},
		{
			name:      "Remove binding with missing project",
			project:   "test-project",
			member:    "serviceAccount:test-account@test-project.iam.gserviceaccount.com",
			role:      "roles/editor",
			expectErr: "project test-project: not found",
		},
		{
			name:    "Remove binding with missing binding role",
			project: "test-project",
			member:  "serviceAccount:test-account@test-project.iam.gserviceaccount.com",
			role:    "roles/editor",
			fn: func(em *emulator.Emulator) {
				em.SetPolicy("test-project", &cloudresourcemanager.Policy{
					Bindings: []*cloudresourcemanager.Binding{
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			expect: &cloudresourcemanager.Policy{
				Bindings: []*cloudresourcemanager.Binding{
					{
						Role:    "roles/owner",
						Members: []string{"user:nada@nav.no"},
					},
				},
			},
		},
		{
			name:    "Remove binding with missing binding member",
			project: "test-project",
			member:  "serviceAccount:test-account@test-project.iam.gserviceaccount.com",
			role:    "roles/editor",
			fn: func(em *emulator.Emulator) {
				em.SetPolicy("test-project", &cloudresourcemanager.Policy{
					Bindings: []*cloudresourcemanager.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			expect: &cloudresourcemanager.Policy{
				Bindings: []*cloudresourcemanager.Binding{
					{
						Role:    "roles/editor",
						Members: []string{"user:nada@nav.no"},
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

			client := crm.NewClient(url, true, nil)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			err := client.RemoveProjectIAMPolicyBindingMemberForRole(ctx, tc.project, tc.role, tc.member)

			if len(tc.expectErr) > 0 {
				require.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expect, em.GetPolicy(tc.project))
			}
		})
	}
}

func TestClient_CreateZonalTagBinding(t *testing.T) {
	tagBinding := &cloudresourcemanager.TagBinding{
		Parent:                 "//resource.my.vm",
		TagValueNamespacedName: "test-project/test-tag-key/test-tag-value",
	}

	expect, err := json.Marshal(tagBinding)
	require.NoError(t, err)

	matcher := func(req *http.Request) bool {
		got, err := io.ReadAll(req.Body)
		if err != nil {
			return false
		}

		require.Equal(t, expect, got)

		return true
	}

	client := &http.Client{Transport: &http.Transport{TLSHandshakeTimeout: 60 * time.Second}}
	httpmock.ActivateNonDefault(client)
	t.Cleanup(httpmock.DeactivateAndReset)

	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://europe-north1-a-cloudresourcemanager.googleapis.com/v3/tagBindings",
		httpmock.NewMatcher("check_content", matcher),
		httpmock.NewStringResponder(http.StatusOK, ""),
	)

	err = crm.NewClient("", false, client).CreateZonalTagBinding(context.Background(), "europe-north1-a", tagBinding.Parent, tagBinding.TagValueNamespacedName)
	require.NoError(t, err)
}
