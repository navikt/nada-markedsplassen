package worker

import (
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/worker/worker_args"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"
	"log/slog"
	"riverqueue.com/riverpro"
	"riverqueue.com/riverpro/driver/riverpropgxv5"
	"time"
)

func RiverConfig(log *zerolog.Logger, workers *river.Workers) *riverpro.Config {
	var level slog.Level

	switch log.GetLevel() {
	case zerolog.DebugLevel:
		level = slog.LevelDebug
	case zerolog.InfoLevel:
		level = slog.LevelInfo
	case zerolog.WarnLevel:
		level = slog.LevelWarn
	case zerolog.ErrorLevel:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slogzerolog.Option{Level: level, Logger: log}.NewZerologHandler())

	return &riverpro.Config{
		Config: river.Config{
			Queues: map[string]river.QueueConfig{
				worker_args.WorkstationQueue: {
					MaxWorkers: 10,
				},
				worker_args.WorkstationConnectivityQueue: {
					MaxWorkers: 10,
				},
				worker_args.MetabaseQueue: {
					MaxWorkers: 10,
				},
			},
			Logger:     logger,
			Workers:    workers,
			JobTimeout: 10 * time.Minute,
		},
	}
}

func RiverClient(config *riverpro.Config, repo *database.Repo) (*riverpro.Client[pgx.Tx], error) {
	client, err := riverpro.NewClient[pgx.Tx](riverpropgxv5.New(repo.GetDBX()), config)
	if err != nil {
		return nil, fmt.Errorf("creating river client: %w", err)
	}

	return client, nil
}
