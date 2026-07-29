package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadValidConfig(t *testing.T) {
	values := map[string]string{
		"CONNECTION_STR":                "postgres://watch",
		"BOT_TOKEN":                     "token",
		"BOT_TYPE":                      "telegram",
		"MAIN_CHAT_ID":                  "main",
		"NEXT_ALLOWED_USER_IDS":         " user-1; ;user-2 ",
		"START_TIME":                    "09:00",
		"END_TIME":                      "18:00",
		"DAYS_OFF":                      "Saturday, Sunday",
		"DATABASE_TIMEOUT":              "3s",
		"UNUSUAL_DAYS_REFRESH_INTERVAL": "10m",
	}

	cfg, err := load(func(name string) string { return values[name] })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DatabaseTimeout != 3*time.Second {
		t.Fatalf("expected 3s database timeout, got %v", cfg.DatabaseTimeout)
	}
	if len(cfg.NextAllowedUserIDs) != 2 {
		t.Fatalf("expected two allowed users, got %v", cfg.NextAllowedUserIDs)
	}
}

func TestLoadRejectsUnsafeConfig(t *testing.T) {
	base := map[string]string{
		"CONNECTION_STR": "postgres://watch",
		"BOT_TOKEN":      "token",
		"BOT_TYPE":       "telegram",
		"MAIN_CHAT_ID":   "main",
	}
	tests := []struct {
		name       string
		key        string
		value      string
		errorMatch string
	}{
		{name: "unsupported bot", key: "BOT_TYPE", value: "slack", errorMatch: "BOT_TYPE"},
		{name: "partial working hours", key: "START_TIME", value: "09:00", errorMatch: "configured together"},
		{name: "unknown day", key: "DAYS_OFF", value: "Funday", errorMatch: "DAYS_OFF"},
		{name: "zero retries", key: "RETRY_COUNT", value: "0", errorMatch: "RETRY_COUNT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := make(map[string]string, len(base)+1)
			for key, value := range base {
				values[key] = value
			}
			values[tt.key] = tt.value

			_, err := load(func(name string) string { return values[name] })
			if err == nil || !strings.Contains(err.Error(), tt.errorMatch) {
				t.Fatalf("expected error containing %q, got %v", tt.errorMatch, err)
			}
		})
	}
}
