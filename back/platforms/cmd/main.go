package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"

	"github.com/mbatimel/AMC/platforms/internal/clients"
	"github.com/mbatimel/AMC/platforms/internal/config"
	"github.com/mbatimel/AMC/platforms/internal/ozon"
	platformsService "github.com/mbatimel/AMC/platforms/internal/service"
	transportHTTP "github.com/mbatimel/AMC/platforms/internal/transport/http"
	"github.com/mbatimel/AMC/platforms/internal/transport/jsonRPC/externalapi"
	"github.com/mbatimel/AMC/platforms/internal/wildberries"
	"github.com/mbatimel/AMC/platforms/internal/yandexmarket"
)

const serviceName = "platforms-api"

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Str("serviceName", serviceName).Logger()

	if err := config.LoadEnvFile(".env"); err != nil && !os.IsNotExist(err) {
		log.Fatal().Err(err).Msg("load env file")
	}
	cfg := config.LoadConfig()

	accessClient := clients.NewAccessClient(cfg.AccessURL)
	productsClient := clients.NewProductsClient(cfg.ProductsInternalURL)
	wildberriesClient := wildberries.New(cfg.WildberriesAPIURL, cfg.WildberriesAPIKey, cfg.MarketplaceClientTimeout, log.Logger)
	ozonClient := ozon.New(cfg.OzonAPIURL, cfg.OzonClientID, cfg.OzonAPIKey, cfg.MarketplaceClientTimeout, log.Logger)
	yandexMarketClient := yandexmarket.New(cfg.YandexMarketAPIURL, cfg.YandexMarketBusinessID, cfg.YandexMarketAPIKey, cfg.MarketplaceClientTimeout, log.Logger)

	svc := platformsService.New(log.Logger, accessClient, productsClient, wildberriesClient, ozonClient, yandexMarketClient)

	app := externalapi.New(
		log.Logger,
		externalapi.PlatformsAPI(externalapi.NewPlatformsAPI(svc)),
	).WithLog().WithMetrics()
	server := &fasthttp.Server{Handler: app.Fiber().Handler()}
	healthServer := transportHTTP.NewHealthServer()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info().Str("address", cfg.BindAddr).Msg("platforms external api server started")
		if serveErr := server.ListenAndServe(cfg.BindAddr); serveErr != nil {
			log.Error().Err(serveErr).Msg("platforms server stopped")
			shutdown <- syscall.SIGTERM
		}
	}()
	go func() {
		if healthErr := healthServer.Start(cfg.HealthAddr); healthErr != nil {
			log.Error().Err(healthErr).Msg("platforms health server stopped")
		}
	}()

	<-shutdown
	if err := healthServer.Stop(); err != nil {
		log.Error().Err(err).Msg("failed to stop health server")
	}
	if err := server.Shutdown(); err != nil {
		log.Error().Err(err).Msg("failed to shutdown platforms server")
	}
}
