package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/SkYNewZ/gh-stars-search-engine/internal/engine"
	"github.com/SkYNewZ/gh-stars-search-engine/internal/slogx"
)

//go:generate go run github.com/vburenin/ifacemaker --file $GOFILE --struct server --iface Server --pkg http --output server_iface.go
type server struct {
	logger     *slog.Logger
	httpServer *http.Server

	search        engine.Engine
	searchTimeout time.Duration
}

// NewServer returns a new HTTP server.
func NewServer(logger *slog.Logger, search engine.Engine, searchTimeout time.Duration) Server {
	if logger == nil {
		logger = slog.Default()
	}

	srv := &server{
		logger:        logger,
		search:        search,
		searchTimeout: searchTimeout,
		httpServer: &http.Server{
			Addr:         "", // will be set by Start
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      nil,
		},
	}

	router := http.NewServeMux()
	router.HandleFunc("/search", srv.searchHandler)
	router.HandleFunc("/health", srv.healthHandler)
	router.HandleFunc("/", srv.uiHandler)

	// setup default middlewares
	r := srv.recoverMiddleware(router)
	r = srv.allowedMethod(http.MethodGet, http.MethodOptions)(r)
	r = srv.loggingMiddleware(r)
	srv.httpServer.Handler = r

	return srv
}

// Start starts the HTTP server.
func (s *server) Start() {
	port := "8080"
	if v := os.Getenv("PORT"); v != "" {
		port = v
	}

	s.logger.Info("starting HTTP server on port " + port)
	s.httpServer.Addr = ":" + port
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.With(slogx.Err(err)).Error("failed to start HTTP server")
	}
}

// Stop stops the HTTP server.
func (s *server) Stop(ctx context.Context) {
	s.logger.Info("stopping HTTP server")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.With(slogx.Err(err)).Error("failed to stop HTTP server")
	}
}
