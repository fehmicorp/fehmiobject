package auth

import (
	"encoding/json"
	"errors"
	"os"
)

type Authenticator struct {
	Keys map[string]string
}

func NewAuthenticator() (*Authenticator, error) {
	a := &Authenticator{Keys: map[string]string{}}
	if raw := os.Getenv("S3_API_KEYS"); raw != "" {
		var m map[string]string
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			return nil, err
		}
		for k, v := range m {
			if k != "" && v != "" {
				a.Keys[k] = v
			}
		}
	}
	if k, s := os.Getenv("S3_ACCESS_HEADER"), os.Getenv("S3_ACCESS_SECRET_KEY"); k != "" && s != "" {
		a.Keys[k] = s
	}
	if len(a.Keys) == 0 {
		return nil, errors.New("set S3_ACCESS_HEADER and S3_ACCESS_SECRET_KEY, or S3_API_KEYS JSON")
	}
	return a, nil
}
