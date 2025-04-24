package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/transport"
	"github.com/rs/zerolog"
)

type StoryEndpoints struct {
	GetObject          http.HandlerFunc
	GetIndex           http.HandlerFunc
	CreateStoryForTeam http.HandlerFunc
	RecreateStoryFiles http.HandlerFunc
	AppendStoryFiles   http.HandlerFunc
	GetStory           http.HandlerFunc
	CreateStory        http.HandlerFunc
	UpdateStory        http.HandlerFunc
	DeleteStory        http.HandlerFunc
}

func NewStoryEndpoints(log zerolog.Logger, h *handlers.StoryHandler) *StoryEndpoints {
	return &StoryEndpoints{
		GetObject:          transport.For(h.GetObject).Build(log),
		GetIndex:           transport.For(h.GetIndex).Build(log),
		CreateStoryForTeam: transport.For(h.CreateStoryForTeam).RequestFromJSON().Build(log),
		RecreateStoryFiles: transport.For(h.RecreateStoryFiles).Build(log),
		AppendStoryFiles:   transport.For(h.AppendStoryFiles).Build(log),
		GetStory:           transport.For(h.GetStory).Build(log),
		CreateStory:        transport.For(h.CreateStory).Build(log),
		UpdateStory:        transport.For(h.UpdateStory).RequestFromJSON().Build(log),
		DeleteStory:        transport.For(h.DeleteStory).Build(log),
	}
}

func stripTrailingSlash(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > 1 && r.URL.Path[len(r.URL.Path)-1] == '/' {
			r.URL.Path = r.URL.Path[:len(r.URL.Path)-1]
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func NewStoryRoutes(
	endpoints *StoryEndpoints,
	auth func(http.Handler) http.Handler,
	nadaToken func(http.Handler) http.Handler,
) AddRoutesFn {
	return func(router chi.Router) {
		router.Route(`/quarto`, func(r chi.Router) {
			r.Route("/{id}", func(r chi.Router) {
				r.With(stripTrailingSlash).Get("/", endpoints.GetIndex)
				r.Get("/*", endpoints.GetObject)
			})

			// Endpoints used programmatically, which rely on the Nada team token
			r.With(nadaToken).Post("/create", endpoints.CreateStoryForTeam)
			r.With(nadaToken).Put("/update/{id}", endpoints.RecreateStoryFiles)
			r.With(nadaToken).Patch("/update/{id}", endpoints.AppendStoryFiles)
		})

		// Endpoints used primarily by the UI for updating stories
		// and uploading files when creating
		router.Route("/api/stories", func(r chi.Router) {
			r.Use(auth)
			r.Get("/{id}", endpoints.GetStory)
			r.Post("/new", endpoints.CreateStory)
			r.Put("/{id}", endpoints.UpdateStory)
			r.Delete("/{id}", endpoints.DeleteStory)
		})
	}
}
