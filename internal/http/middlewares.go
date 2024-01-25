package http

import (
	"log/slog"
	"net/http"

	"github.com/SkYNewZ/gh-stars-search-engine/internal/slogx"
)

func (s *server) recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.With(slogx.Err(err.(error))).Error("panic recovered")
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (s *server) allowedMethod(methods ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, m := range methods {
				if r.Method == m {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		})
	}
}

func (s *server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.With(slog.Group(
			"request",
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
			slog.String("referer", r.Referer()),
		)).Debug("handling request")
		next.ServeHTTP(w, r)
	})
}
