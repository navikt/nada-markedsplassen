package syncers

import (
	"context"
	"time"

	"github.com/navikt/nada-backend/pkg/errs"

	"github.com/navikt/nada-backend/pkg/leaderelection"
	"github.com/rs/zerolog"
)

const (
	DefaultDelaySec = 60
)

type Runner interface {
	Name() string
	RunOnce(ctx context.Context, log zerolog.Logger) error
}

type Syncer struct {
	runner Runner

	runAtStart      bool
	onlyRunIfLeader bool

	initialDelay time.Duration
	runInterval  time.Duration

	log zerolog.Logger
}

type Option func(*Syncer)

func DefaultOptions() []Option {
	return []Option{
		WithRunAtStart(),
		WithInitialDelaySec(DefaultDelaySec),
		WithOnlyRunIfLeader(),
	}
}

func WithRunAtStart() Option {
	return func(s *Syncer) {
		s.runAtStart = true
	}
}

func WithInitialDelaySec(initialDelaySec int) Option {
	return func(s *Syncer) {
		s.initialDelay = time.Duration(initialDelaySec) * time.Second
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
		"initial_delay":      s.initialDelay.String(),
		"run_interval":       s.runInterval.String(),
	}).Msg("starting syncer")

	if s.initialDelay > 0 {
		s.log.Info().Msg("sleeping before starting")

		time.Sleep(s.initialDelay)
	}

	if s.runAtStart && s.shouldRun() {
		s.log.Info().Msg("running initial sync")

		err := s.runner.RunOnce(ctx, s.log)
		if err != nil {
			s.log.Error().Err(err).Strs("stack", errs.OpStack(err)).Msg("running initial sync")
		}
	}

	ticker := time.NewTicker(s.runInterval)

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

			err := s.runner.RunOnce(ctx, s.log)
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
		s.log.Error().Err(err).Strs("stack", errs.OpStack(err)).Msg("checking leader status")

		return false
	}

	return isLeader
}

func New(runIntervalSec int, runner Runner, log zerolog.Logger, options ...Option) *Syncer {
	s := &Syncer{
		runner:      runner,
		runInterval: time.Duration(runIntervalSec) * time.Second,
		log:         log.With().Str("syncer", runner.Name()).Logger(),
	}

	for _, option := range options {
		option(s)
	}

	return s
}
