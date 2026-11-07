package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func Load(configPath string) (*Config, error) {
	// Load .env for local development.
	if os.Getenv("APP_ENV") != "production" {
		_ = godotenv.Load()
	}

	v := viper.New()

	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	v.SetEnvPrefix("FEHMI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "fehmi-s3")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "development")

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 9000)
	v.SetDefault("server.admin_port", 9001)
	v.SetDefault("server.read_timeout", "300s")
	v.SetDefault("server.write_timeout", "300s")
	v.SetDefault("server.idle_timeout", "120s")
	v.SetDefault("server.max_header_bytes", 1048576)

	v.SetDefault("database.port", 5432)
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_connections", 50)
	v.SetDefault("database.min_connections", 5)

	v.SetDefault("cache.host", "localhost")
	v.SetDefault("cache.port", 6379)
	v.SetDefault("cache.database", 0)
	v.SetDefault("cache.default_ttl", "300s")

	v.SetDefault("storage.driver", "filesystem")
	v.SetDefault("storage.path", "./data")
	v.SetDefault("storage.staging_path", "./data/staging")
	v.SetDefault("storage.max_upload_size", int64(10*1024*1024*1024))
	v.SetDefault("storage.chunk_size", int64(16*1024*1024))

	v.SetDefault("security.checksum_algorithm", "sha256")
	v.SetDefault("security.signature_version", "AWS4-HMAC-SHA256")
	v.SetDefault("security.presigned_url_expiry", "15m")

	v.SetDefault("audit.enabled", true)
	v.SetDefault("audit.retention_days", 90)

	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")

	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.path", "/metrics")
}

func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}

	if c.Server.AdminPort < 1 || c.Server.AdminPort > 65535 {
		return fmt.Errorf("server.admin_port must be between 1 and 65535")
	}

	if c.Server.Port == c.Server.AdminPort {
		return fmt.Errorf("server and admin ports must be different")
	}

	if c.Database.Host == "" || c.Database.Name == "" || c.Database.User == "" {
		return fmt.Errorf("database host, name and user are required")
	}

	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return fmt.Errorf("database.port is invalid")
	}

	if c.Cache.Host == "" || c.Cache.Port < 1 || c.Cache.Port > 65535 {
		return fmt.Errorf("cache host and valid port are required")
	}

	if c.Storage.Driver != "filesystem" {
		return fmt.Errorf("unsupported storage driver: %s", c.Storage.Driver)
	}

	if c.Storage.Path == "" || c.Storage.StagingPath == "" {
		return fmt.Errorf("storage paths are required")
	}

	if c.Storage.MaxUploadSize <= 0 || c.Storage.ChunkSize <= 0 {
		return fmt.Errorf("storage upload size and chunk size must be positive")
	}

	if c.Storage.ChunkSize > c.Storage.MaxUploadSize {
		return fmt.Errorf("storage chunk size cannot exceed maximum upload size")
	}

	if c.Audit.RetentionDays < 0 {
		return fmt.Errorf("audit.retention_days cannot be negative")
	}

	return nil
}
