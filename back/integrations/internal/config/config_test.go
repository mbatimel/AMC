package config

import (
	"os"
	"testing"
	"time"
)

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	old, existed := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("set env %s: %v", key, err)
	}
	t.Cleanup(func() {
		if existed {
			os.Setenv(key, old)
		} else {
			os.Unsetenv(key)
		}
	})
}

func requiredEnv(t *testing.T) {
	t.Helper()
	setEnv(t, "PG_DB", "amc")
	setEnv(t, "PG_USER", "amc")
	setEnv(t, "PG_PASSWORD", "secret")
	setEnv(t, "ONEC_BASE_URL", "http://pviserver/UT/odata/standard.odata")
	setEnv(t, "ONEC_USER", "site")
	setEnv(t, "ONEC_PASSWORD", "site-pass")
}

func TestLoadConfig_Defaults(t *testing.T) {
	requiredEnv(t)
	os.Unsetenv("SYNC_INTERVAL")
	os.Unsetenv("HEALTH_ADDR")
	os.Unsetenv("PG_HOST")
	os.Unsetenv("PG_PORT")

	cfg := LoadConfig()

	if cfg.PGHost != "localhost" {
		t.Errorf("expected default PGHost=localhost, got %s", cfg.PGHost)
	}
	if cfg.PGPort != "5432" {
		t.Errorf("expected default PGPort=5432, got %s", cfg.PGPort)
	}
	if cfg.HealthAddr != ":9096" {
		t.Errorf("expected default HealthAddr=:9096, got %s", cfg.HealthAddr)
	}
	if cfg.SyncInterval != 24*time.Hour {
		t.Errorf("expected default SyncInterval=24h, got %s", cfg.SyncInterval)
	}
	if cfg.OnecBaseURL != "http://pviserver/UT/odata/standard.odata" {
		t.Errorf("unexpected OnecBaseURL: %s", cfg.OnecBaseURL)
	}
}

func TestLoadConfig_CustomSyncInterval(t *testing.T) {
	requiredEnv(t)
	setEnv(t, "SYNC_INTERVAL", "1h")

	cfg := LoadConfig()

	if cfg.SyncInterval != time.Hour {
		t.Errorf("expected SyncInterval=1h, got %s", cfg.SyncInterval)
	}
}

func TestLoadConfig_DefaultRequestTimeout(t *testing.T) {
	requiredEnv(t)
	os.Unsetenv("ONEC_REQUEST_TIMEOUT")

	cfg := LoadConfig()

	if cfg.OnecRequestTimeout != 30*time.Second {
		t.Errorf("expected default OnecRequestTimeout=30s, got %s", cfg.OnecRequestTimeout)
	}
}

func TestLoadConfig_CustomRequestTimeout(t *testing.T) {
	requiredEnv(t)
	setEnv(t, "ONEC_REQUEST_TIMEOUT", "5s")

	cfg := LoadConfig()

	if cfg.OnecRequestTimeout != 5*time.Second {
		t.Errorf("expected OnecRequestTimeout=5s, got %s", cfg.OnecRequestTimeout)
	}
}
