package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fehmicorp/fehmi-s3/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type RouterConfig struct {
	App     config.AppConfig
	Metrics config.MetricsConfig
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Core HTTP middleware.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health/live"))
	r.Use(middleware.Timeout(60 * time.Second))

	// Root endpoint.
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"name":    cfg.App.Name,
			"version": cfg.App.Version,
			"status":  "running",
		})
	})

	// Health endpoints.
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "ok",
		})
	})

	r.Get("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "ready",
		})
	})

	// Prometheus metrics.
	if cfg.Metrics.Enabled {
		r.Handle(cfg.Metrics.Path, promhttp.Handler())
	}

	// System API.
	r.Route("/api/v1/system", func(r chi.Router) {
		r.Get("/info", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"name":        cfg.App.Name,
				"version":     cfg.App.Version,
				"environment": cfg.App.Environment,
				"status":      "running",
			})
		})
	})

	// S3 API routes will be registered here.
	r.Route("/s3", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{
				"status": "available",
			})
		})
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return
	}
}
