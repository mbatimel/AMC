package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"

	"github.com/mbatimel/AMC/users/internal/clients"
	"github.com/mbatimel/AMC/users/internal/config"
	usersService "github.com/mbatimel/AMC/users/internal/service"
	postgres "github.com/mbatimel/AMC/users/internal/storage/postgres"
	transportHttp "github.com/mbatimel/AMC/users/internal/transport/http"
	"github.com/mbatimel/AMC/users/internal/transport/jsonRPC/externalapi"
)

const serviceName = "users-api"

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

	postgresStorage := postgres.New(pool)
	accessClient := clients.NewAccessClient(cfg.AccessURL)
	svc := usersService.New(log.Logger, postgresStorage, accessClient)

	app := externalapi.New(log.Logger, externalapi.UsersAPI(externalapi.NewUsersAPI(svc))).WithLog().WithMetrics()
	server := &fasthttp.Server{Handler: app.Fiber().Handler()}
	healthServer := transportHttp.NewHealthServer()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info().Str("address", cfg.BindAddr).Msg("users external api server started")
		if serveErr := server.ListenAndServe(cfg.BindAddr); serveErr != nil {
			log.Error().Err(serveErr).Msg("users server stopped")
			shutdown <- syscall.SIGTERM
		}
	}()
	go func() {
		if healthErr := healthServer.Start(":9093"); healthErr != nil {
			log.Error().Err(healthErr).Msg("users health server stopped")
		}
	}()

	<-shutdown
	if err = healthServer.Stop(); err != nil {
		log.Error().Err(err).Msg("failed to stop health server")
	}
	if err = server.Shutdown(); err != nil {
		log.Error().Err(err).Msg("failed to shutdown users server")
	}
}
