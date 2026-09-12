package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type Config struct {
	BindAddr            string
	HealthAddr          string
	AccessURL           string
	ProductsInternalURL string

	WildberriesAPIURL string
	WildberriesAPIKey string

	OzonAPIURL   string
	OzonClientID string
	OzonAPIKey   string

	YandexMarketAPIURL     string
	YandexMarketAPIKey     string
	YandexMarketBusinessID int64

	MarketplaceClientTimeout time.Duration
}

func LoadConfig() Config {
	return Config{
		BindAddr:            GetEnv("BIND_ADDR", ":8086"),
		HealthAddr:          GetEnv("HEALTH_ADDR", ":9096"),
		AccessURL:           GetEnv("ACCESS_URL", "http://localhost:8080"),
		ProductsInternalURL: GetEnv("PRODUCTS_INTERNAL_URL", "http://localhost:8091"),

		WildberriesAPIURL: GetEnv("WILDBERRIES_API_URL", "https://content-api.wildberries.ru"),
		WildberriesAPIKey: os.Getenv("WILDBERRIES_API_KEY"),

		OzonAPIURL:   GetEnv("OZON_API_URL", "https://api-seller.ozon.ru"),
		OzonClientID: os.Getenv("OZON_CLIENT_ID"),
		OzonAPIKey:   os.Getenv("OZON_API_KEY"),

		YandexMarketAPIURL:     GetEnv("YANDEX_MARKET_API_URL", "https://api.partner.market.yandex.ru"),
		YandexMarketAPIKey:     os.Getenv("YANDEX_MARKET_API_KEY"),
		YandexMarketBusinessID: getEnvInt64("YANDEX_MARKET_BUSINESS_ID", 0),

		MarketplaceClientTimeout: getEnvDuration("MARKETPLACE_CLIENT_TIMEOUT", 15*time.Second),
	}
}

func getEnvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Fatal().Err(err).Str("key", key).Msg("invalid integer environment variable")
	}
	return parsed
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
