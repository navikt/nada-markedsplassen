package artifactregistry_test

import (
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"context"
	"github.com/navikt/nada-backend/pkg/artifactregistry"
	"github.com/navikt/nada-backend/pkg/artifactregistry/emulator"
	"github.com/rs/zerolog"
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
