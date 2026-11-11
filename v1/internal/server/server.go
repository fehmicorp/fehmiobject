package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/fehmicorp/fehmi-s3/internal/config"
	"github.com/fehmicorp/fehmi-s3/pkg/logger"
)

type Server struct {
	httpServer *http.Server
	log        *logger.Logger
}

func New(
	cfg config.ServerConfig,
	handler http.Handler,
	log *logger.Logger,
) *Server {
	addr := net.JoinHostPort(cfg.Host, fmt.Sprint(cfg.Port))

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ReadHeaderTimeout: 10 * time.Second,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
	}

	return &Server{
		httpServer: httpServer,
		log:        log,
	}
}

// Start starts the HTTP server and blocks until it stops.
func (s *Server) Start() error {
	s.log.Info().
		Str("address", s.httpServer.Addr).
		Msg("Starting Fehmi S3 HTTP server")

	err := s.httpServer.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("http server failed: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info().Msg("Shutting down HTTP server")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Error().
			Err(err).
			Msg("Graceful shutdown failed")

		return err
	}

	s.log.Info().Msg("HTTP server stopped")

	return nil
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return s.httpServer.Addr
}

// Handler returns the configured HTTP handler.
func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}
