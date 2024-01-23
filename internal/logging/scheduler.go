package logging

import (
	"log/slog"

	"github.com/robfig/cron/v3"

	"github.com/SkYNewZ/gh-stars-search-engine/internal/slogx"
)

// loggerScheduler implements cron.Logger.
type loggerScheduler struct {
	logger *slog.Logger
}

// NewLoggerScheduler returns a *slog.Logger that implements cron.Logger.
func NewLoggerScheduler(logger *slog.Logger) cron.Logger {
	return &loggerScheduler{logger: logger}
}

func (l *loggerScheduler) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, keysAndValues...)
}

func (l *loggerScheduler) Error(err error, msg string, keysAndValues ...interface{}) {
	l.logger.With(slogx.Err(err)).Error(msg, keysAndValues...)
}
