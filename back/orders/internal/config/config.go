package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// defaultIntegrationsTimeout must stay LARGER than the integrations service's
// own outbound timeout to 1С (ONEC_ORDERS_REQUEST_TIMEOUT, default 15s), plus a
// margin for the extra hop. This timeout wraps that one: if it expired first,
// orders would roll the order back and tell the customer it failed while 1С may
// already have created the document — a guaranteed "double order" on every slow
// 1С call instead of a rare race.
const defaultIntegrationsTimeout = 20 * time.Second

type Config struct {
	PGHost              string
	PGPort              string
	PGDB                string
	PGUser              string
	PGPassword          string
	BindAddr            string
	AccessURL           string
	VATRate             float64
	IntegrationsURL     string
	IntegrationsTimeout time.Duration
}

func LoadConfig() Config {
	cfg := Config{
		PGHost:              GetEnv("PG_HOST", "localhost"),
		PGPort:              GetEnv("PG_PORT", "5432"),
		PGDB:                os.Getenv("PG_DB"),
		PGUser:              os.Getenv("PG_USER"),
		PGPassword:          os.Getenv("PG_PASSWORD"),
		BindAddr:            GetEnv("BIND_ADDR", ":8082"),
		AccessURL:           os.Getenv("ACCESS_URL"),
		IntegrationsURL:     os.Getenv("INTEGRATIONS_URL"),
		IntegrationsTimeout: getEnvDuration("INTEGRATIONS_TIMEOUT", defaultIntegrationsTimeout),
	}

	if cfg.PGDB == "" || cfg.PGUser == "" || cfg.PGPassword == "" {
		log.Fatal().Msg("PG_DB, PG_USER and PG_PASSWORD must be specified")
	}
	if cfg.AccessURL == "" {
		cfg.AccessURL = "http://localhost:8080"
		log.Warn().Msg("ACCESS_URL must be specified")
	}
	if cfg.IntegrationsURL == "" {
		log.Fatal().Msg("INTEGRATIONS_URL must be specified")
	}
	if cfg.IntegrationsTimeout <= 0 {
		log.Fatal().Str("INTEGRATIONS_TIMEOUT", os.Getenv("INTEGRATIONS_TIMEOUT")).Msg("INTEGRATIONS_TIMEOUT must be positive")
	}

	vatRateStr := GetEnv("NDS_VALUE", "22")
	vatRate, err := strconv.ParseFloat(vatRateStr, 64)
	if err != nil {
		log.Fatal().Err(err).Str("NDS_VALUE", vatRateStr).Msg("invalid NDS_VALUE")
	}
	cfg.VATRate = vatRate

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
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
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
		value = os.Expand(value, func(name string) string {
			return os.Getenv(name)
		})
		if err = os.Setenv(key, value); err != nil {
			return err
		}
	}
	return scanner.Err()
}
