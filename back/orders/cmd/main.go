package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"

	accessTransport "github.com/mbatimel/AMC/access/pkg/client/transport"
	"github.com/mbatimel/AMC/orders/internal/config"
	ordersService "github.com/mbatimel/AMC/orders/internal/service"
	postgres "github.com/mbatimel/AMC/orders/internal/storage/postgres"
	transportHttp "github.com/mbatimel/AMC/orders/internal/transport/http"
	"github.com/mbatimel/AMC/orders/internal/transport/jsonRPC/externalapi"
)

const serviceName = "orders-api"

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
	access := accessTransport.NewClientAccessAPI(cfg.AccessURL)
	svc := ordersService.NewOrdersApiService(log.Logger, postgresStorage, access, cfg.VATRate)

	app := externalapi.New(log.Logger, externalapi.OrdersAPI(externalapi.NewOrdersAPI(svc))).WithLog().WithMetrics()
	server := &fasthttp.Server{
		Handler: app.Fiber().Handler(),
	}

	healthServer := transportHttp.NewHealthServer()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info().Str("address", cfg.BindAddr).Msg("orders external api server started")
		if serveErr := server.ListenAndServe(cfg.BindAddr); serveErr != nil {
			log.Fatal().Err(serveErr).Msg("failed to listen and serve orders server")
		}
	}()

	go func() {
		if healthErr := healthServer.Start(":9092"); healthErr != nil {
			log.Error().Err(healthErr).Msg("failed to start health server")
		}
	}()

	<-shutdown

	if err = healthServer.Stop(); err != nil {
		log.Error().Err(err).Msg("failed to stop health server")
	}

	if err = server.Shutdown(); err != nil {
		log.Error().Err(err).Msg("failed to shutdown server")
	}
}
