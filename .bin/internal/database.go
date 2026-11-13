package internal

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct{ Pool *pgxpool.Pool }

func OpenDatabase(ctx context.Context, c *Config) (*Database, error) {
	u := url.URL{Scheme: "postgres", User: url.UserPassword(c.Database.User, c.Database.Password), Host: fmt.Sprintf("%s:%d", c.Database.Host, c.Database.Port), Path: c.Database.Name}
	q := u.Query()
	q.Set("sslmode", c.Database.SSLMode)
	u.RawQuery = q.Encode()
	pc, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		return nil, err
	}
	pc.MaxConns = c.Database.MaxConnections
	pc.MinConns = c.Database.MinConnections
	pc.MaxConnLifetime = 30 * time.Minute
	p, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		return nil, err
	}
	ping, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if err = p.Ping(ping); err != nil {
		p.Close()
		return nil, err
	}
	d := &Database{Pool: p}
	if err = d.Migrate(ctx); err != nil {
		p.Close()
		return nil, err
	}
	return d, nil
}
func (d *Database) Migrate(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS buckets (name TEXT PRIMARY KEY, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS objects (bucket TEXT NOT NULL REFERENCES buckets(name) ON DELETE CASCADE, object_key TEXT NOT NULL, size BIGINT NOT NULL, etag TEXT NOT NULL, checksum_sha256 TEXT NOT NULL, content_type TEXT NOT NULL, storage_path TEXT NOT NULL, updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), PRIMARY KEY(bucket, object_key))`,
		`CREATE INDEX IF NOT EXISTS objects_bucket_updated_idx ON objects(bucket, updated_at DESC)`,
		`CREATE TABLE IF NOT EXISTS audit_events (id BIGSERIAL PRIMARY KEY, access_key TEXT, action TEXT NOT NULL, bucket TEXT, object_key TEXT, remote_addr TEXT, status_code INTEGER NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
	}
	for _, statement := range statements {
		if _, err := d.Pool.Exec(ctx, statement); err != nil {
			return fmt.Errorf("database migration: %w", err)
		}
	}
	return nil
}

func (d *Database) Close() {
	if d != nil && d.Pool != nil {
		d.Pool.Close()
	}
}
