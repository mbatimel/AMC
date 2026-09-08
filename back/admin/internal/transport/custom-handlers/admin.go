// back/admin/internal/transport/custom-handlers/admin.go
package custom_handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	externalapi "github.com/mbatimel/AMC/admin/pkg/interfaces/externalapi"
	"github.com/rs/zerolog/log"
)

const ServiceName = "admin"

func Login(ctx *fiber.Ctx, svc externalapi.AdminAPI, email string, password string) error {
	return handle(ctx, "post", "/v1/admin/auth/login", "Login", map[string]interface{}{
		"email": email,
	}, func() (interface{}, error) {
		return svc.Login(ctx.UserContext(), email, password)
	})
}

func Logout(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID) error {
	return handle(ctx, "post", "/v1/admin/auth/logout", "Logout", map[string]interface{}{
		"userID": userID,
	}, func() (interface{}, error) {
		return nil, svc.Logout(ctx.UserContext(), userID)
	})
}

func GetSession(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID) error {
	return handle(ctx, "get", "/v1/admin/auth/session", "GetSession", map[string]interface{}{
		"userID": userID,
	}, func() (interface{}, error) {
		return svc.GetSession(ctx.UserContext(), userID)
	})
}

func ListAuditLog(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID, limit int, offset int) error {
	return handle(ctx, "get", "/v1/admin/audit-log", "ListAuditLog", map[string]interface{}{
		"userID": userID,
		"limit":  limit,
		"offset": offset,
	}, func() (interface{}, error) {
		return svc.ListAuditLog(ctx.UserContext(), userID, limit, offset)
	})
}

func handle(
	ctx *fiber.Ctx,
	method string,
	path string,
	methodName string,
	fields map[string]interface{},
	call func() (interface{}, error),
) error {
	var err error

	defer func(begin time.Time) {
		fields["method"] = method
		fields["path"] = path
		fields["methodName"] = methodName
		fields["serviceName"] = ServiceName
		fields["took"] = time.Since(begin).String()

		l := log.Info()
		if err != nil {
			l = log.Error().Err(err)
		}

		l.Fields(fields).Msg("call")
	}(time.Now())

	data, err := call()
	if err != nil {
		sendResponse(ctx, log.Logger, nil, err)
		return nil
	}

	sendResponse(ctx, log.Logger, data, nil)
	return nil
}
