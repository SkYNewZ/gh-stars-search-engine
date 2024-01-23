//go:build wireinject

package main

import (
	"log/slog"
	"time"

	"github.com/google/wire"
	"github.com/robfig/cron/v3"

	"github.com/SkYNewZ/gh-stars-search-engine/internal/logging"
	"github.com/SkYNewZ/gh-stars-search-engine/internal/slogx"
)

// setupScheduler is used by wire to inject the scheduler.
func setupScheduler(location string, logger *slog.Logger) (*cron.Cron, error) {
	panic(wire.Build(
		provideSchedulerLocation,
		provideSchedulerOptions,
		cron.New,
	))
}

func provideSchedulerLocation(location string) (*time.Location, error) {
	panic(wire.Build(time.LoadLocation))
}

func provideSchedulerOptions(loc *time.Location, logger *slog.Logger) []cron.Option {
	schedulerLogger := logger.With(slogx.Component("scheduler"))
	return []cron.Option{
		cron.WithLocation(loc),
		cron.WithLogger(logging.NewLoggerScheduler(schedulerLogger)),
	}
}
