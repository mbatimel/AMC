package custom_handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	externalapi "github.com/mbatimel/AMC/users/pkg/interfaces/externalAPI"
	"github.com/rs/zerolog/log"
)

const ServiceName = "users"

func CreateUser(ctx *fiber.Ctx, svc externalapi.UsersAPI, email string, phone string, firstName string, lastName string, middleName string, role string, status string, clientID string, companyName string, inn string, isActive bool) error {
	return handle(ctx, "post", "/v1/users", "CreateUser", map[string]interface{}{
		"email":    email,
		"role":     role,
		"status":   status,
		"clientID": clientID,
	}, func() (interface{}, error) {
		return svc.CreateUser(ctx.UserContext(), email, phone, firstName, lastName, middleName, role, status, clientID, companyName, inn, isActive)
	})
}

func GetUser(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID) error {
	return handle(ctx, "get", "/v1/users/{userID}", "GetUser", map[string]interface{}{
		"userID": userID,
	}, func() (interface{}, error) {
		return svc.GetUser(ctx.UserContext(), userID)
	})
}

func ListUsers(ctx *fiber.Ctx, svc externalapi.UsersAPI, q string, role string, status string, clientID string, isActive bool, limit int, offset int, sort string) error {
	return handle(ctx, "get", "/v1/users", "ListUsers", map[string]interface{}{
		"q":        q,
		"role":     role,
		"status":   status,
		"clientID": clientID,
		"isActive": isActive,
		"limit":    limit,
		"offset":   offset,
		"sort":     sort,
	}, func() (interface{}, error) {
		return svc.ListUsers(ctx.UserContext(), q, role, status, clientID, isActive, limit, offset, sort)
	})
}

func UpdateUser(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID, email string, phone string, firstName string, lastName string, middleName string, role string, status string, clientID string, companyName string, inn string, isActive bool) error {
	return handle(ctx, "patch", "/v1/users/{userID}", "UpdateUser", map[string]interface{}{
		"userID":   userID,
		"email":    email,
		"role":     role,
		"status":   status,
		"clientID": clientID,
	}, func() (interface{}, error) {
		return svc.UpdateUser(ctx.UserContext(), userID, email, phone, firstName, lastName, middleName, role, status, clientID, companyName, inn, isActive)
	})
}

func DeleteUser(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID) error {
	return handle(ctx, "delete", "/v1/users/{userID}", "DeleteUser", map[string]interface{}{
		"userID": userID,
	}, func() (interface{}, error) {
		return svc.DeleteUser(ctx.UserContext(), userID)
	})
}

func ActivateUser(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID) error {
	return handle(ctx, "post", "/v1/users/{userID}/activate", "ActivateUser", map[string]interface{}{
		"userID": userID,
	}, func() (interface{}, error) {
		return svc.ActivateUser(ctx.UserContext(), userID)
	})
}

func DeactivateUser(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID) error {
	return handle(ctx, "post", "/v1/users/{userID}/deactivate", "DeactivateUser", map[string]interface{}{
		"userID": userID,
	}, func() (interface{}, error) {
		return svc.DeactivateUser(ctx.UserContext(), userID)
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
