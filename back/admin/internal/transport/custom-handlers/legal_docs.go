// back/admin/internal/transport/custom-handlers/legal_docs.go
package custom_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	externalapi "github.com/mbatimel/AMC/admin/pkg/interfaces/externalapi"
)

func ListLegalDocs(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID) error {
	return handle(ctx, "get", "/v1/admin/legal-docs", "ListLegalDocs", map[string]interface{}{
		"userID": userID,
	}, func() (interface{}, error) {
		return svc.ListLegalDocs(ctx.UserContext(), userID)
	})
}

func ListPublicLegalDocs(ctx *fiber.Ctx, svc externalapi.AdminAPI) error {
	return handle(ctx, "get", "/v1/legal-docs", "ListPublicLegalDocs", map[string]interface{}{}, func() (interface{}, error) {
		return svc.ListPublicLegalDocs(ctx.UserContext())
	})
}

func CreateLegalDoc(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID, id string, name string, version string, summary string, fileName string, fileContentBase64 string) error {
	return handle(ctx, "post", "/v1/admin/legal-docs", "CreateLegalDoc", map[string]interface{}{
		"userID": userID,
		"id":     id,
	}, func() (interface{}, error) {
		return svc.CreateLegalDoc(ctx.UserContext(), userID, id, name, version, summary, fileName, fileContentBase64)
	})
}

func ReplaceLegalDocFile(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID, docID string, version string, summary string, fileName string, fileContentBase64 string) error {
	return handle(ctx, "patch", "/v1/admin/legal-docs/:docID", "ReplaceLegalDocFile", map[string]interface{}{
		"userID": userID,
		"docID":  docID,
	}, func() (interface{}, error) {
		return svc.ReplaceLegalDocFile(ctx.UserContext(), userID, docID, version, summary, fileName, fileContentBase64)
	})
}

func DeleteLegalDoc(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID, docID string) error {
	return handle(ctx, "delete", "/v1/admin/legal-docs/:docID", "DeleteLegalDoc", map[string]interface{}{
		"userID": userID,
		"docID":  docID,
	}, func() (interface{}, error) {
		return svc.DeleteLegalDoc(ctx.UserContext(), userID, docID)
	})
}

func ListLegalDocVersions(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID, docID string) error {
	return handle(ctx, "get", "/v1/admin/legal-docs/:docID/versions", "ListLegalDocVersions", map[string]interface{}{
		"userID": userID,
		"docID":  docID,
	}, func() (interface{}, error) {
		return svc.ListLegalDocVersions(ctx.UserContext(), userID, docID)
	})
}
