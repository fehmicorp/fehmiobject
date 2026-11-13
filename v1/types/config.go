package types

import "time"

type Config struct {
	App      App      `yaml:"app"`
	Server   Server   `yaml:"server"`
	Database Database `yaml:"database"`
	Cache    Cache    `yaml:"cache"`
	Storage  Storage  `yaml:"storage"`
	Security Security `yaml:"security"`
	Audit    Audit    `yaml:"audit"`
	Logging  Logging  `yaml:"logging"`
	Metrics  Metrics  `yaml:"metrics"`
}

type App struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

type Server struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	InitialSlug    string        `yaml:"slug"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	MaxHeaderBytes int           `yaml:"max_header_bytes"`
}

type Database struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Name           string `yaml:"name"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	SSLMode        string `yaml:"sslmode"`
	MaxConnections int32  `yaml:"max_connections"`
	MinConnections int32  `yaml:"min_connections"`
}

type Cache struct {
	Host       string        `yaml:"host"`
	Port       int           `yaml:"port"`
	Password   string        `yaml:"password"`
	Database   int           `yaml:"database"`
	DefaultTTL time.Duration `yaml:"default_ttl"`
}

type Storage struct {
	Driver         string `yaml:"driver"`
	Path           string `yaml:"path"`
	StagingPath    string `yaml:"staging_path"`
	HotPath        string `yaml:"hot_path"`
	PersistentPath string `yaml:"persistent_path"`
	MaxUploadSize  int64  `yaml:"max_upload_size"`
	ChunkSize      int64  `yaml:"chunk_size"`
}

type Security struct {
	ChecksumAlgorithm  string        `yaml:"checksum_algorithm"`
	SignatureVersion   string        `yaml:"signature_version"`
	PresignedURLExpiry time.Duration `yaml:"presigned_url_expiry"`
}

type Audit struct {
	Enabled       bool `yaml:"enabled"`
	RetentionDays int  `yaml:"retention_days"`
}

type Logging struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

type Metrics struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}
