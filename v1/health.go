package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Health struct {
	DB            *Database
	Cache         *Cache
	Name, Version string
}

func (h *Health) Live(w http.ResponseWriter, r *http.Request) {
	jsonOut(w, 200, map[string]any{"status": "alive", "app": h.Name, "version": h.Version})
}
func (h *Health) Ready(w http.ResponseWriter, r *http.Request) {
	c, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	db := h.DB != nil && h.DB.Pool.Ping(c) == nil
	cache := h.Cache != nil && h.Cache.Client.Ping(c).Err() == nil
	status := 200
	state := "ready"
	if !db || !cache {
		status = 503
		state = "not_ready"
	}
	jsonOut(w, status, map[string]any{"status": state, "postgres": db, "valkey": cache})
}
func jsonOut(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
