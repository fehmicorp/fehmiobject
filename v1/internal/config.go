package internal

import (
	"fmt"
	"os"
	"strconv"
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
		InitialSlug    string        `yaml:"slug"`
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

	// Override important fields from Environment Variables
	envStr("S3_ENV", &c.App.Environment)
	envStr("S3_HOST", &c.Server.Host)
	envInt("S3_PORT", &c.Server.Port)
	envInt("S3_ADMIN_PORT", &c.Server.AdminPort)
	envStr("S3_SLUG", &c.Server.InitialSlug)
	envStr("S3_DB_HOST", &c.Database.Host)
	envInt("S3_DB_PORT", &c.Database.Port)
	envStr("S3_DB_NAME", &c.Database.Name)
	envStr("S3_DB_USERNAME", &c.Database.User)
	envStr("S3_DB_PASSWORD", &c.Database.Password)
	envStr("S3_DB_SSLMODE", &c.Database.SSLMode)
	envStr("S3_CACHE_HOST", &c.Cache.Host)
	envInt("S3_CACHE_PORT", &c.Cache.Port)
	envStr("S3_CACHE_PASSWORD", &c.Cache.Password)

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

func envStr(k string, p *string) {
	if v := os.Getenv(k); v != "" {
		*p = v
	}
}

func envInt(k string, p *int) {
	if v := os.Getenv(k); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			*p = parsed
		}
	}
}

func envBool(k string, p *bool) {
	if v := os.Getenv(k); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			*p = parsed
		}
	}
}

func defaultConfig() *Config {
	c := &Config{}
	// App
	c.App.Name = "fehmi-s3"
	c.App.Version = "0.1.0"
	c.App.Environment = "production"

	// Server
	c.Server.Host = "0.0.0.0"
	c.Server.Port = 9000
	c.Server.AdminPort = 9001
	c.Server.InitialSlug = "/"
	c.Server.ReadTimeout = 300 * time.Second
	c.Server.WriteTimeout = 300 * time.Second
	c.Server.IdleTimeout = 120 * time.Second
	c.Server.MaxHeaderBytes = 1 << 20 // 1 MB

	// Database (PostgreSQL)
	c.Database.Host = "localhost"
	c.Database.Port = 5432
	c.Database.SSLMode = "disable"
	c.Database.MaxConnections = 20
	c.Database.MinConnections = 2

	// Cache (Valkey / Redis)
	c.Cache.Host = "localhost"
	c.Cache.Port = 6379
	c.Cache.Database = 0
	c.Cache.DefaultTTL = 15 * time.Minute

	// Storage
	c.Storage.Driver = "filesystem"
	c.Storage.MaxUploadSize = 10 << 30 // 10 GB
	c.Storage.ChunkSize = 8 << 20      // 8 MB

	// Security
	c.Security.ChecksumAlgorithm = "sha256"
	c.Security.SignatureVersion = "v4"
	c.Security.PresignedURLExpiry = 15 * time.Minute

	// Audit
	c.Audit.Enabled = true
	c.Audit.RetentionDays = 90

	// Logging
	c.Logging.Level = "info"
	c.Logging.Format = "json"
	c.Logging.Output = "stdout"

	// Metrics
	c.Metrics.Enabled = true
	c.Metrics.Path = "/metrics"

	return c
}
