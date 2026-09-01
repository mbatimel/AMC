// back/admin/cmd/main.go
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
	authTransport "github.com/mbatimel/AMC/auth/pkg/client/transport"

	usersClient "github.com/mbatimel/AMC/admin/internal/client/users"
	"github.com/mbatimel/AMC/admin/internal/config"
	"github.com/mbatimel/AMC/admin/internal/mailer"
	adminService "github.com/mbatimel/AMC/admin/internal/service"
	postgres "github.com/mbatimel/AMC/admin/internal/storage/postgres"
	customHandlers "github.com/mbatimel/AMC/admin/internal/transport/custom-handlers"
	transportHttp "github.com/mbatimel/AMC/admin/internal/transport/http"
	"github.com/mbatimel/AMC/admin/internal/transport/jsonRPC/externalapi"
)

const serviceName = "admin-api"

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
	auth := authTransport.NewClientAuthAPI(cfg.AuthURL)
	users := usersClient.New(cfg.UsersURL)
	mail := mailer.NewSMTPMailer(
		log.Logger, cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword,
		cfg.SMTPFrom, cfg.SMTPTLS, cfg.SMTPTimeout,
	)
	s3Client, err := objectstorage.New(objectstorage.Config{
		Endpoint: cfg.S3Endpoint, PublicEndpoint: cfg.S3PublicEndpoint,
		AccessKey: cfg.S3AccessKey, SecretKey: cfg.S3SecretKey,
		Bucket: cfg.S3Bucket, Region: cfg.S3Region, UseSSL: cfg.S3UseSSL,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create S3 client")
	}
	svc := adminService.NewAdminApiService(log.Logger, postgresStorage, auth, access, users, mail, adminService.WithObjectStorage(s3Client, cfg.S3MaxFileSize))

	bannerRoutes := customHandlers.NewBannerRoutes(log.Logger, svc, cfg.S3MaxFileSize)
	app := externalapi.New(
		log.Logger,
		externalapi.MaxBodySize(int(cfg.S3MaxFileSize+1024*1024)),
		externalapi.AdminAPI(externalapi.NewAdminAPI(svc)),
		externalapi.Service(bannerRoutes),
	).WithLog().WithMetrics()
	server := &fasthttp.Server{
		Handler: app.Fiber().Handler(),
	}

	healthServer := transportHttp.NewHealthServer()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info().Str("address", cfg.BindAddr).Msg("admin external api server started")
		if serveErr := server.ListenAndServe(cfg.BindAddr); serveErr != nil {
			log.Fatal().Err(serveErr).Msg("failed to listen and serve admin server")
		}
	}()

	go func() {
		if healthErr := healthServer.Start(":9093"); healthErr != nil {
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
