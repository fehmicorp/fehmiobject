# Fehmi Blob Storage - initial flat Go implementation

## Run (PowerShell)
```powershell
cd D:\Fehmi\S3\v1
go mod tidy
$env:FEHMI_ACCESS_KEY = "FEHMIADMIN"
$env:FEHMI_SECRET_KEY = "replace-with-a-long-random-secret"
go run . -config ..\configs\config.yaml
```

For multiple credentials:
```powershell
$env:FEHMI_API_KEYS = '{"APP_ONE":"secret-one","APP_TWO":"secret-two"}'
```

Optional path/password overrides:
```powershell
$env:FEHMI_DB_PASSWORD = "your-postgres-password"
$env:FEHMI_CACHE_PASSWORD = "your-valkey-password"
$env:FEHMI_STORAGE_HOT_PATH = "D:\FehmiData\hot"
$env:FEHMI_STORAGE_PERSISTENT_PATH = "X:\FehmiISCSI\objects"
$env:FEHMI_STORAGE_STAGING_PATH = "D:\FehmiData\staging"
```

The persistent path is the authoritative storage tier; hot storage is a best-effort local cache. Mount and verify the iSCSI filesystem on the host before starting the service. Never mount a normal XFS/ext4 filesystem read-write from multiple hosts simultaneously.

## Current implementation
PostgreSQL startup schema, Valkey health connectivity, SigV4 Authorization-header validation for environment-configured keys, path-style bucket create/list/delete, object PUT/GET/HEAD/DELETE, basic object listing XML, staged upload with SHA-256, persistent write and local hot-cache promotion, health/readiness endpoints and audit rows.

## Not yet implemented
Full AWS S3 conformance, multipart uploads, presigned URLs, streaming SigV4 chunk signatures, ListObjectsV2 continuation tokens, IAM policy scopes, API key administration/rotation UI, versioning, ACLs, encryption at rest, real Prometheus metrics, background reconciliation and safe multi-node clustered filesystem support. Treat this as a development baseline, not production-ready storage software. Credentials are read from environment variables; never commit real secrets.
