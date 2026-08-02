package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

type Config struct {
	PGHost           string
	PGPort           string
	PGDB             string
	PGUser           string
	PGPassword       string
	BindAddr         string
	HealthAddr       string
	AccessURL        string
	S3Endpoint       string
	S3PublicEndpoint string
	S3AccessKey      string
	S3SecretKey      string
	S3Bucket         string
	S3Region         string
	S3UseSSL         bool
	S3MaxFileSize    int64
}

func LoadConfig() Config {
	cfg := Config{
		PGHost:           GetEnv("PG_HOST", "localhost"),
		PGPort:           GetEnv("PG_PORT", "5432"),
		PGDB:             os.Getenv("PG_DB"),
		PGUser:           os.Getenv("PG_USER"),
		PGPassword:       os.Getenv("PG_PASSWORD"),
		BindAddr:         GetEnv("BIND_ADDR", ":8081"),
		HealthAddr:       GetEnv("HEALTH_ADDR", ":9091"),
		AccessURL:        GetEnv("ACCESS_URL", "http://localhost:8080"),
		S3Endpoint:       os.Getenv("S3_ENDPOINT"),
		S3PublicEndpoint: os.Getenv("S3_PUBLIC_ENDPOINT"),
		S3AccessKey:      os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:      os.Getenv("S3_SECRET_KEY"),
		S3Bucket:         os.Getenv("S3_BUCKET"),
		S3Region:         GetEnv("S3_REGION", "us-east-1"),
		S3UseSSL:         getEnvBool("S3_USE_SSL", false),
		S3MaxFileSize:    getEnvInt64("S3_MAX_FILE_SIZE", 10*1024*1024),
	}
	if cfg.PGDB == "" || cfg.PGUser == "" || cfg.PGPassword == "" {
		log.Fatal().Msg("PG_DB, PG_USER and PG_PASSWORD must be specified")
	}
	if cfg.S3Endpoint == "" || cfg.S3PublicEndpoint == "" || cfg.S3AccessKey == "" || cfg.S3SecretKey == "" || cfg.S3Bucket == "" {
		log.Fatal().Msg("S3_ENDPOINT, S3_PUBLIC_ENDPOINT, S3_ACCESS_KEY, S3_SECRET_KEY and S3_BUCKET must be specified")
	}
	return cfg
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		log.Fatal().Err(err).Str("key", key).Msg("invalid boolean environment variable")
	}
	return parsed
}

func getEnvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		log.Fatal().Err(err).Str("key", key).Msg("invalid positive integer environment variable")
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
