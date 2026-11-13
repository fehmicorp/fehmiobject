package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fehmicorp/s3object/v1/internal/auth"
	"github.com/fehmicorp/s3object/v1/internal/cache"
	"github.com/fehmicorp/s3object/v1/internal/config"
	"github.com/fehmicorp/s3object/v1/internal/db"
	"github.com/fehmicorp/s3object/v1/internal/server"
)

func main() {
	conf := flag.String("config", "config.yaml", "configuration YAML path")
	flag.Parse()
	c, e := config.LoadConfig(*conf)
	if e != nil {
		log.Fatalf("config: %v", e)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, e := db.Connect(ctx, &c.Database)
	if e != nil {
		log.Fatalf("MongoDB: %v", e)
	}
	defer db.Close(ctx)

	memory, e := cache.OpenCache(ctx, &c.Cache)
	if e != nil {
		log.Fatalf("Valkey: %v", e)
	}
	defer memory.Close()

	authInst, e := auth.NewAuthenticator()
	if e != nil {
		log.Fatalf("credentials: %v", e)
	}

	// store, e := storage.NewStorage(c)
	// if e != nil {
	// 	log.Fatalf("storage: %v", e)
	// }

	// audit := &audit.Audit{DB: db, Enabled: c.Audit.Enabled}

	health := &server.Health{DB: db, Cache: memory, Name: c.App.Name, Version: c.App.Version}

	srv := server.NewServer(c, authInst, health)

	done := make(chan error, 1)
	go func() { done <- srv.HTTP.ListenAndServe() }()
	select {
	case e := <-done:
		if e != nil {
			log.Fatalf("HTTP server: %v", e)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if e := srv.Shutdown(shutdownCtx); e != nil {
			log.Printf("shutdown: %v", e)
		}
	}
}
