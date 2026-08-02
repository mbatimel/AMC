package custom_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	externalapi "github.com/mbatimel/AMC/products/pkg/interfaces/externalapi"
	"github.com/mbatimel/AMC/products/pkg/models"
)

func CreatePromotion(ctx *fiber.Ctx, svc externalapi.ProductsAPI, userID uuid.UUID, name string, discountPercent float64, startsAt string, endsAt string, products []models.PromotionProduct) error {
	return handle(ctx, "post", "/v1/promotions", "CreatePromotion", map[string]interface{}{
		"userID": userID, "name": name,
	}, func() (interface{}, error) {
		return svc.CreatePromotion(ctx.UserContext(), userID, name, discountPercent, startsAt, endsAt, products)
	})
}

func GetPromotion(ctx *fiber.Ctx, svc externalapi.ProductsAPI, promotionID uuid.UUID) error {
	return handle(ctx, "get", "/v1/promotions/{promotionID}", "GetPromotion", map[string]interface{}{
		"promotionID": promotionID,
	}, func() (interface{}, error) {
		return svc.GetPromotion(ctx.UserContext(), promotionID)
	})
}

func ListPromotions(ctx *fiber.Ctx, svc externalapi.ProductsAPI, limit *int, offset *int) error {
	return handle(ctx, "get", "/v1/promotions", "ListPromotions", map[string]interface{}{
		"limit": limit, "offset": offset,
	}, func() (interface{}, error) {
		return svc.ListPromotions(ctx.UserContext(), limit, offset)
	})
}

func UpdatePromotion(ctx *fiber.Ctx, svc externalapi.ProductsAPI, userID uuid.UUID, promotionID uuid.UUID, name string, discountPercent float64, startsAt string, endsAt string, products []models.PromotionProduct) error {
	return handle(ctx, "patch", "/v1/promotions/{promotionID}", "UpdatePromotion", map[string]interface{}{
		"userID": userID, "promotionID": promotionID,
	}, func() (interface{}, error) {
		return svc.UpdatePromotion(ctx.UserContext(), userID, promotionID, name, discountPercent, startsAt, endsAt, products)
	})
}

func DeletePromotion(ctx *fiber.Ctx, svc externalapi.ProductsAPI, userID uuid.UUID, promotionID uuid.UUID) error {
	return handle(ctx, "delete", "/v1/promotions/{promotionID}", "DeletePromotion", map[string]interface{}{
		"userID": userID, "promotionID": promotionID,
	}, func() (interface{}, error) {
		return svc.DeletePromotion(ctx.UserContext(), userID, promotionID)
	})
}
