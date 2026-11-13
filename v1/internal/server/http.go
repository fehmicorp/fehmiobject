package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/fehmicorp/s3object/v1/internal/auth"
	"github.com/fehmicorp/s3object/v1/types"
)

type AppServer struct {
	HTTP *http.Server
	Auth *auth.Authenticator
}

func NewServer(c *types.Config, a *auth.Authenticator, h *Health) *AppServer {
	mux := http.NewServeMux()
	slug := c.Server.InitialSlug
	live := fmt.Sprintf("%s/health/live", slug)
	ready := fmt.Sprintf("%s/health/ready", slug)
	health := fmt.Sprintf("%s/health", slug)
	mux.HandleFunc(live, h.Live)
	mux.HandleFunc(ready, h.Ready)
	mux.HandleFunc(health, h.Ready)
	if c.Metrics.Enabled && c.Metrics.Path != "" {
		mux.HandleFunc(c.Metrics.Path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("# Fehmi Blob Storage metrics endpoint\n"))
		})
	}
	app := &AppServer{Auth: a}
	app.HTTP = &http.Server{
		Addr:              c.Server.Host + ":" + strconv.Itoa(c.Server.Port),
		Handler:           logRequests(mux),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       c.Server.ReadTimeout,
		WriteTimeout:      c.Server.WriteTimeout,
		IdleTimeout:       c.Server.IdleTimeout,
		MaxHeaderBytes:    c.Server.MaxHeaderBytes,
	}
	return app
}

func (s *AppServer) Start() error {
	return s.HTTP.ListenAndServe()
}

func (s *AppServer) Shutdown(ctx context.Context) error {
	return s.HTTP.Shutdown(ctx)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = time.Now()
		next.ServeHTTP(w, r)
	})
}
