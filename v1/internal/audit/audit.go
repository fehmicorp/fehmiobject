package audit

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/fehmicorp/s3object/v1/internal/db"
)

type Audit struct {
	DB      *db.DB
	Enabled bool
}

func (a *Audit) Record(ctx context.Context, r *http.Request, key, action, bucket, obj string, status int) {
	if a == nil || !a.Enabled || a.DB == nil || a.DB.Database == nil {
		return
	}

	host, _, e := net.SplitHostPort(r.RemoteAddr)
	if e != nil {
		host = r.RemoteAddr
	}

	// Construct the document, omitting empty optional fields
	event := map[string]any{
		"action":      action,
		"remote_addr": host,
		"status_code": status,
		"timestamp":   time.Now(),
	}

	if key != "" {
		event["access_key"] = key
	}
	if bucket != "" {
		event["bucket"] = bucket
	}
	if obj != "" {
		event["object_key"] = obj
	}

	_, err := a.DB.Database.Collection("auditlogs").InsertOne(ctx, event)
	if err != nil {
		log.Printf("audit insert error: %v", err)
	}
}
