package internal

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type accessKeyContextKey struct{}
type AppServer struct {
	HTTP *http.Server
	Auth *Authenticator
	S3   *S3Handler
}

func NewServer(c *Config, a *Authenticator, s *S3Handler, h *Health) *AppServer {
	mux := http.NewServeMux()
	slug := c.Server.InitialSlug
	live := fmt.Sprintf("%s/health/live", slug)
	ready := fmt.Sprintf("%s/health/ready", slug)
	health := fmt.Sprintf("%s/health", slug)
	mux.HandleFunc(live, h.Live)
	mux.HandleFunc(ready, h.Ready)
	mux.HandleFunc(health, h.Ready)
	if c.Metrics.Enabled && c.Metrics.Path != "" {
		mux.HandleFunc(c.Metrics.Path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("# Fehmi Blob Storage metrics endpoint\n"))
		})
	}
	app := &AppServer{Auth: a, S3: s}
	mux.Handle(slug, app.auth(c, s))
	app.HTTP = &http.Server{Addr: c.Server.Host + ":" + strconv.Itoa(c.Server.Port), Handler: logRequests(mux), ReadHeaderTimeout: 10 * time.Second, ReadTimeout: c.Server.ReadTimeout, WriteTimeout: c.Server.WriteTimeout, IdleTimeout: c.Server.IdleTimeout, MaxHeaderBytes: c.Server.MaxHeaderBytes}
	return app
}

func (a *AppServer) auth(c *Config, next http.Handler) http.Handler {
	slug := c.Server.InitialSlug
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, fmt.Sprintf("%s/health", slug)) {
			next.ServeHTTP(w, r)
			return
		}
		key, e := a.Auth.Authenticate(r)
		if e != nil {
			http.Error(w, "AccessDenied: "+e.Error(), 403)
			a.S3.Audit.Record(r.Context(), r, "", "AuthenticationFailed", "", "", 403)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), accessKeyContextKey{}, key)))
	})
}

func (a *AppServer) Start() error {
	log.Printf("Fehmi Blob Storage listening at %s", a.HTTP.Addr)
	e := a.HTTP.ListenAndServe()
	if e == http.ErrServerClosed {
		return nil
	}
	return e
}

func (a *AppServer) Shutdown(c context.Context) error { return a.HTTP.Shutdown(c) }

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s (%s)", r.Method, r.URL.Path, r.RemoteAddr, time.Since(t))
	})
}
