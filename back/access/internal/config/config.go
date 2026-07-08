package config

import (
	"bufio"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

type Config struct {
	PGHost      string
	PGPort      string
	PGDB        string
	PGUser      string
	PGPassword  string
	HTTPAddress string
}

func LoadConfig() Config {
	cfg := Config{
		PGHost:      GetEnv("PG_HOST", "localhost"),
		PGPort:      GetEnv("PG_PORT", "5432"),
		PGDB:        os.Getenv("PG_DB"),
		PGUser:      os.Getenv("PG_USER"),
		PGPassword:  os.Getenv("PG_PASSWORD"),
		HTTPAddress: GetEnv("ACCESS_INTERNAL_ADDRESS", ":8080"),
	}

	if cfg.PGDB == "" || cfg.PGUser == "" || cfg.PGPassword == "" {
		log.Fatal().Msg("PG_DB, PG_USER and PG_PASSWORD must be specified")
	}
	return cfg
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
