package config

import "time"

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Security SecurityConfig `mapstructure:"security"`
	Audit    AuditConfig    `mapstructure:"audit"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

type ServerConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	AdminPort      int           `mapstructure:"admin_port"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

type DatabaseConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	Name           string `mapstructure:"name"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	SSLMode        string `mapstructure:"sslmode"`
	MaxConnections int32  `mapstructure:"max_connections"`
	MinConnections int32  `mapstructure:"min_connections"`
}

type CacheConfig struct {
	Host       string        `mapstructure:"host"`
	Port       int           `mapstructure:"port"`
	Password   string        `mapstructure:"password"`
	Database   int           `mapstructure:"database"`
	DefaultTTL time.Duration `mapstructure:"default_ttl"`
}

type StorageConfig struct {
	Driver        string `mapstructure:"driver"`
	Path          string `mapstructure:"path"`
	StagingPath   string `mapstructure:"staging_path"`
	MaxUploadSize int64  `mapstructure:"max_upload_size"`
	ChunkSize     int64  `mapstructure:"chunk_size"`
}

type SecurityConfig struct {
	ChecksumAlgorithm  string        `mapstructure:"checksum_algorithm"`
	SignatureVersion   string        `mapstructure:"signature_version"`
	PresignedURLExpiry time.Duration `mapstructure:"presigned_url_expiry"`
}

type AuditConfig struct {
	Enabled       bool `mapstructure:"enabled"`
	RetentionDays int  `mapstructure:"retention_days"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}
