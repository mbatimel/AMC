package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/mbatimel/AMC/integrations/internal/config"
	"github.com/mbatimel/AMC/integrations/internal/onec"
	"github.com/mbatimel/AMC/integrations/internal/service"
	"github.com/mbatimel/AMC/integrations/internal/storage/postgres"
	transportHTTP "github.com/mbatimel/AMC/integrations/internal/transport/http"
)

const serviceName = "onec-sync"

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Str("serviceName", serviceName).Logger()

	if err := config.LoadEnvFile(".env"); err != nil && !os.IsNotExist(err) {
		log.Fatal().Err(err).Msg("load env file")
	}
	cfg := config.LoadConfig()

	pool, err := postgres.NewPool(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer pool.Close()

	storageImpl := postgres.New(pool)
	onecClient := onec.New(cfg.OnecBaseURL, cfg.OnecUser, cfg.OnecPassword, log.Logger)
	svc := service.New(log.Logger, onecClient, storageImpl)

	healthServer := transportHTTP.NewHealthServer()
	go func() {
		if healthErr := healthServer.Start(cfg.HealthAddr); healthErr != nil {
			log.Error().Err(healthErr).Msg("onec-sync health server stopped")
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-shutdown
		cancel()
	}()

	log.Info().Dur("interval", cfg.SyncInterval).Msg("onec sync worker started")
	runLoop(ctx, cfg.SyncInterval, svc.RunSync, func(err error) {
		log.Error().Err(err).Msg("onec sync run failed")
	})

	if err = healthServer.Stop(); err != nil {
		log.Error().Err(err).Msg("failed to stop health server")
	}
}
