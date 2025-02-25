package metabase_users

import (
	"context"
	"fmt"
	"github.com/navikt/nada-backend/pkg/syncers"
	"github.com/rs/zerolog"
	"strconv"
	"time"

	"github.com/navikt/nada-backend/pkg/service"
)

var _ syncers.Runner = &Runner{}

type Runner struct {
	api service.MetabaseAPI
}

func New(api service.MetabaseAPI) *Runner {
	return &Runner{
		api: api,
	}
}

func (r *Runner) Name() string {
	return "MetabaseDeactivateUsersWithNoActivity"
}

func (r *Runner) RunOnce(ctx context.Context, log zerolog.Logger) error {
	users, err := r.api.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("getting users: %w", err)
	}

	for _, user := range users {
		if user.LastLogin == nil {
			log.Info().Str("id", strconv.Itoa(user.ID)).Str("email", user.Email).Msg("User with no activity found")
			continue
		}

		if user.LastLogin.Before(time.Now().AddDate(-1, 0, 0)) {
			log.Info().Str("id", strconv.Itoa(user.ID)).Str("last_login", user.LastLogin.String()).Str("email", user.Email).Msg("User with no activity for over a year found")
		}
	}

	return nil
}
