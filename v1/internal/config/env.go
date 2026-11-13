package config

import (
	"os"
	"strconv"
)

func EnvStr(k string, p *string) {
	if v := os.Getenv(k); v != "" {
		*p = v
	}
}

func EnvInt(k string, p *int) {
	if v := os.Getenv(k); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			*p = parsed
		}
	}
}

func EnvBool(k string, p *bool) {
	if v := os.Getenv(k); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			*p = parsed
		}
	}
}
