package syncers

import (
	"context"
	"github.com/navikt/nada-backend/pkg/leaderelection"
	"github.com/rs/zerolog"
	"time"
)

type Runner interface {
	Name() string
	RunOnce(ctx context.Context) error
}

type Syncer struct {
	runner Runner

	runAtStart      bool
	onlyRunIfLeader bool

	initialDelaySec time.Duration
	runIntervalSec  time.Duration

	log zerolog.Logger
}

type Option func(*Syncer)

func WithRunAtStart() Option {
	return func(s *Syncer) {
		s.runAtStart = true
	}
}

func WithInitialDelaySec(initialDelaySec int) Option {
	return func(s *Syncer) {
		s.initialDelaySec = time.Duration(initialDelaySec) * time.Second
	}
}

func WithOnlyRunIfLeader() Option {
	return func(s *Syncer) {
		s.onlyRunIfLeader = true
	}
}

func (s *Syncer) Run(ctx context.Context) {
	s.log.Info().Fields(map[string]interface{}{
		"only_run_if_leader": s.onlyRunIfLeader,
		"run_at_start":       s.runAtStart,
		"initial_delay_sec":  s.initialDelaySec.String(),
		"run_interval_sec":   s.runIntervalSec.String(),
	}).Msg("starting syncer")

	if s.initialDelaySec > 0 {
		s.log.Info().Msg("sleeping before starting")

		time.Sleep(s.initialDelaySec)
	}

	if s.runAtStart && s.shouldRun() {
		s.log.Info().Msg("running initial sync")

		err := s.runner.RunOnce(ctx)
		if err != nil {
			s.log.Error().Err(err).Msg("running initial sync")
		}
	}

	ticker := time.NewTicker(s.runIntervalSec)

	for {
		select {
		case <-ctx.Done():
			s.log.Info().Msg("context done, stopping syncer")
			return
		case <-ticker.C:
			if !s.shouldRun() {
				continue
			}
			s.log.Info().Msg("running sync")

			err := s.runner.RunOnce(ctx)
			if err != nil {
				s.log.Error().Err(err).Msg("running sync")
			}
		}
	}
}

func (s *Syncer) shouldRun() bool {
	if !s.onlyRunIfLeader {
		return true
	}

	isLeader, err := leaderelection.IsLeader()
	if err != nil {
		s.log.Error().Err(err).Msg("checking leader status")

		return false
	}

	return isLeader
}

func New(runIntervalSec int, runner Runner, log zerolog.Logger, options ...Option) *Syncer {
	s := &Syncer{
		runner:         runner,
		runIntervalSec: time.Duration(runIntervalSec) * time.Second,
		log:            log.With().Str("syncer", runner.Name()).Logger(),
	}

	for _, option := range options {
		option(s)
	}

	return s
}
