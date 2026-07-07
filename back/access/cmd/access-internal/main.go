package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/mbatimel/AMC/access/internal/repository"
	"github.com/mbatimel/AMC/access/internal/service"
	internalapi "github.com/mbatimel/AMC/access/internal/transport/jsonRPC/internalapi"
)

type config struct {
	PGHost      string
	PGPort      string
	PGDB        string
	PGUser      string
	PGPassword  string
	HTTPAddress string
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	if err := loadEnvFile(".env"); err != nil && !os.IsNotExist(err) {
		log.Fatal().Err(err).Msg("load env file")
	}

	cfg := loadConfig()
	db, err := openPostgres(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("connect postgres")
	}
	defer db.Close()

	accessRepository := repository.New(db)
	accessService := service.New(accessRepository)

	server := internalapi.New(
		log.Logger,
		internalapi.AccessAPI(internalapi.NewAccessAPI(accessService)),
	)

	go func() {
		log.Info().Str("address", cfg.HTTPAddress).Msg("access internal server started")
		err := server.Fiber().Listen(cfg.HTTPAddress)
		internalapi.ExitOnError(log.Logger, err, "serve access internal on "+cfg.HTTPAddress)
	}()

	waitShutdown(server)
}

func loadConfig() config {
	cfg := config{
		PGHost:      getEnv("PG_HOST", "localhost"),
		PGPort:      getEnv("PG_PORT", "5432"),
		PGDB:        os.Getenv("PG_DB"),
		PGUser:      os.Getenv("PG_USER"),
		PGPassword:  os.Getenv("PG_PASSWORD"),
		HTTPAddress: getEnv("ACCESS_INTERNAL_ADDRESS", ":8080"),
	}

	if cfg.PGDB == "" || cfg.PGUser == "" || cfg.PGPassword == "" {
		log.Fatal().Msg("PG_DB, PG_USER and PG_PASSWORD must be specified")
	}
	return cfg
}

func openPostgres(cfg config) (*sql.DB, error) {
	port, err := strconv.Atoi(cfg.PGPort)
	if err != nil {
		return nil, fmt.Errorf("parse PG_PORT: %w", err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s sslmode=disable user=%s password=%s",
		cfg.PGHost,
		port,
		cfg.PGDB,
		cfg.PGUser,
		cfg.PGPassword,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func loadEnvFile(path string) error {
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

func waitShutdown(server *internalapi.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	server.Shutdown()
}
