package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mbatimel/AMC/auth/internal/config"
	"github.com/mbatimel/AMC/auth/internal/service"
	"github.com/mbatimel/AMC/auth/internal/storage/postgres"
	"github.com/mbatimel/AMC/auth/internal/storage/redis"
	externalapi "github.com/mbatimel/AMC/auth/internal/transport/jsonRPC/externalapi"
)

const redisDB = 0

func main() {
	cfg := config.Values()
	logger := cfg.Logger()

	conn, err := postgres.NewManager(logger, cfg.Postgres)
	if err != nil {
		logger.Fatal().Err(err).Msg("connect postgres")
	}

	redisAddr := "localhost:6379"
	if len(cfg.RedisAddrs) > 0 {
		redisAddr = cfg.RedisAddrs[0]
	}
	cache := redis.NewCacheRepo(redisAddr, cfg.RedisPassword, redisDB, logger)

	storage := postgres.NewStorage(conn, logger)
	authService := service.NewAuthService(storage, cache, logger)

	server := externalapi.New(
		logger,
		externalapi.AuthAPI(externalapi.NewAuthAPI(authService)),
	)
	server.WithMetrics()
	server.ServeMetrics(logger, cfg.MetricsPath, cfg.MetricsBind)
	server.ServeHealth(cfg.HealthBind, map[string]string{"status": "ok"})

	go func() {
		logger.Info().Str("address", cfg.ServiceBind).Msg("auth external server started")
		err := server.Fiber().Listen(cfg.ServiceBind)
		externalapi.ExitOnError(logger, err, "serve auth external on "+cfg.ServiceBind)
	}()

	waitShutdown(server)
}

func waitShutdown(server *externalapi.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	server.Shutdown()
}
