package urllists

import (
	"context"
	"fmt"

	"github.com/navikt/nada-backend/pkg/syncers"
	"github.com/rs/zerolog"

	"github.com/navikt/nada-backend/pkg/service"
)

var _ syncers.Runner = &Runner{}

type Runner struct {
	api service.WorkstationsService
}

func New(api service.WorkstationsService) *Runner {
	return &Runner{
		api: api,
	}
}

func (r *Runner) Name() string {
	return "WorkstationURLListsAccessEnsurer"
}

func (r *Runner) RunOnce(ctx context.Context, log zerolog.Logger) error {
	activeURLLists, err := r.api.GetWorkstationActiveURLListsForAll(ctx)
	if err != nil {
		return fmt.Errorf("getting active urlists: %w", err)
	}

	for _, urlList := range activeURLLists {
		err := r.api.EnsureWorkstationURLList(ctx, urlList)
		if err != nil {
			return fmt.Errorf("ensuring url list %s for user %s: %w", urlList.URLList, urlList.Slug, err)
		}
	}

	return nil
}
