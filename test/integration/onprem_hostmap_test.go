package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/routes"
	"github.com/navikt/nada-backend/pkg/service/core/storage/postgres"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// nolint: tparallel
func TestOnpremHostmap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	log := zerolog.New(os.Stdout)

	c := NewContainers(t, log)
	defer c.Cleanup()

	zlog := zerolog.New(os.Stdout)
	r := TestRouter(zlog)

	core.NewOnpremMappingService(
		ctx,
		"",
		cloudStorageAPI service.CloudStorageAPI,
		dvhAPI service.DatavarehusAPI
	)

	{
		store := postgres.NewKeywordsStorage(repo)
		s := core.NewKeywordsService(store, "nada@nav.no")
		h := handlers.NewKeywordsHandler(s)
		e := routes.NewKeywordEndpoints(zlog, h)
		f := routes.NewKeywordRoutes(e, injectUser(&service.User{
			Email: "bob@example.com",
			GoogleGroups: []service.Group{
				{
					Name:  "nada",
					Email: "nada@nav.no",
				},
			},
		}))
		f(r)
	}

	server := httptest.NewServer(r)
	defer server.Close()

	t.Run("Update keywords", func(t *testing.T) {
		NewTester(t, server).
			Post(ctx, &service.UpdateKeywordsDto{
				ObsoleteKeywords: []string{"keyword1"},
				ReplacedKeywords: []string{"keyword2"},
				NewText:          []string{"keyword2_replaced"},
			}, "/api/keywords").
			HasStatusCode(http.StatusNoContent)
	})

	t.Run("Get keywords", func(t *testing.T) {
		got := &service.KeywordsList{}

		expect := &service.KeywordsList{
			KeywordItems: []service.KeywordItem{
				{
					Keyword: "keyword3",
					Count:   1,
				},
				{
					Keyword: "keyword2_replaced",
					Count:   1,
				},
			},
		}

		NewTester(t, server).
			Get(ctx, "/api/keywords").
			HasStatusCode(http.StatusOK).
			Expect(expect, got)
	})
}
