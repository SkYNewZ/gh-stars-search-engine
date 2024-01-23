package logging

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/motemen/go-loghttp"
	"github.com/motemen/go-nuts/roundtime"

	"github.com/SkYNewZ/gh-stars-search-engine/internal/slogx"
)

// NewLoggerTransport is used to create a new logger transport.
func NewLoggerTransport(logger *slog.Logger) http.RoundTripper {
	return &loghttp.Transport{
		Transport:   http.DefaultTransport, // use it to get the default values
		LogRequest:  logRequest(logger),
		LogResponse: logResponse(logger),
	}
}

// logRequest is used to log the request.
func logRequest(logger *slog.Logger) func(req *http.Request) {
	return func(req *http.Request) {
		logger.
			With(slog.String("method", req.Method)).
			With(slog.String("url", req.URL.String())).
			Debug("URL being requested")
	}
}

func logResponse(logger *slog.Logger) func(resp *http.Response) {
	return func(resp *http.Response) {
		ctx := resp.Request.Context()
		l := logger.
			With(slog.Int("status", resp.StatusCode)).
			With(slog.String("url", resp.Request.URL.String()))

		if start, ok := ctx.Value(loghttp.ContextKeyRequestStart).(time.Time); ok {
			l.With(slogx.Duration(roundtime.Duration(time.Since(start), 2))).Debug("Got response")
			return
		}

		l.Debug("Got response")
	}
}
