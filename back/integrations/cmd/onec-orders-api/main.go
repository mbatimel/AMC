package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"

	"github.com/mbatimel/AMC/integrations/internal/config"
	customErrors "github.com/mbatimel/AMC/integrations/internal/errors"
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

	// The tg-generated OnecOrdersAPI service is deliberately NOT registered
	// here (no internalapi.OnecOrdersAPI(...) option). tg maps this method's
	// `http-args` annotations to query-string parameters and wraps responses in
	// the RestResponse envelope, neither of which matches the 1С contract
	// (JSON body, bare {"ok": true}) — see docs/superpowers/specs/
	// 2026-08-26-onec-orders-integration-1c-contract.md. internalapi.New is
	// still used for the *fiber.App and its middleware (recover, request
	// logging, metrics); WithLog/WithMetrics nil-guard the absent service.
	// The route itself is hand-written below, same escape hatch as PushOrder.
	app := internalapi.New(log.Logger).WithLog().WithMetrics()
	transportHTTP.RegisterWebhookRoute(app.Fiber(), webhookSvc, log.Logger)
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
	if err != nil {
		return "", mapOrdersClientError(err)
	}
	return resp.Status, nil
}

func (a ordersClientAdapter) UpdateOrderStatus(ctx context.Context, userID, orderID uuid.UUID, status, paymentStatus, comment, changedBy string) error {
	if _, err := a.cli.UpdateOrderStatus(ctx, userID, orderID, status, paymentStatus, comment, changedBy); err != nil {
		return mapOrdersClientError(err)
	}
	return nil
}

// mapOrdersClientError turns the generated orders client's untyped error into a
// typed *customErrors.Error so the 1С webhook can answer with the status code
// the contract promises (notably 404 for an unknown client_order_id, which 1С
// must NOT retry — everything used to collapse to 500, which it does retry).
//
// The generated client (orders/pkg/client/transport/ordersapi-http-client.go,
// "GENERATED — DO NOT EDIT") reports every non-200 response as a plain
// fmt.Errorf("HTTP error: %d. URL: %s, ...") with no typed carrier, so parsing
// that deterministic prefix is the only way to recover the upstream status.
// If the prefix ever changes, Sscanf fails and we fall back to 500 — the safe
// default (retryable, and honest about "we don't know").
func mapOrdersClientError(err error) error {
	if err == nil {
		return nil
	}

	var statusCode int
	if _, scanErr := fmt.Sscanf(err.Error(), "HTTP error: %d.", &statusCode); scanErr != nil {
		statusCode = 0
	}

	switch statusCode {
	case http.StatusNotFound:
		return customErrors.NotFoundError().SetOuterError(err)
	case http.StatusUnauthorized, http.StatusForbidden:
		return customErrors.UnauthorizedError().SetOuterError(err)
	default:
		return customErrors.InternalServerError().SetOuterError(err)
	}
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
