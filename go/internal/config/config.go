package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	// Comma separated tokens
	Tokens []string
	// Comma separated notification URLs
	NotificationURLs []string
	// MaxRetries for each character attendance
	MaxRetries int
}

const (
	envTokens           = "TOKENS"
	envNotificationURLs = "NOTIFICATION_URLS"
	envMaxRetries       = "MAX_RETRIES"
)

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		Tokens:           splitAndTrim(os.Getenv(envTokens)),
		NotificationURLs: splitAndTrim(os.Getenv(envNotificationURLs)),
		MaxRetries:       3,
	}

	if v := os.Getenv(envMaxRetries); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MaxRetries = n
		}
	}

	return cfg, nil
}

func splitAndTrim(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

