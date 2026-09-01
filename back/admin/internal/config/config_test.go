package config

import (
	"testing"
	"time"
)

func TestLoadConfigInvalidOptionalSMTPValuesUseDefaults(t *testing.T) {
	required := map[string]string{
		"PG_DB":              "admin",
		"PG_USER":            "admin",
		"PG_PASSWORD":        "secret",
		"S3_ENDPOINT":        "http://minio:9000",
		"S3_PUBLIC_ENDPOINT": "http://localhost:9000",
		"S3_ACCESS_KEY":      "access",
		"S3_SECRET_KEY":      "secret",
		"S3_BUCKET":          "admin",
	}
	for key, value := range required {
		t.Setenv(key, value)
	}
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_FROM", "")
	t.Setenv("SMTP_TLS", "not-a-bool")
	t.Setenv("SMTP_TIMEOUT", "not-a-duration")

	cfg := LoadConfig()

	if !cfg.SMTPTLS {
		t.Fatal("SMTPTLS = false, want safe default true")
	}
	if cfg.SMTPTimeout != 10*time.Second {
		t.Fatalf("SMTPTimeout = %s, want 10s", cfg.SMTPTimeout)
	}
}
