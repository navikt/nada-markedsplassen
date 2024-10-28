package artifactregistry_test

import (
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"cloud.google.com/go/iam/apiv1/iampb"
	"context"
	"github.com/navikt/nada-backend/pkg/artifactregistry"
	"github.com/navikt/nada-backend/pkg/artifactregistry/emulator"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestClient_ListContainerImagesWithTag(t *testing.T) {
	testCases := []struct {
		name      string
		id        *artifactregistry.ContainerRepositoryIdentifier
		tag       string
		fn        func(e *emulator.Emulator)
		expect    any
		expectErr bool
	}{
		{
			name: "success",
			id: &artifactregistry.ContainerRepositoryIdentifier{
				Project:    "test",
				Location:   "eu-north1",
				Repository: "test",
			},
			tag: "latest",
			fn: func(e *emulator.Emulator) {
				e.SetImages(map[string][]*artifactregistrypb.DockerImage{
					"projects/test/locations/eu-north1/repositories/test": {
						{
							Name: "projects/test/locations/eu-north1/repositories/test/dockerImages/nginx@sha256:e9954c1fc875017be1c3e36eca16be2d9e9bccc4bf072163515467d6a823c7cf",
							Uri:  "eu-north1-docker.pkg.dev/test/test/nginx@sha256:e9954c1fc875017be1c3e36eca16be2d9e9bccc4bf072163515467d6a823c7cf",
							Tags: []string{"latest"},
						},
					},
				})
			},
			expect: []*artifactregistry.ContainerImage{
				{
					Name: "projects/test/locations/eu-north1/repositories/test/dockerImages/nginx@sha256:e9954c1fc875017be1c3e36eca16be2d9e9bccc4bf072163515467d6a823c7cf",
					URI:  "eu-north1-docker.pkg.dev/test/test/nginx:latest",
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log := zerolog.New(zerolog.NewConsoleWriter())

			e := emulator.New(log)
			tc.fn(e)
			url := e.Run()

			c := artifactregistry.New(url, true)
			images, err := c.ListContainerImagesWithTag(context.Background(), tc.id, tc.tag)

			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expect, images)
		})
	}
}

func TestClient_AddArtifactRegistryPolicyBinding(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		id        *artifactregistry.ContainerRepositoryIdentifier
		binding   *artifactregistry.Binding
		fn        func(*emulator.Emulator)
		expect    any
		expectErr bool
	}{
		{
			name: "Add valid policy binding, with no existing bindings",
			id: &artifactregistry.ContainerRepositoryIdentifier{
				Project:    "test",
				Location:   "eu-north1",
				Repository: "knast-images",
			},
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("knast-images", &iampb.Policy{})
			},
			binding: &artifactregistry.Binding{
				Role:    "roles/owner",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iampb.Policy{
				"knast-images": {
					Bindings: []*iampb.Binding{
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				},
			},
		},
		{
			name: "Add valid policy binding, with existing bindings",
			id: &artifactregistry.ContainerRepositoryIdentifier{
				Project:    "test",
				Location:   "eu-north1",
				Repository: "knast-images",
			},
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("knast-images", &iampb.Policy{
					Bindings: []*iampb.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			binding: &artifactregistry.Binding{
				Role:    "roles/owner",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iampb.Policy{
				"knast-images": {
					Bindings: []*iampb.Binding{
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
			name: "Add valid policy binding, with duplicate binding",
			id: &artifactregistry.ContainerRepositoryIdentifier{
				Project:    "test",
				Location:   "eu-north1",
				Repository: "knast-images",
			},
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("knast-images", &iampb.Policy{
					Bindings: []*iampb.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			binding: &artifactregistry.Binding{
				Role:    "roles/editor",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iampb.Policy{
				"knast-images": {
					Bindings: []*iampb.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
					},
				},
			},
		},
		{
			name: "Add valid policy binding, with duplicat bindings, and other",
			id: &artifactregistry.ContainerRepositoryIdentifier{
				Project:    "test",
				Location:   "eu-north1",
				Repository: "knast-images",
			},
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("knast-images", &iampb.Policy{
					Bindings: []*iampb.Binding{
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
			binding: &artifactregistry.Binding{
				Role:    "roles/owner",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iampb.Policy{
				"knast-images": {
					Bindings: []*iampb.Binding{
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
			name: "Add valid policy binding, with same role, new member",
			id: &artifactregistry.ContainerRepositoryIdentifier{
				Project:    "test",
				Location:   "eu-north1",
				Repository: "knast-images",
			},
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("knast-images", &iampb.Policy{
					Bindings: []*iampb.Binding{
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
			binding: &artifactregistry.Binding{
				Role:    "roles/owner",
				Members: []string{"user:nais@nav.no"},
			},
			expect: map[string]*iampb.Policy{
				"knast-images": {
					Bindings: []*iampb.Binding{
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
			name: "Add valid policy binding, with same role, one new member",
			id: &artifactregistry.ContainerRepositoryIdentifier{
				Project:    "test",
				Location:   "eu-north1",
				Repository: "knast-images",
			},
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("knast-images", &iampb.Policy{
					Bindings: []*iampb.Binding{
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
			binding: &artifactregistry.Binding{
				Role:    "roles/owner",
				Members: []string{"user:nada@nav.no", "user:nais@nav.no"},
			},
			expect: map[string]*iampb.Policy{
				"knast-images": {
					Bindings: []*iampb.Binding{
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

			client := artifactregistry.New(url, true)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			err := client.AddArtifactRegistryPolicyBinding(ctx, tc.id, tc.binding)

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

func TestClient_RemoveServiceAccountPolicyBinding(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		id        *artifactregistry.ContainerRepositoryIdentifier
		binding   *artifactregistry.Binding
		fn        func(*emulator.Emulator)
		expect    any
		expectErr bool
	}{
		{
			name: "Remove policy binding, with no existing bindings",
			id: &artifactregistry.ContainerRepositoryIdentifier{
				Project:    "test",
				Location:   "eu-north1",
				Repository: "knast-images",
			},
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("knast-images", &iampb.Policy{})
			},
			binding: &artifactregistry.Binding{
				Role:    "roles/owner",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iampb.Policy{
				"knast-images": {},
			},
		},
		{
			name: "Remove policy binding, with existing bindings",
			id: &artifactregistry.ContainerRepositoryIdentifier{
				Project:    "test",
				Location:   "eu-north1",
				Repository: "knast-images",
			},
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("knast-images", &iampb.Policy{
					Bindings: []*iampb.Binding{
						{
							Role:    "roles/editor",
							Members: []string{"user:nada@nav.no"},
						},
					},
				})
			},
			binding: &artifactregistry.Binding{
				Role:    "roles/editor",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iampb.Policy{
				"knast-images": {},
			},
		},
		{
			name: "Remove valid policy binding, with other binding",
			id: &artifactregistry.ContainerRepositoryIdentifier{
				Project:    "test",
				Location:   "eu-north1",
				Repository: "knast-images",
			},
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("knast-images", &iampb.Policy{
					Bindings: []*iampb.Binding{
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
			binding: &artifactregistry.Binding{
				Role:    "roles/editor",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iampb.Policy{
				"knast-images": {
					Bindings: []*iampb.Binding{
						{
							Role:    "roles/owner",
							Members: []string{"user:nada@nav.no"},
						},
					},
				},
			},
		},
		{
			name: "Remove policy binding, only one member",
			id: &artifactregistry.ContainerRepositoryIdentifier{
				Project:    "test",
				Location:   "eu-north1",
				Repository: "knast-images",
			},
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("knast-images", &iampb.Policy{
					Bindings: []*iampb.Binding{
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
			binding: &artifactregistry.Binding{
				Role:    "roles/editor",
				Members: []string{"user:nada@nav.no"},
			},
			expect: map[string]*iampb.Policy{
				"knast-images": {
					Bindings: []*iampb.Binding{
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
			name: "Remove policy binding, multiple member",
			id: &artifactregistry.ContainerRepositoryIdentifier{
				Project:    "test",
				Location:   "eu-north1",
				Repository: "knast-images",
			},
			fn: func(em *emulator.Emulator) {
				em.SetIamPolicy("knast-images", &iampb.Policy{
					Bindings: []*iampb.Binding{
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
			binding: &artifactregistry.Binding{
				Role:    "roles/editor",
				Members: []string{"user:nada@nav.no", "user:nais@nav.no"},
			},
			expect: map[string]*iampb.Policy{
				"knast-images": {
					Bindings: []*iampb.Binding{
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

			client := artifactregistry.New(url, true)

			ctx := context.Background()

			if tc.fn != nil {
				tc.fn(em)
			}

			err := client.RemoveArtifactRegistryPolicyBinding(ctx, tc.id, tc.binding)

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
