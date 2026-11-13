package config

import (
	"fmt"
	"os"
	"time"

	"github.com/fehmicorp/s3object/v1/types"
	"go.yaml.in/yaml/v2"
)

func LoadConfig(file string) (*types.Config, error) {
	c := defaultConfig()
	if file != "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		if err = yaml.Unmarshal(b, c); err != nil {
			return nil, fmt.Errorf("unmarshal yaml: %w", err)
		}
	}
	cfg, err := GetEnvConfig(c)
	if err != nil {
		return nil, fmt.Errorf("Error in Env: %w", err)
	}
	return cfg, nil
}

func GetEnvConfig(c *types.Config) (*types.Config, error) {
	EnvStr("S3_ENV", &c.App.Environment)
	EnvInt("S3_PORT", &c.Server.Port)
	EnvStr("S3_SLUG", &c.Server.InitialSlug)
	EnvStr("S3_DB_HOST", &c.Database.Host)
	EnvInt("S3_DB_PORT", &c.Database.Port)
	EnvStr("S3_DB_NAME", &c.Database.Name)
	EnvStr("S3_DB_USERNAME", &c.Database.User)
	EnvStr("S3_DB_PASSWORD", &c.Database.Password)
	EnvStr("S3_CACHE_HOST", &c.Cache.Host)
	EnvInt("S3_CACHE_PORT", &c.Cache.Port)
	EnvStr("S3_CACHE_PASSWORD", &c.Cache.Password)
	if c.Storage.HotPath == "" {
		c.Storage.HotPath = c.Storage.Path + "/hot"
	}
	if c.Storage.PersistentPath == "" {
		c.Storage.PersistentPath = c.Storage.Path + "/persistent"
	}
	if c.Storage.StagingPath == "" {
		c.Storage.StagingPath = c.Storage.Path + "/staging"
	}
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 9000
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 300 * time.Second
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = 120 * time.Second
	}
	if c.Server.MaxHeaderBytes == 0 {
		c.Server.MaxHeaderBytes = 1 << 20
	}
	if c.Storage.MaxUploadSize == 0 {
		c.Storage.MaxUploadSize = 10 << 30
	}
	if c.Database.MaxConnections < 1 {
		c.Database.MaxConnections = 20
	}
	if c.Storage.Path == "" || c.Database.Host == "" || c.Database.Name == "" || c.Database.User == "" || c.Cache.Host == "" || c.Cache.Port < 1 {
		return nil, fmt.Errorf("storage.path, PostgreSQL host/name/user and Valkey host/port are required")
	}
	if c.Storage.Driver != "" && c.Storage.Driver != "filesystem" {
		return nil, fmt.Errorf("only filesystem storage driver is supported")
	}
	return c, nil
}
