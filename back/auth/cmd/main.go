package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mbatimel/AMC/objectstorage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"

	accessTransport "github.com/mbatimel/AMC/access/pkg/client/transport"
	"github.com/mbatimel/AMC/auth/internal/client/fns"
	"github.com/mbatimel/AMC/auth/internal/config"
	authService "github.com/mbatimel/AMC/auth/internal/service"
	postgres "github.com/mbatimel/AMC/auth/internal/storage/postgres"
	customHandlers "github.com/mbatimel/AMC/auth/internal/transport/custom-handlers"
	transportHttp "github.com/mbatimel/AMC/auth/internal/transport/http"
	"github.com/mbatimel/AMC/auth/internal/transport/jsonRPC/externalapi"
)

const serviceName = "auth-api"

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
	fnsClient := fns.New(cfg.FnsAddr, cfg.FnsKey, log.Logger)
	s3Client, err := objectstorage.New(objectstorage.Config{
		Endpoint:       cfg.S3Endpoint,
		PublicEndpoint: cfg.S3PublicEndpoint,
		AccessKey:      cfg.S3AccessKey,
		SecretKey:      cfg.S3SecretKey,
		Bucket:         cfg.S3Bucket,
		Region:         cfg.S3Region,
		UseSSL:         cfg.S3UseSSL,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create S3 client")
	}
	svc := authService.NewAuthApiService(log.Logger, postgresStorage, access, fnsClient,
		authService.WithObjectStorage(s3Client, cfg.S3MaxFileSize))

	registerIPRoutes := customHandlers.NewRegisterIPRoutes(svc, cfg.S3MaxFileSize)
	maxBodySize := int(cfg.S3MaxFileSize) + (1 << 20)

	app := externalapi.New(log.Logger,
		externalapi.MaxBodySize(maxBodySize),
		externalapi.AuthAPI(externalapi.NewAuthAPI(svc)),
		externalapi.Service(registerIPRoutes),
	).WithLog().WithMetrics()
	server := &fasthttp.Server{
		Handler:            app.Fiber().Handler(),
		MaxRequestBodySize: maxBodySize,
	}

	healthServer := transportHttp.NewHealthServer()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info().Str("address", cfg.BindAddr).Msg("auth external api server started")
		if serveErr := server.ListenAndServe(cfg.BindAddr); serveErr != nil {
			log.Fatal().Err(serveErr).Msg("failed to listen and serve auth server")
		}
	}()

	go func() {
		if healthErr := healthServer.Start(":9091"); healthErr != nil {
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
