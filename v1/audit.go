package main

import (
	"context"
	"log"
	"net"
	"net/http"
)

type Audit struct {
	DB      *Database
	Enabled bool
}

func (a *Audit) Record(ctx context.Context, r *http.Request, key, action, bucket, obj string, status int) {
	if a == nil || !a.Enabled || a.DB == nil {
		return
	}
	host, _, e := net.SplitHostPort(r.RemoteAddr)
	if e != nil {
		host = r.RemoteAddr
	}
	_, e = a.DB.Pool.Exec(ctx, `INSERT INTO audit_events(access_key,action,bucket,object_key,remote_addr,status_code) VALUES($1,$2,$3,$4,$5,$6)`, nullable(key), action, nullable(bucket), nullable(obj), host, status)
	if e != nil {
		log.Printf("audit insert: %v", e)
	}
}
func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
