package custom_handlers

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	customErrors "github.com/mbatimel/AMC/auth/internal/errors"
	"github.com/mbatimel/AMC/auth/internal/service"
)

// RegisterIPService is implemented by internal/service.service (via internal/service.NewAuthApiService).
type RegisterIPService interface {
	RegisterIP(
		ctx context.Context,
		email, password, shortName, inn, directorFullName, phone string,
		file service.RequisitesFile,
	) (uuid.UUID, error)
}

// RegisterIPRoutes serves POST /api/v1/auth/register/ip as multipart/form-data,
// replacing the tg-generated JSON route for the same path.
type RegisterIPRoutes struct {
	service     RegisterIPService
	maxFileSize int64
}

func NewRegisterIPRoutes(svc RegisterIPService, maxFileSize int64) *RegisterIPRoutes {
	return &RegisterIPRoutes{service: svc, maxFileSize: maxFileSize}
}

func (r *RegisterIPRoutes) SetRoutes(app *fiber.App) {
	app.Post("/api/v1/auth/register/ip", r.registerIP)
}

func readRequisitesFile(header *multipart.FileHeader, maxFileSize int64) (service.RequisitesFile, error) {
	if header == nil || header.Size <= 0 {
		return service.RequisitesFile{}, customErrors.RequisitesFileRequiredError()
	}
	if header.Size > maxFileSize {
		return service.RequisitesFile{}, customErrors.RequisitesFileTooLargeError()
	}
	file, err := header.Open()
	if err != nil {
		return service.RequisitesFile{}, customErrors.RequisitesFileRequiredError()
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, maxFileSize+1))
	if err != nil || int64(len(content)) > maxFileSize {
		return service.RequisitesFile{}, customErrors.RequisitesFileTooLargeError()
	}
	return service.RequisitesFile{FileName: header.Filename, Content: content}, nil
}

func writeRegisterIPResponse(ctx *fiber.Ctx, data interface{}, respErr error) {
	ctx.Response().Header.SetContentType("application/json")
	if respErr != nil {
		status := fiber.StatusInternalServerError
		if coder, ok := respErr.(interface{ Code() int }); ok {
			if c := coder.Code(); c != 0 {
				status = c
			}
		}
		ctx.Status(status)
		_ = json.NewEncoder(ctx).Encode(respErr)
		return
	}
	ctx.Status(fiber.StatusOK)
	_ = json.NewEncoder(ctx).Encode(data)
}

func (r *RegisterIPRoutes) registerIP(ctx *fiber.Ctx) error {
	fields := map[string]string{
		"email":            ctx.FormValue("email"),
		"password":         ctx.FormValue("password"),
		"shortName":        ctx.FormValue("shortName"),
		"inn":              ctx.FormValue("inn"),
		"directorFullName": ctx.FormValue("directorFullName"),
		"phone":            ctx.FormValue("phone"),
	}
	for _, field := range []string{"email", "password", "shortName", "inn", "directorFullName", "phone"} {
		if strings.TrimSpace(fields[field]) == "" {
			writeRegisterIPResponse(ctx, nil, customErrors.ValidationError(field))
			return nil
		}
	}

	header, err := ctx.FormFile("requisitesFile")
	if err != nil {
		writeRegisterIPResponse(ctx, nil, customErrors.RequisitesFileRequiredError())
		return nil
	}
	file, err := readRequisitesFile(header, r.maxFileSize)
	if err != nil {
		writeRegisterIPResponse(ctx, nil, err)
		return nil
	}

	userID, err := r.service.RegisterIP(
		ctx.UserContext(),
		fields["email"], fields["password"], fields["shortName"], fields["inn"],
		fields["directorFullName"], fields["phone"], file,
	)
	if err != nil {
		writeRegisterIPResponse(ctx, nil, err)
		return nil
	}

	writeRegisterIPResponse(ctx, struct {
		UserID uuid.UUID `json:"userID,omitempty"`
	}{userID}, nil)
	return nil
}
