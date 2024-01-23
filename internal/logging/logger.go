package logging

import (
	"context"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

type loggerContextKey struct{}

// FromContext returns the logger from the context.
// If no logger is found, a new logger is created.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerContextKey{}).(*slog.Logger); ok {
		return logger
	}

	slog.Warn("logger not found in app store, returning default logger")
	return slog.Default()
}

// WithContext returns a new context with the logger.
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

// New returns a new logger.
func New(level slog.Leveler) *slog.Logger {
	slogOpts := &slog.HandlerOptions{
		AddSource:   false,
		Level:       level,
		ReplaceAttr: nil, // use custom levels
	}

	var stdoutHandler slog.Handler = slog.NewJSONHandler(os.Stderr, slogOpts)
	if os.Getenv("SLOG_FORMATTER") == "dev" {
		stdoutHandler = tint.NewHandler(os.Stderr, &tint.Options{
			AddSource:   slogOpts.AddSource,
			Level:       slogOpts.Level,
			ReplaceAttr: slogOpts.ReplaceAttr,
			TimeFormat:  "2006-01-02 15:04:05.000Z",
			NoColor:     false,
		})
	}

	return slog.New(stdoutHandler)
}
