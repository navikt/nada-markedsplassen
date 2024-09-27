package securewebproxy_test

import (
	"context"
	"testing"

	"github.com/navikt/nada-backend/pkg/securewebproxy"
	"github.com/navikt/nada-backend/pkg/securewebproxy/emulator"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	log := zerolog.New(zerolog.NewConsoleWriter())
	ctx := context.Background()

	e := emulator.New(log)
	url := e.Run()

	client := securewebproxy.New(url, true)

	id := &securewebproxy.URLListIdentifier{
		Project:  "test",
		Location: "europe-north1",
		Slug:     "myurllist",
	}

	urls := []string{"example.com", "github.com/navikt"}

	t.Run("Get URL list that does not exist", func(t *testing.T) {
		_, err := client.GetURLList(ctx, id)
		require.Error(t, err)
		assert.Equal(t, securewebproxy.ErrNotExist, err)
	})

	t.Run("Create URL list", func(t *testing.T) {
		err := client.CreateURLList(ctx, &securewebproxy.URLListCreateOpts{
			ID:          id,
			Description: "My URL list",
			URLS:        urls,
		})
		require.NoError(t, err)
	})

	t.Run("Create URL list that already exists", func(t *testing.T) {
		err := client.CreateURLList(ctx, &securewebproxy.URLListCreateOpts{
			ID:          id,
			Description: "My URL list",
			URLS:        urls,
		})
		require.Error(t, err)
		assert.Equal(t, securewebproxy.ErrExist, err)
	})

	t.Run("Get URL list that exists", func(t *testing.T) {
		got, err := client.GetURLList(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, urls, got)
	})

	t.Run("Delete URL list", func(t *testing.T) {
		err := client.DeleteURLList(ctx, id)
		require.NoError(t, err)
	})

	t.Run("Delete URL list that does not exist", func(t *testing.T) {
		err := client.DeleteURLList(ctx, id)
		require.NoError(t, err)
	})
}
