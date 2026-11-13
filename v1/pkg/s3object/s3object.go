package s3object

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/fehmicorp/s3object/v1/internal"
)

type Service struct {
	db     *internal.Database
	cache  *internal.Cache
	server *internal.AppServer
}

// Open initializes the Fehmi Blob Storage engine.
func Open(ctx context.Context, configPath string) (*Service, error) {
	cfg, err := internal.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	db, err := internal.OpenDatabase(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	cache, err := internal.OpenCache(ctx, cfg)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("open cache: %w", err)
	}

	auth, err := internal.NewAuthenticator()
	if err != nil {
		cache.Close()
		db.Close()
		return nil, fmt.Errorf("initialize authentication: %w", err)
	}

	store, err := internal.NewStorage(cfg)
	if err != nil {
		cache.Close()
		db.Close()
		return nil, fmt.Errorf("initialize storage: %w", err)
	}

	audit := &internal.Audit{
		DB:      db,
		Enabled: cfg.Audit.Enabled,
	}

	s3 := &internal.S3Handler{
		DB:        db,
		Store:     store,
		Audit:     audit,
		MaxUpload: cfg.Storage.MaxUploadSize,
	}

	health := &internal.Health{
		DB:      db,
		Cache:   cache,
		Name:    cfg.App.Name,
		Version: cfg.App.Version,
	}

	server := internal.NewServer(cfg, auth, s3, health)

	return &Service{
		db:     db,
		cache:  cache,
		server: server,
	}, nil
}

// Handler returns the HTTP handler for mounting into another application.
func (s *Service) Handler() http.Handler {
	return s.server.Handler()
}

// Shutdown closes the service gracefully.
func (s *Service) Shutdown(ctx context.Context) error {
	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			return err
		}
	}

	if s.cache != nil {
		s.cache.Close()
	}

	if s.db != nil {
		s.db.Close()
	}

	return nil
}

// Close performs a graceful shutdown with a default timeout.
func (s *Service) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.Shutdown(ctx)
}
