// Package internalapi describes the public 1С→AMC order-status webhook.
// @tg version=0.0.1
// @tg backend=integrations
// @tg title=`onec-orders-api`
// @tg servers=
//
//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/internalapi --outSwagger ../../../swaggers/internalapi/swagger.yaml
package internalAPI

import (
	"context"

	"github.com/google/uuid"
)

// OnecOrdersAPI
//
// NOTE: the generated HTTP route for OnecOrderStatusWebhook is deliberately NOT
// wired into onec-orders-api. tg maps `http-args` to QUERY STRING parameters and
// wraps responses in the RestResponse envelope, but the 1С contract specifies a
// JSON body and a bare {"ok": true} response. The endpoint is hand-written in
// internal/transport/http.RegisterWebhookRoute — change that when the contract
// changes. These annotations (and the generated swagger) are kept as
// documentation of intent and for potential `tg client` consumers only.
//
// @tg http-server metrics log
// @tg http-prefix=/api
// @tg 200=github.com/mbatimel/AMC/integrations/swaggers/internalapi/models:Resp200
// @tg 400=github.com/mbatimel/AMC/integrations/swaggers/internalapi/models:Err400
// @tg 401=github.com/mbatimel/AMC/integrations/swaggers/internalapi/models:Err401
// @tg 404=github.com/mbatimel/AMC/integrations/swaggers/internalapi/models:Err404
// @tg 409=github.com/mbatimel/AMC/integrations/swaggers/internalapi/models:Err409
// @tg 500=github.com/mbatimel/AMC/integrations/swaggers/internalapi/models:Err500
type OnecOrdersAPI interface {
	// OnecOrderStatusWebhook is called by 1С when an order reaches a new
	// stage it needs to report (v1: only "delivered").
	// @tg http-method=POST
	// @tg http-path=/v1/onec/orders/status
	// @tg http-headers=apiKey|X-Onec-Api-Key
	// @tg http-args=clientOrderID|clientOrderID
	// @tg http-args=status|status
	// @tg http-args=onecDocumentNumber|onecDocumentNumber
	// @tg http-args=comment|comment
	// @tg http-response=github.com/mbatimel/AMC/integrations/internal/transport/custom-handlers:OnecOrderStatusWebhook
	// @tg summary=`Статус заказа от 1С`
	// @tg desc=`Приём вебхука 1С о смене стадии заказа; в v1 допустим только status=delivered`
	// @tg uuidPackage=github.com/google/uuid
	OnecOrderStatusWebhook(ctx context.Context, apiKey string, clientOrderID uuid.UUID, status string, onecDocumentNumber string, comment string) (ok bool, err error)
}
