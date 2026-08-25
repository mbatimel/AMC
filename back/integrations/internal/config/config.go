package config

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type Config struct {
	PGHost       string
	PGPort       string
	PGDB         string
	PGUser       string
	PGPassword   string
	HealthAddr   string
	OnecBaseURL  string
	OnecUser     string
	OnecPassword string
	SyncInterval time.Duration
}

func LoadConfig() Config {
	cfg := Config{
		PGHost:       GetEnv("PG_HOST", "localhost"),
		PGPort:       GetEnv("PG_PORT", "5432"),
		PGDB:         os.Getenv("PG_DB"),
		PGUser:       os.Getenv("PG_USER"),
		PGPassword:   os.Getenv("PG_PASSWORD"),
		HealthAddr:   GetEnv("HEALTH_ADDR", ":9096"),
		OnecBaseURL:  os.Getenv("ONEC_BASE_URL"),
		OnecUser:     os.Getenv("ONEC_USER"),
		OnecPassword: os.Getenv("ONEC_PASSWORD"),
		SyncInterval: getEnvDuration("SYNC_INTERVAL", 24*time.Hour),
	}
	if cfg.PGDB == "" || cfg.PGUser == "" || cfg.PGPassword == "" {
		log.Fatal().Msg("PG_DB, PG_USER and PG_PASSWORD must be specified")
	}
	if cfg.OnecBaseURL == "" || cfg.OnecUser == "" || cfg.OnecPassword == "" {
		log.Fatal().Msg("ONEC_BASE_URL, ONEC_USER and ONEC_PASSWORD must be specified")
	}
	return cfg
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Fatal().Err(err).Str("key", key).Msg("invalid duration environment variable")
	}
	return parsed
}

func GetEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func LoadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		value = os.Expand(value, os.Getenv)
		if err = os.Setenv(key, value); err != nil {
			return err
		}
	}
	return scanner.Err()
}
