package config

import (
	"time"

	"github.com/fehmicorp/s3object/v1/types"
)

func defaultConfig() *types.Config {
	c := &types.Config{}
	c.App.Name = "fehmi-s3"
	c.App.Version = "0.1.0"
	c.App.Environment = "production"

	c.Server.Host = "0.0.0.0"
	c.Server.Port = 9000
	c.Server.InitialSlug = "/"
	c.Server.ReadTimeout = 300 * time.Second
	c.Server.WriteTimeout = 300 * time.Second
	c.Server.IdleTimeout = 120 * time.Second
	c.Server.MaxHeaderBytes = 1 << 20 // 1 MB

	c.Database.Host = "localhost"
	c.Database.Port = 27017
	c.Database.Name = "s3object"
	c.Database.SSLMode = "disable"
	c.Database.MaxConnections = 20
	c.Database.MinConnections = 2

	c.Cache.Host = "localhost"
	c.Cache.Port = 6379
	c.Cache.Database = 0
	c.Cache.DefaultTTL = 15 * time.Minute

	c.Storage.Driver = "filesystem"
	c.Storage.MaxUploadSize = 10 << 30 // 10 GB
	c.Storage.ChunkSize = 8 << 20      // 8 MB

	c.Security.ChecksumAlgorithm = "sha256"
	c.Security.SignatureVersion = "v4"
	c.Security.PresignedURLExpiry = 15 * time.Minute

	c.Audit.Enabled = true
	c.Audit.RetentionDays = 90

	c.Logging.Level = "info"
	c.Logging.Format = "json"
	c.Logging.Output = "stdout"

	c.Metrics.Enabled = true
	c.Metrics.Path = "/metrics"

	return c
}
