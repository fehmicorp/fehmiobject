package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

type Authenticator struct {
	Keys   map[string]string
	Region string
}

func NewAuthenticator() (*Authenticator, error) {
	a := &Authenticator{Keys: map[string]string{}, Region: "us-east-1"}
	if raw := os.Getenv("FEHMI_API_KEYS"); raw != "" {
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
	if k, s := os.Getenv("FEHMI_ACCESS_KEY"), os.Getenv("FEHMI_SECRET_KEY"); k != "" && s != "" {
		a.Keys[k] = s
	}
	if len(a.Keys) == 0 {
		return nil, errors.New("set FEHMI_ACCESS_KEY and FEHMI_SECRET_KEY, or FEHMI_API_KEYS JSON")
	}
	return a, nil
}
func (a *Authenticator) Authenticate(r *http.Request) (string, error) {
	const p = "AWS4-HMAC-SHA256 "
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, p) {
		return "", errors.New("missing SigV4 Authorization header")
	}
	f := map[string]string{}
	for _, x := range strings.Split(strings.TrimPrefix(auth, p), ",") {
		v := strings.SplitN(strings.TrimSpace(x), "=", 2)
		if len(v) == 2 {
			f[v[0]] = v[1]
		}
	}
	scope := strings.Split(f["Credential"], "/")
	if len(scope) != 5 || scope[2] != a.Region || scope[3] != "s3" || scope[4] != "aws4_request" {
		return "", errors.New("invalid credential scope")
	}
	secret, ok := a.Keys[scope[0]]
	if !ok {
		return "", errors.New("unknown access key")
	}
	signed := strings.Split(f["SignedHeaders"], ";")
	if len(signed) == 0 || f["Signature"] == "" {
		return "", errors.New("invalid signed headers")
	}
	date := r.Header.Get("X-Amz-Date")
	t, err := time.Parse("20060102T150405Z", date)
	if err != nil {
		return "", errors.New("invalid X-Amz-Date")
	}
	if d := time.Since(t); d > 15*time.Minute || d < -15*time.Minute {
		return "", errors.New("request timestamp outside 15-minute window")
	}
	payload := r.Header.Get("X-Amz-Content-Sha256")
	if payload == "" {
		return "", errors.New("missing x-amz-content-sha256")
	}
	headers := ""
	for _, name := range signed {
		name = strings.ToLower(name)
		v := r.Host
		if name != "host" {
			v = strings.Join(r.Header.Values(http.CanonicalHeaderKey(name)), ",")
		}
		headers += name + ":" + strings.TrimSpace(v) + "\n"
	}
	uri := r.URL.EscapedPath()
	if uri == "" {
		uri = "/"
	}
	canonical := strings.Join([]string{r.Method, uri, canonicalQuery(r.URL.Query()), headers, strings.Join(signed, ";"), payload}, "\n")
	sum := sha256.Sum256([]byte(canonical))
	dateKey := hmac256([]byte("AWS4"+secret), scope[1])
	regionKey := hmac256(dateKey, scope[2])
	serviceKey := hmac256(regionKey, "s3")
	signingKey := hmac256(serviceKey, "aws4_request")
	toSign := "AWS4-HMAC-SHA256\n" + date + "\n" + strings.Join(scope[1:], "/") + "\n" + hex.EncodeToString(sum[:])
	expected := hex.EncodeToString(hmac256(signingKey, toSign))
	if !hmac.Equal([]byte(strings.ToLower(expected)), []byte(strings.ToLower(f["Signature"]))) {
		return "", errors.New("signature mismatch")
	}
	return scope[0], nil
}
func hmac256(k []byte, s string) []byte {
	h := hmac.New(sha256.New, k)
	_, _ = h.Write([]byte(s))
	return h.Sum(nil)
}
func canonicalQuery(q url.Values) string {
	type pair struct{ k, v string }
	var a []pair
	esc := func(s string) string {
		x := url.QueryEscape(s)
		x = strings.ReplaceAll(x, "+", "%20")
		return strings.ReplaceAll(x, "%7E", "~")
	}
	for k, vs := range q {
		for _, v := range vs {
			a = append(a, pair{esc(k), esc(v)})
		}
	}
	sort.Slice(a, func(i, j int) bool {
		if a[i].k == a[j].k {
			return a[i].v < a[j].v
		}
		return a[i].k < a[j].k
	})
	out := make([]string, 0, len(a))
	for _, p := range a {
		out = append(out, p.k+"="+p.v)
	}
	return strings.Join(out, "&")
}
