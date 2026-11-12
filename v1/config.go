package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App struct {
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Environment string `yaml:"environment"`
	} `yaml:"app"`
	Server struct {
		Host           string        `yaml:"host"`
		Port           int           `yaml:"port"`
		AdminPort      int           `yaml:"admin_port"`
		ReadTimeout    time.Duration `yaml:"read_timeout"`
		WriteTimeout   time.Duration `yaml:"write_timeout"`
		IdleTimeout    time.Duration `yaml:"idle_timeout"`
		MaxHeaderBytes int           `yaml:"max_header_bytes"`
	} `yaml:"server"`
	Database struct {
		Host           string `yaml:"host"`
		Port           int    `yaml:"port"`
		Name           string `yaml:"name"`
		User           string `yaml:"user"`
		Password       string `yaml:"password"`
		SSLMode        string `yaml:"sslmode"`
		MaxConnections int32  `yaml:"max_connections"`
		MinConnections int32  `yaml:"min_connections"`
	} `yaml:"database"`
	Cache struct {
		Host       string        `yaml:"host"`
		Port       int           `yaml:"port"`
		Password   string        `yaml:"password"`
		Database   int           `yaml:"database"`
		DefaultTTL time.Duration `yaml:"default_ttl"`
	} `yaml:"cache"`
	Storage struct {
		Driver         string `yaml:"driver"`
		Path           string `yaml:"path"`
		StagingPath    string `yaml:"staging_path"`
		HotPath        string `yaml:"hot_path"`
		PersistentPath string `yaml:"persistent_path"`
		MaxUploadSize  int64  `yaml:"max_upload_size"`
		ChunkSize      int64  `yaml:"chunk_size"`
	} `yaml:"storage"`
	Security struct {
		ChecksumAlgorithm  string        `yaml:"checksum_algorithm"`
		SignatureVersion   string        `yaml:"signature_version"`
		PresignedURLExpiry time.Duration `yaml:"presigned_url_expiry"`
	} `yaml:"security"`
	Audit struct {
		Enabled       bool `yaml:"enabled"`
		RetentionDays int  `yaml:"retention_days"`
	} `yaml:"audit"`
	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
		Output string `yaml:"output"`
	} `yaml:"logging"`
	Metrics struct {
		Enabled bool   `yaml:"enabled"`
		Path    string `yaml:"path"`
	} `yaml:"metrics"`
}

func LoadConfig(file string) (*Config, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	c := new(Config)
	if err = yaml.Unmarshal(b, c); err != nil {
		return nil, err
	}
	env("FEHMI_DB_PASSWORD", &c.Database.Password)
	env("FEHMI_CACHE_PASSWORD", &c.Cache.Password)
	env("FEHMI_STORAGE_HOT_PATH", &c.Storage.HotPath)
	env("FEHMI_STORAGE_PERSISTENT_PATH", &c.Storage.PersistentPath)
	env("FEHMI_STORAGE_STAGING_PATH", &c.Storage.StagingPath)
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
func env(k string, p *string) {
	if v := os.Getenv(k); v != "" {
		*p = v
	}
}
