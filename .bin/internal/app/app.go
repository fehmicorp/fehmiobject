package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fehmicorp/fehmi-s3/internal/cache/valkey"
	"github.com/fehmicorp/fehmi-s3/internal/config"
	"github.com/fehmicorp/fehmi-s3/internal/metadata/postgres"
	"github.com/fehmicorp/fehmi-s3/internal/server"
	"github.com/fehmicorp/fehmi-s3/pkg/logger"
)

type App struct {
	Config *config.Config
	Logger *logger.Logger

	DB    *postgres.DB
	Cache *valkey.Client

	Server *server.Server
}

func New(ctx context.Context, configPath string) (*App, error) {
	// 1. Load configuration.
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load configuration: %w", err)
	}

	// 2. Initialize logger.
	log := logger.New(logger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
		Output: cfg.Logging.Output,
	})

	log.Info().
		Str("version", cfg.App.Version).
		Str("environment", cfg.App.Environment).
		Msg("Initializing Fehmi S3")

	app := &App{
		Config: cfg,
		Logger: log,
	}

	// 3. Connect PostgreSQL.
	db, err := postgres.New(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("initialize postgres: %w", err)
	}

	app.DB = db

	log.Info().Msg("PostgreSQL connected successfully")

	// 4. Connect Valkey.
	cache, err := valkey.New(ctx, cfg.Cache)
	if err != nil {
		app.Close()
		return nil, fmt.Errorf("initialize valkey: %w", err)
	}

	app.Cache = cache

	log.Info().Msg("Valkey connected successfully")

	// 5. Initialize health service.
	healthService := server.NewHealthService(
		app.DB,
		app.Cache,
	)

	// 6. Initialize HTTP router.
	router := server.NewRouter(server.RouterConfig{
		App:     cfg.App,
		Metrics: cfg.Metrics,
		Health:  healthService,
	})

	// 7. Initialize HTTP server.
	app.Server = server.New(
		cfg.Server,
		router,
		log.WithComponent("http"),
	)

	log.Info().Msg("Fehmi S3 initialized successfully")

	return app, nil
}

// Run starts the application and waits for shutdown.
func (a *App) Run(ctx context.Context) error {
	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- a.Server.Start()
	}()

	select {
	case err := <-serverErrors:
		if err != nil {
			return err
		}

		return nil

	case <-ctx.Done():
		a.Logger.Info().Msg("Application shutdown requested")

	case <-time.After(0):
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	if err := a.Server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	a.Logger.Info().Msg("Fehmi S3 stopped successfully")

	return nil
}

// Close releases infrastructure resources.
func (a *App) Close() {
	if a == nil {
		return
	}

	if a.Cache != nil {
		if err := a.Cache.Close(); err != nil && a.Logger != nil {
			a.Logger.Error().
				Err(err).
				Msg("Failed to close Valkey")
		}
	}

	if a.DB != nil {
		a.DB.Close()
	}

	if a.Logger != nil {
		_ = a.Logger.Sync()
	}
}

// Execute starts Fehmi S3 with operating system signal handling.
func Execute(configPath string) error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	app, err := New(ctx, configPath)
	if err != nil {
		return err
	}

	defer app.Close()

	err = app.Run(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
