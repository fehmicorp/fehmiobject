package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/fehmicorp/s3object/v1/internal/cache"
	"github.com/fehmicorp/s3object/v1/internal/db"
)

type Health struct {
	DB            *db.DB
	Cache         *cache.Cache
	Name, Version string
}

func (h *Health) Live(w http.ResponseWriter, r *http.Request) {
	jsonOut(w, 200, map[string]any{"status": "alive", "app": h.Name, "version": h.Version})
}

func (h *Health) Ready(w http.ResponseWriter, r *http.Request) {
	c, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Ping MongoDB to check database readiness
	mongoOk := false
	if h.DB != nil && h.DB.Client != nil {
		if err := h.DB.Client.Ping(c, nil); err == nil {
			mongoOk = true
		}
	}

	status := 200
	state := "ready"
	if !mongoOk {
		status = 503
		state = "not_ready"
	}

	jsonOut(w, status, map[string]any{"status": state, "mongodb": mongoOk})
}

func jsonOut(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
