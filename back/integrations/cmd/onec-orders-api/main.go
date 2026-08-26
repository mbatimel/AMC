package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"

	"github.com/mbatimel/AMC/integrations/internal/config"
	"github.com/mbatimel/AMC/integrations/internal/onecorders"
	"github.com/mbatimel/AMC/integrations/internal/service"
	"github.com/mbatimel/AMC/integrations/internal/storage/postgres"
	transportHTTP "github.com/mbatimel/AMC/integrations/internal/transport/http"
	"github.com/mbatimel/AMC/integrations/internal/transport/jsonRPC/internalapi"
	ordersTransport "github.com/mbatimel/AMC/orders/pkg/client/transport"
)

const serviceName = "onec-orders-api"

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Str("serviceName", serviceName).Logger()

	if err := config.LoadEnvFile(".env"); err != nil && !os.IsNotExist(err) {
		log.Fatal().Err(err).Msg("load env file")
	}
	cfg := config.LoadConfig()
	if cfg.OnecOrdersBaseURL == "" || cfg.OnecOrdersUser == "" || cfg.OnecOrdersPassword == "" {
		log.Fatal().Msg("ONEC_ORDERS_BASE_URL, ONEC_ORDERS_USER and ONEC_ORDERS_PASSWORD must be specified")
	}
	if cfg.OnecWebhookAPIKey == "" || cfg.OrdersURL == "" || cfg.OrdersSystemUserID == uuid.Nil {
		log.Fatal().Msg("ONEC_WEBHOOK_API_KEY, ORDERS_URL and ORDERS_SYSTEM_USER_ID must be specified")
	}

	pool, err := postgres.NewPool(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer pool.Close()
	storageImpl := postgres.New(pool)

	pushClient := onecorders.New(cfg.OnecOrdersBaseURL, cfg.OnecOrdersUser, cfg.OnecOrdersPassword, cfg.OnecOrdersRequestTimeout, log.Logger)
	ordersClient := ordersClientAdapter{cli: ordersTransport.NewClientOrdersAPI(cfg.OrdersURL)}

	webhookSvc := service.NewOnecOrdersService(log.Logger, ordersClient, cfg.OrdersSystemUserID, cfg.OnecWebhookAPIKey)

	app := internalapi.New(log.Logger, internalapi.OnecOrdersAPI(internalapi.NewOnecOrdersAPI(webhookSvc))).WithLog().WithMetrics()
	transportHTTP.RegisterPushOrderRoute(app.Fiber(), pushOrderAdapter{client: pushClient, storage: storageImpl}, log.Logger)

	server := &fasthttp.Server{Handler: app.Fiber().Handler()}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info().Str("address", cfg.OnecOrdersBindAddr).Msg("onec-orders-api started")
		if serveErr := server.ListenAndServe(cfg.OnecOrdersBindAddr); serveErr != nil {
			log.Fatal().Err(serveErr).Msg("failed to listen and serve onec-orders-api")
		}
	}()

	<-shutdown
	if err = server.Shutdown(); err != nil {
		log.Error().Err(err).Msg("failed to shutdown server")
	}
}

// ordersClientAdapter adapts the generated orders API client to
// service.OrdersClient: the generated client returns response structs that
// wrap the values service.OrdersClient expects as bare returns.
type ordersClientAdapter struct {
	cli *ordersTransport.ClientOrdersAPI
}

func (a ordersClientAdapter) GetOrderStatus(ctx context.Context, userID, orderID uuid.UUID) (string, error) {
	resp, err := a.cli.GetOrderStatus(ctx, userID, orderID)
	return resp.Status, err
}

func (a ordersClientAdapter) UpdateOrderStatus(ctx context.Context, userID, orderID uuid.UUID, status, paymentStatus, comment, changedBy string) error {
	_, err := a.cli.UpdateOrderStatus(ctx, userID, orderID, status, paymentStatus, comment, changedBy)
	return err
}

// pushOrderAdapter implements transportHTTP.OnecPusher. It pushes an order
// to 1С via onecorders.Client and records the attempt as an integration job
// via the storage layer (Task 11).
type pushOrderAdapter struct {
	client  *onecorders.Client
	storage *postgres.Storage
}

func (a pushOrderAdapter) PushOrder(ctx context.Context, req onecorders.PushOrderRequest) (onecorders.PushOrderResult, error) {
	systemID, err := a.storage.UpsertIntegrationSystem(ctx, "onec_orders", "1С:УТ 10.3 (заказы)")
	if err != nil {
		log.Error().Err(err).Msg("upsert integration system for push order failed")
	}
	jobID, jobErr := a.storage.CreateIntegrationJob(ctx, systemID, "outbound", "order_create")
	if jobErr != nil {
		log.Error().Err(jobErr).Msg("create integration job for push order failed")
	}

	result, pushErr := a.client.PushOrder(ctx, req)
	status := "success"
	lastError := ""
	if pushErr != nil {
		status = "failed"
		lastError = pushErr.Error()
	}
	if finishErr := a.storage.FinishSyncJob(ctx, jobID, status, lastError); finishErr != nil {
		log.Error().Err(finishErr).Msg("finish integration job for push order failed")
	}
	return result, pushErr
}
