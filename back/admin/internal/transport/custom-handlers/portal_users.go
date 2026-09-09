package custom_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	externalapi "github.com/mbatimel/AMC/admin/pkg/interfaces/externalapi"
)

func InviteAdmin(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID, email string, name string) error {
	return handle(ctx, "post", "/v1/admin/portal-users/invite", "InviteAdmin", map[string]interface{}{
		"userID": userID,
		"email":  email,
		"name":   name,
	}, func() (interface{}, error) {
		return svc.InviteAdmin(ctx.UserContext(), userID, email, name)
	})
}
