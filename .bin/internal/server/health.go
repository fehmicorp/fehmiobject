package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/fehmicorp/fehmi-s3/internal/cache/valkey"
	"github.com/fehmicorp/fehmi-s3/internal/metadata/postgres"
)

type HealthService struct {
	DB    *postgres.DB
	Cache *valkey.Client
}

type DependencyHealth struct {
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

type HealthResponse struct {
	Status       string                      `json:"status"`
	Service      string                      `json:"service"`
	Timestamp    time.Time                   `json:"timestamp"`
	Dependencies map[string]DependencyHealth `json:"dependencies,omitempty"`
}

func NewHealthService(
	db *postgres.DB,
	cache *valkey.Client,
) *HealthService {
	return &HealthService{
		DB:    db,
		Cache: cache,
	}
}

// Live checks whether the application process is running.
func (h *HealthService) Live(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:    "alive",
		Service:   "fehmi-s3",
		Timestamp: time.Now().UTC(),
	})
}

// Ready verifies the required infrastructure dependencies.
func (h *HealthService) Ready(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:       "ready",
		Service:      "fehmi-s3",
		Timestamp:    time.Now().UTC(),
		Dependencies: make(map[string]DependencyHealth),
	}

	ctx := r.Context()

	// Check PostgreSQL.
	response.Dependencies["postgres"] = checkDependency(
		ctx,
		func(ctx context.Context) error {
			if h.DB == nil {
				return context.Canceled
			}
			return h.DB.Ping(ctx)
		},
	)

	// Check Valkey.
	response.Dependencies["valkey"] = checkDependency(
		ctx,
		func(ctx context.Context) error {
			if h.Cache == nil {
				return context.Canceled
			}
			return h.Cache.Ping(ctx)
		},
	)

	for _, dependency := range response.Dependencies {
		if dependency.Status != "healthy" {
			response.Status = "degraded"
			break
		}
	}

	statusCode := http.StatusOK

	if response.Status != "ready" {
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(w, statusCode, response)
}

func checkDependency(
	parent context.Context,
	check func(context.Context) error,
) DependencyHealth {
	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	defer cancel()

	start := time.Now()

	err := check(ctx)

	duration := time.Since(start)

	result := DependencyHealth{
		Status:   "healthy",
		Duration: duration.String(),
	}

	if err != nil {
		result.Status = "unhealthy"
		result.Error = err.Error()
	}

	return result
}

// Health returns a general application health response.
func (h *HealthService) Health(w http.ResponseWriter, r *http.Request) {
	h.Ready(w, r)
}

// JSON helper for health responses.
func writeHealthJSON(
	w http.ResponseWriter,
	status int,
	response HealthResponse,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(response)
}
