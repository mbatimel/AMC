package custom_handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	externalapi "github.com/mbatimel/AMC/users/pkg/interfaces/externalAPI"
	"github.com/rs/zerolog/log"
)

const ServiceName = "users"

func CreateUser(ctx *fiber.Ctx, svc externalapi.UsersAPI, adminUserID uuid.UUID, email string, phone string, firstName string, lastName string, middleName string, role string, status string, clientID string, companyName string, inn string, isActive bool) error {
	return handle(ctx, "post", "/v1/users", "CreateUser", map[string]interface{}{
		"adminUserID": adminUserID, "role": role, "clientID": clientID,
	}, func() (interface{}, error) {
		return svc.CreateUser(ctx.UserContext(), adminUserID, email, phone, firstName, lastName, middleName, role, status, clientID, companyName, inn, isActive)
	})
}

func GetUser(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID) error {
	return handle(ctx, "get", "/v1/users/{userID}", "GetUser", map[string]interface{}{"userID": userID}, func() (interface{}, error) {
		return svc.GetUser(ctx.UserContext(), userID)
	})
}

func ListUsers(ctx *fiber.Ctx, svc externalapi.UsersAPI, q string, role string, status string, clientID string, isActive *bool, limit int, offset int, sort string) error {
	return handle(ctx, "get", "/v1/users", "ListUsers", map[string]interface{}{
		"role": role, "status": status, "clientID": clientID, "isActive": isActive,
		"limit": limit, "offset": offset, "sort": sort, "hasQuery": q != "",
	}, func() (interface{}, error) {
		return svc.ListUsers(ctx.UserContext(), q, role, status, clientID, isActive, limit, offset, sort)
	})
}

func UpdateUser(ctx *fiber.Ctx, svc externalapi.UsersAPI, adminUserID uuid.UUID, userID uuid.UUID, email string, phone string, firstName string, lastName string, middleName string, role string, status string, clientID string, companyName string, inn string, isActive *bool) error {
	return handle(ctx, "patch", "/v1/users/{userID}", "UpdateUser", map[string]interface{}{
		"adminUserID": adminUserID, "userID": userID, "role": role, "status": status, "clientID": clientID,
	}, func() (interface{}, error) {
		return svc.UpdateUser(ctx.UserContext(), adminUserID, userID, email, phone, firstName, lastName, middleName, role, status, clientID, companyName, inn, isActive)
	})
}

func DeleteUser(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID) error {
	return handle(ctx, "delete", "/v1/users/{userID}", "DeleteUser", map[string]interface{}{"userID": userID}, func() (interface{}, error) {
		return svc.DeleteUser(ctx.UserContext(), userID)
	})
}

func ActivateUser(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID) error {
	return handle(ctx, "post", "/v1/users/{userID}/activate", "ActivateUser", map[string]interface{}{"userID": userID}, func() (interface{}, error) {
		return svc.ActivateUser(ctx.UserContext(), userID)
	})
}

func DeactivateUser(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID) error {
	return handle(ctx, "post", "/v1/users/{userID}/deactivate", "DeactivateUser", map[string]interface{}{"userID": userID}, func() (interface{}, error) {
		return svc.DeactivateUser(ctx.UserContext(), userID)
	})
}

func GetProfile(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID) error {
	return handle(ctx, "get", "/v1/profile", "GetProfile", map[string]interface{}{"userID": userID}, func() (interface{}, error) {
		return svc.GetProfile(ctx.UserContext(), userID)
	})
}

func UpdateProfile(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID, email string, phone string, firstName string, lastName string, middleName string) error {
	return handle(ctx, "patch", "/v1/profile", "UpdateProfile", map[string]interface{}{"userID": userID}, func() (interface{}, error) {
		return svc.UpdateProfile(ctx.UserContext(), userID, email, phone, firstName, lastName, middleName)
	})
}

func ListUserClients(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID) error {
	return handle(ctx, "get", "/v1/profile/clients", "ListUserClients", map[string]interface{}{"userID": userID}, func() (interface{}, error) {
		return svc.ListUserClients(ctx.UserContext(), userID)
	})
}

func GetClientDetails(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID, clientID uuid.UUID) error {
	return handle(ctx, "get", "/v1/profile/clients/{clientID}", "GetClientDetails", map[string]interface{}{
		"userID": userID, "clientID": clientID,
	}, func() (interface{}, error) {
		return svc.GetClientDetails(ctx.UserContext(), userID, clientID)
	})
}

func GetClientConditions(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID, clientID uuid.UUID) error {
	return handle(ctx, "get", "/v1/profile/clients/{clientID}/conditions", "GetClientConditions", map[string]interface{}{
		"userID": userID, "clientID": clientID,
	}, func() (interface{}, error) {
		return svc.GetClientConditions(ctx.UserContext(), userID, clientID)
	})
}

func SwitchActiveClient(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID, clientID uuid.UUID) error {
	return handle(ctx, "post", "/v1/profile/clients/{clientID}/activate", "SwitchActiveClient", map[string]interface{}{
		"userID": userID, "clientID": clientID,
	}, func() (interface{}, error) {
		return svc.SwitchActiveClient(ctx.UserContext(), userID, clientID)
	})
}

func ListFavorites(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID) error {
	return handle(ctx, "get", "/v1/favorites", "ListFavorites", map[string]interface{}{"userID": userID}, func() (interface{}, error) {
		return svc.ListFavorites(ctx.UserContext(), userID)
	})
}

func AddFavorite(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID, productID string) error {
	return handle(ctx, "post", "/v1/favorites", "AddFavorite", map[string]interface{}{
		"userID": userID, "productID": productID,
	}, func() (interface{}, error) {
		return svc.AddFavorite(ctx.UserContext(), userID, productID)
	})
}

func DeleteFavorites(ctx *fiber.Ctx, svc externalapi.UsersAPI, userID uuid.UUID, productIDs []string) error {
	return handle(ctx, "delete", "/v1/favorites", "DeleteFavorites", map[string]interface{}{
		"userID": userID, "count": len(productIDs),
	}, func() (interface{}, error) {
		return svc.DeleteFavorites(ctx.UserContext(), userID, productIDs)
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
		event := log.Info()
		if err != nil {
			event = log.Error().Err(err)
		}
		event.Fields(fields).Msg("call")
	}(time.Now())

	data, err := call()
	if err != nil {
		sendResponse(ctx, log.Logger, nil, err)
		return nil
	}

	sendResponse(ctx, log.Logger, data, nil)
	return nil
}
