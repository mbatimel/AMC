package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mbatimel/AMC/objectstorage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"

	"github.com/mbatimel/AMC/products/internal/clients"
	"github.com/mbatimel/AMC/products/internal/config"
	productsService "github.com/mbatimel/AMC/products/internal/service"
	postgres "github.com/mbatimel/AMC/products/internal/storage/postgres"
	customHandlers "github.com/mbatimel/AMC/products/internal/transport/custom-handlers"
	transportHTTP "github.com/mbatimel/AMC/products/internal/transport/http"
	"github.com/mbatimel/AMC/products/internal/transport/jsonRPC/externalapi"
)

const serviceName = "products-api"

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
	s3Client, err := objectstorage.New(objectstorage.Config{
		Endpoint: cfg.S3Endpoint, PublicEndpoint: cfg.S3PublicEndpoint,
		AccessKey: cfg.S3AccessKey, SecretKey: cfg.S3SecretKey,
		Bucket: cfg.S3Bucket, Region: cfg.S3Region, UseSSL: cfg.S3UseSSL,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create S3 client")
	}
	svc := productsService.New(log.Logger, postgresStorage, accessClient, productsService.WithObjectStorage(s3Client, cfg.S3MaxFileSize))

	imageRoutes := customHandlers.NewProductImageRoutes(log.Logger, svc, cfg.S3MaxFileSize)
	app := externalapi.New(
		log.Logger,
		externalapi.MaxBodySize(int(cfg.S3MaxFileSize*customHandlers.MaxBatchImageFiles)),
		externalapi.ProductsAPI(externalapi.NewProductsAPI(svc)),
		externalapi.Service(imageRoutes),
	).WithLog().WithMetrics()
	server := &fasthttp.Server{Handler: app.Fiber().Handler()}
	healthServer := transportHTTP.NewHealthServer()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info().Str("address", cfg.BindAddr).Msg("products external api server started")
		if serveErr := server.ListenAndServe(cfg.BindAddr); serveErr != nil {
			log.Error().Err(serveErr).Msg("products server stopped")
			shutdown <- syscall.SIGTERM
		}
	}()
	go func() {
		if healthErr := healthServer.Start(cfg.HealthAddr); healthErr != nil {
			log.Error().Err(healthErr).Msg("products health server stopped")
		}
	}()

	<-shutdown
	if err = healthServer.Stop(); err != nil {
		log.Error().Err(err).Msg("failed to stop health server")
	}
	if err = server.Shutdown(); err != nil {
		log.Error().Err(err).Msg("failed to shutdown products server")
	}
}
