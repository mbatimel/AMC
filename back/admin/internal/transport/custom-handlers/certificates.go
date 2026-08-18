// back/admin/internal/transport/custom-handlers/certificates.go
package custom_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	externalapi "github.com/mbatimel/AMC/admin/pkg/interfaces/externalapi"
)

func ListCertificates(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID) error {
	return handle(ctx, "get", "/v1/admin/certificates", "ListCertificates", map[string]interface{}{
		"userID": userID,
	}, func() (interface{}, error) {
		return svc.ListCertificates(ctx.UserContext(), userID)
	})
}

func ListPublicCertificates(ctx *fiber.Ctx, svc externalapi.AdminAPI) error {
	return handle(ctx, "get", "/v1/certificates", "ListPublicCertificates", map[string]interface{}{}, func() (interface{}, error) {
		return svc.ListPublicCertificates(ctx.UserContext())
	})
}

func CreateCertificate(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID, title string, sortOrder int, isActive bool, fileName string, fileContentBase64 string) error {
	return handle(ctx, "post", "/v1/admin/certificates", "CreateCertificate", map[string]interface{}{
		"userID": userID,
		"title":  title,
	}, func() (interface{}, error) {
		return svc.CreateCertificate(ctx.UserContext(), userID, title, sortOrder, isActive, fileName, fileContentBase64)
	})
}

func UpdateCertificate(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID, certID uuid.UUID, title string, sortOrder int, isActive bool, fileName string, fileContentBase64 string) error {
	return handle(ctx, "patch", "/v1/admin/certificates/:certID", "UpdateCertificate", map[string]interface{}{
		"userID": userID,
		"certID": certID,
	}, func() (interface{}, error) {
		return svc.UpdateCertificate(ctx.UserContext(), userID, certID, title, sortOrder, isActive, fileName, fileContentBase64)
	})
}

func DeleteCertificate(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID, certID uuid.UUID) error {
	return handle(ctx, "delete", "/v1/admin/certificates/:certID", "DeleteCertificate", map[string]interface{}{
		"userID": userID,
		"certID": certID,
	}, func() (interface{}, error) {
		return svc.DeleteCertificate(ctx.UserContext(), userID, certID)
	})
}
