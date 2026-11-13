package internal

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type S3Handler struct {
	DB        *Database
	Store     *Storage
	Audit     *Audit
	MaxUpload int64
}
type s3Error struct {
	XMLName  xml.Name `xml:"Error"`
	Code     string   `xml:"Code"`
	Message  string   `xml:"Message"`
	Resource string   `xml:"Resource"`
}
type bucketList struct {
	XMLName xml.Name     `xml:"ListAllMyBucketsResult"`
	XMLNS   string       `xml:"xmlns,attr"`
	Buckets []bucketItem `xml:"Buckets>Bucket"`
}
type bucketItem struct {
	Name    string    `xml:"Name"`
	Created time.Time `xml:"CreationDate"`
}
type objectList struct {
	XMLName  xml.Name     `xml:"ListBucketResult"`
	Name     string       `xml:"Name"`
	Prefix   string       `xml:"Prefix"`
	KeyCount int          `xml:"KeyCount"`
	MaxKeys  int          `xml:"MaxKeys"`
	Contents []objectItem `xml:"Contents"`
}
type objectItem struct {
	Key          string    `xml:"Key"`
	LastModified time.Time `xml:"LastModified"`
	ETag         string    `xml:"ETag"`
	Size         int64     `xml:"Size"`
	StorageClass string    `xml:"StorageClass"`
}

func (s *S3Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, k := s3Path(r.URL)
	var e error
	switch {
	case b == "" && r.Method == "GET":
		e = s.listBuckets(w, r)
	case b != "" && k == "" && r.Method == "PUT":
		e = s.createBucket(w, r, b)
	case b != "" && k == "" && r.Method == "DELETE":
		e = s.deleteBucket(w, r, b)
	case b != "" && k == "" && r.Method == "GET":
		e = s.listObjects(w, r, b)
	case b != "" && k != "" && r.Method == "PUT":
		e = s.putObject(w, r, b, k)
	case b != "" && k != "" && r.Method == "GET":
		e = s.getObject(w, r, b, k)
	case b != "" && k != "" && r.Method == "HEAD":
		e = s.headObject(w, r, b, k)
	case b != "" && k != "" && r.Method == "DELETE":
		e = s.deleteObject(w, r, b, k)
	default:
		s.fail(w, 501, "NotImplemented", "operation not implemented", r.URL.Path)
		return
	}
	if e != nil {
		status := 500
		msg := e.Error()
		if errors.Is(e, pgx.ErrNoRows) {
			status = 404
			msg = "The requested resource does not exist"
		}
		if strings.Contains(msg, "not empty") {
			status = 409
		}
		if strings.Contains(msg, "maximum upload") {
			status = 413
		}
		s.fail(w, status, http.StatusText(status), msg, r.URL.Path)
	}
}
func s3Path(u *url.URL) (string, string) {
	p := strings.TrimPrefix(u.EscapedPath(), "/")
	if p == "" {
		return "", ""
	}
	a := strings.SplitN(p, "/", 2)
	b, _ := url.PathUnescape(a[0])
	if len(a) == 1 {
		return b, ""
	}
	k, _ := url.PathUnescape(a[1])
	return b, k
}
func (s *S3Handler) listBuckets(w http.ResponseWriter, r *http.Request) error {
	rows, e := s.DB.Pool.Query(r.Context(), `SELECT name,created_at FROM buckets ORDER BY name`)
	if e != nil {
		return e
	}
	defer rows.Close()
	out := bucketList{XMLNS: "http://s3.amazonaws.com/doc/2006-03-01/"}
	for rows.Next() {
		var b bucketItem
		if e = rows.Scan(&b.Name, &b.Created); e != nil {
			return e
		}
		out.Buckets = append(out.Buckets, b)
	}
	if e = rows.Err(); e != nil {
		return e
	}
	return xmlOut(w, 200, out)
}
func (s *S3Handler) createBucket(w http.ResponseWriter, r *http.Request, b string) error {
	if len(b) < 3 || len(b) > 63 {
		return fmt.Errorf("invalid bucket name")
	}
	_, e := s.DB.Pool.Exec(r.Context(), `INSERT INTO buckets(name) VALUES($1)`, b)
	if e != nil {
		return e
	}
	w.Header().Set("Location", "/"+b)
	w.WriteHeader(200)
	s.record(r, "CreateBucket", b, "", 200)
	return nil
}
func (s *S3Handler) deleteBucket(w http.ResponseWriter, r *http.Request, b string) error {
	var n int
	if e := s.DB.Pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM objects WHERE bucket=$1`, b).Scan(&n); e != nil {
		return e
	}
	if n > 0 {
		return fmt.Errorf("bucket is not empty")
	}
	tag, e := s.DB.Pool.Exec(r.Context(), `DELETE FROM buckets WHERE name=$1`, b)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	w.WriteHeader(204)
	s.record(r, "DeleteBucket", b, "", 204)
	return nil
}
func (s *S3Handler) listObjects(w http.ResponseWriter, r *http.Request, b string) error {
	prefix := r.URL.Query().Get("prefix")
	max := 1000
	if v := r.URL.Query().Get("max-keys"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n >= 0 && n <= 1000 {
			max = n
		}
	}
	rows, e := s.DB.Pool.Query(r.Context(), `SELECT object_key,size,etag,updated_at FROM objects WHERE bucket=$1 AND object_key LIKE $2 ESCAPE '\' ORDER BY object_key LIMIT $3`, b, strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(prefix)+"%", max)
	if e != nil {
		return e
	}
	defer rows.Close()
	out := objectList{Name: b, Prefix: prefix, MaxKeys: max}
	for rows.Next() {
		var o objectItem
		if e = rows.Scan(&o.Key, &o.Size, &o.ETag, &o.LastModified); e != nil {
			return e
		}
		o.StorageClass = "STANDARD"
		out.Contents = append(out.Contents, o)
	}
	out.KeyCount = len(out.Contents)
	if e = rows.Err(); e != nil {
		return e
	}
	return xmlOut(w, 200, out)
}
func (s *S3Handler) putObject(w http.ResponseWriter, r *http.Request, b, k string) error {
	if r.ContentLength > s.MaxUpload {
		return fmt.Errorf("maximum upload size exceeded")
	}
	payload := r.Header.Get("X-Amz-Content-Sha256")
	if payload == "STREAMING-AWS4-HMAC-SHA256-PAYLOAD" {
		return fmt.Errorf("streaming SigV4 payload is not supported yet")
	}
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	obj, e := s.Store.Put(b, k, r.Body)
	if e != nil {
		return e
	}
	if payload != "" && payload != "UNSIGNED-PAYLOAD" && !strings.EqualFold(payload, obj.SHA256) {
		_ = s.Store.Delete(b, k)
		return fmt.Errorf("payload checksum does not match signed x-amz-content-sha256")
	}
	_, e = s.DB.Pool.Exec(r.Context(), `INSERT INTO objects(bucket,object_key,size,etag,checksum_sha256,content_type,storage_path) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(bucket,object_key) DO UPDATE SET size=EXCLUDED.size,etag=EXCLUDED.etag,checksum_sha256=EXCLUDED.checksum_sha256,content_type=EXCLUDED.content_type,storage_path=EXCLUDED.storage_path,updated_at=NOW()`, b, k, obj.Size, obj.ETag, obj.SHA256, ct, obj.Path)
	if e != nil {
		return e
	}
	w.Header().Set("ETag", `"`+obj.ETag+`"`)
	w.WriteHeader(200)
	s.record(r, "PutObject", b, k, 200)
	return nil
}
func (s *S3Handler) getObject(w http.ResponseWriter, r *http.Request, b, k string) error {
	var size int64
	var etag, ct string
	if e := s.DB.Pool.QueryRow(r.Context(), `SELECT size,etag,content_type FROM objects WHERE bucket=$1 AND object_key=$2`, b, k).Scan(&size, &etag, &ct); e != nil {
		return e
	}
	f, tier, e := s.Store.Open(b, k)
	if e != nil {
		return e
	}
	defer f.Close()
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.Header().Set("ETag", `"`+etag+`"`)
	w.Header().Set("X-Fehmi-Storage-Tier", tier)
	w.WriteHeader(200)
	_, e = io.Copy(w, f)
	s.record(r, "GetObject", b, k, 200)
	return e
}
func (s *S3Handler) headObject(w http.ResponseWriter, r *http.Request, b, k string) error {
	var n int64
	var et, ct string
	e := s.DB.Pool.QueryRow(r.Context(), `SELECT size,etag,content_type FROM objects WHERE bucket=$1 AND object_key=$2`, b, k).Scan(&n, &et, &ct)
	if e != nil {
		return e
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Length", strconv.FormatInt(n, 10))
	w.Header().Set("ETag", `"`+et+`"`)
	w.WriteHeader(200)
	return nil
}
func (s *S3Handler) deleteObject(w http.ResponseWriter, r *http.Request, b, k string) error {
	var exists bool
	e := s.DB.Pool.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM objects WHERE bucket=$1 AND object_key=$2)`, b, k).Scan(&exists)
	if e != nil {
		return e
	}
	if exists {
		if e = s.Store.Delete(b, k); e != nil {
			return e
		}
		_, e = s.DB.Pool.Exec(r.Context(), `DELETE FROM objects WHERE bucket=$1 AND object_key=$2`, b, k)
		if e != nil {
			return e
		}
	}
	w.WriteHeader(204)
	s.record(r, "DeleteObject", b, k, 204)
	return nil
}
func (s *S3Handler) record(r *http.Request, a, b, k string, status int) {
	key, _ := r.Context().Value(accessKeyContextKey{}).(string)
	s.Audit.Record(r.Context(), r, key, a, b, k, status)
}
func (s *S3Handler) fail(w http.ResponseWriter, status int, code, msg, res string) {
	_ = xmlOut(w, status, s3Error{Code: code, Message: msg, Resource: res})
}
func xmlOut(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	return xml.NewEncoder(w).Encode(v)
}
