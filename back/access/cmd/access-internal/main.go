package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	config "github.com/mbatimel/AMC/access/internal/config"
	"github.com/mbatimel/AMC/access/internal/repository"
	"github.com/mbatimel/AMC/access/internal/service"
	postgres "github.com/mbatimel/AMC/access/internal/storage/postgres"
	internalapi "github.com/mbatimel/AMC/access/internal/transport/jsonRPC/internalapi"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	if err := config.LoadEnvFile(".env"); err != nil && !os.IsNotExist(err) {
		log.Fatal().Err(err).Msg("load env file")
	}

	cfg := config.LoadConfig()
	db, err := postgres.OpenPostgres(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("connect postgres")
	}
	defer db.Close()

	accessStorage := postgres.New(db)
	accessRepository := repository.New(accessStorage)
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

func waitShutdown(server *internalapi.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	server.Shutdown()
}
