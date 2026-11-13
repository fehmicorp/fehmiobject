package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fehmicorp/s3object/v1/internal"
)

func main() {
	config := flag.String("config", "config.yaml", "configuration YAML path")
	flag.Parse()
	c, e := internal.LoadConfig(*config)
	if e != nil {
		log.Fatalf("config: %v", e)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, e := internal.OpenDatabase(ctx, c)
	if e != nil {
		log.Fatalf("PostgreSQL: %v", e)
	}
	defer db.Close()
	cache, e := internal.OpenCache(ctx, c)
	if e != nil {
		log.Fatalf("Valkey: %v", e)
	}
	defer cache.Close()
	auth, e := internal.NewAuthenticator()
	if e != nil {
		log.Fatalf("credentials: %v", e)
	}
	store, e := internal.NewStorage(c)
	if e != nil {
		log.Fatalf("storage: %v", e)
	}
	audit := &internal.Audit{DB: db, Enabled: c.Audit.Enabled}
	s3 := &internal.S3Handler{DB: db, Store: store, Audit: audit, MaxUpload: c.Storage.MaxUploadSize}
	health := &internal.Health{DB: db, Cache: cache, Name: c.App.Name, Version: c.App.Version}
	srv := internal.NewServer(c, auth, s3, health)
	done := make(chan error, 1)
	go func() { done <- srv.Start() }()
	select {
	case e := <-done:
		if e != nil {
			log.Fatalf("HTTP server: %v", e)
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if e := srv.Shutdown(shutdown); e != nil {
			log.Printf("shutdown: %v", e)
		}
	}
}
