package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/mbatimel/AMC/products/internal/clients"
	"github.com/mbatimel/AMC/products/internal/config"
	productsService "github.com/mbatimel/AMC/products/internal/service"
	"github.com/mbatimel/AMC/products/internal/storage/postgres"
	internalapi "github.com/mbatimel/AMC/products/internal/transport/jsonRPC/internalapi"
)

const serviceName = "products-internal-api"

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
	svc := productsService.New(log.Logger, postgresStorage, accessClient)

	server := internalapi.New(
		log.Logger,
		internalapi.ProductsInternalAPI(internalapi.NewProductsInternalAPI(svc)),
	).WithLog().WithMetrics()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info().Str("address", cfg.InternalBindAddr).Msg("products internal api server started")
		if serveErr := server.Fiber().Listen(cfg.InternalBindAddr); serveErr != nil {
			log.Error().Err(serveErr).Msg("products internal server stopped")
			shutdown <- syscall.SIGTERM
		}
	}()

	<-shutdown
}
