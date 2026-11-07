package postgres

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/fehmicorp/fehmi-s3/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg config.DatabaseConfig) (*DB, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("database host is required")
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("invalid database port")
	}

	if cfg.Name == "" || cfg.User == "" {
		return nil, fmt.Errorf("database name and user are required")
	}

	if cfg.MaxConnections <= 0 {
		return nil, fmt.Errorf("database max_connections must be positive")
	}

	if cfg.MinConnections < 0 ||
		cfg.MinConnections > cfg.MaxConnections {
		return nil, fmt.Errorf("invalid database min_connections")
	}

	// Build PostgreSQL connection URL safely.
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, fmt.Sprint(cfg.Port)),
		Path:   "/" + cfg.Name,
	}

	query := url.Values{}
	query.Set("sslmode", cfg.SSLMode)
	u.RawQuery = query.Encode()

	// Parse pool configuration.
	poolConfig, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	// Connection pool settings.
	poolConfig.MaxConns = cfg.MaxConnections
	poolConfig.MinConns = cfg.MinConnections

	poolConfig.MaxConnLifetime = 60 * time.Minute
	poolConfig.MaxConnIdleTime = 15 * time.Minute
	poolConfig.HealthCheckPeriod = 30 * time.Second

	// Establish database connection.
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	db := &DB{
		pool: pool,
	}

	// Verify database connectivity.
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres connection failed: %w", err)
	}

	return db, nil
}

// Pool returns the underlying PostgreSQL connection pool.
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// Ping verifies database connectivity.
func (db *DB) Ping(ctx context.Context) error {
	if db == nil || db.pool == nil {
		return fmt.Errorf("postgres database is not initialized")
	}

	return db.pool.Ping(ctx)
}

// Close gracefully closes the PostgreSQL connection pool.
func (db *DB) Close() {
	if db != nil && db.pool != nil {
		db.pool.Close()
	}
}

// Stats returns PostgreSQL connection pool statistics.
func (db *DB) Stats() *pgxpool.Stat {
	if db == nil || db.pool == nil {
		return nil
	}

	return db.pool.Stat()
}
