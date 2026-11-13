package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Storage struct {
	Hot, Persistent, Staging string
	MaxUpload                int64
}
type StoredObject struct {
	Path   string
	Size   int64
	SHA256 string
	ETag   string
}

func NewStorage(c *Config) (*Storage, error) {
	s := &Storage{Hot: c.Storage.HotPath, Persistent: c.Storage.PersistentPath, Staging: c.Storage.StagingPath, MaxUpload: c.Storage.MaxUploadSize}
	for _, p := range []string{s.Hot, s.Persistent, s.Staging} {
		if err := os.MkdirAll(p, 0750); err != nil {
			return nil, err
		}
	}
	return s, nil
}
func objectID(b, k string) string {
	x := sha256.Sum256([]byte(b + "\x00" + k))
	return hex.EncodeToString(x[:])
}
func (s *Storage) obj(root, b, k string) string {
	id := objectID(b, k)
	return filepath.Join(root, id[:2], id[2:4], id)
}
func (s *Storage) Put(b, k string, r io.Reader) (*StoredObject, error) {
	stage := filepath.Join(s.Staging, objectID(b, k)+".upload")
	f, err := os.OpenFile(stage, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	n, e := io.Copy(io.MultiWriter(f, h), io.LimitReader(r, s.MaxUpload+1))
	if e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e == nil {
		e = ce
	}
	if e != nil {
		_ = os.Remove(stage)
		return nil, e
	}
	if n > s.MaxUpload {
		_ = os.Remove(stage)
		return nil, fmt.Errorf("maximum upload size exceeded")
	}
	dst := s.obj(s.Persistent, b, k)
	if err = os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		_ = os.Remove(stage)
		return nil, err
	}
	if err = copyAtomic(stage, dst); err != nil {
		_ = os.Remove(stage)
		return nil, err
	}
	_ = os.Remove(stage)
	hot := s.obj(s.Hot, b, k)
	if os.MkdirAll(filepath.Dir(hot), 0750) == nil {
		_ = copyAtomic(dst, hot)
	}
	sum := hex.EncodeToString(h.Sum(nil))
	return &StoredObject{dst, n, sum, sum}, nil
}
func copyAtomic(src, dst string) error {
	in, e := os.Open(src)
	if e != nil {
		return e
	}
	defer in.Close()
	tmp, e := os.CreateTemp(filepath.Dir(dst), ".fehmi-*")
	if e != nil {
		return e
	}
	name := tmp.Name()
	defer os.Remove(name)
	_, e = io.Copy(tmp, in)
	if e == nil {
		e = tmp.Sync()
	}
	ce := tmp.Close()
	if e == nil {
		e = ce
	}
	if e != nil {
		return e
	}
	return os.Rename(name, dst)
}
func (s *Storage) Open(b, k string) (*os.File, string, error) {
	hot := s.obj(s.Hot, b, k)
	if f, e := os.Open(hot); e == nil {
		return f, "hot", nil
	}
	p := s.obj(s.Persistent, b, k)
	f, e := os.Open(p)
	if e != nil {
		return nil, "", e
	}
	if os.MkdirAll(filepath.Dir(hot), 0750) == nil {
		_ = copyAtomic(p, hot)
	}
	return f, "persistent", nil
}
func (s *Storage) Delete(b, k string) error {
	var errs []error
	for _, p := range []string{s.obj(s.Hot, b, k), s.obj(s.Persistent, b, k)} {
		if e := os.Remove(p); e != nil && !os.IsNotExist(e) {
			errs = append(errs, e)
		}
	}
	return errors.Join(errs...)
}
