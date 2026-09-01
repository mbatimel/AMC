package custom_handlers

import (
	"context"
	"io"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

type CompanyRequestService interface {
	SendCompanyRequest(ctx context.Context, input models.CompanyRequestInput) (models.CompanyRequestResponse, error)
}

type CompanyRequestRoutes struct {
	service      CompanyRequestService
	logger       zerolog.Logger
	maxFileSize  int64
	maxFileCount int
}

func NewCompanyRequestRoutes(logger zerolog.Logger, service CompanyRequestService, maxFileSize int64, maxFileCount int) *CompanyRequestRoutes {
	return &CompanyRequestRoutes{
		logger: logger, service: service, maxFileSize: maxFileSize, maxFileCount: maxFileCount,
	}
}

func (r *CompanyRequestRoutes) SetRoutes(app *fiber.App) {
	app.Post("/api/v1/company-requests", r.create)
}

func readCompanyAttachment(header *multipart.FileHeader, maxFileSize int64) (models.ImageFile, error) {
	if header == nil || header.Size <= 0 || header.Size > maxFileSize {
		return models.ImageFile{}, customErrors.BadRequestError().AddCause("field", "attachments.fileSize")
	}
	file, err := header.Open()
	if err != nil {
		return models.ImageFile{}, customErrors.BadRequestError().AddCause("field", "attachments.file")
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, maxFileSize+1))
	if err != nil || int64(len(content)) > maxFileSize {
		return models.ImageFile{}, customErrors.BadRequestError().AddCause("field", "attachments.fileSize")
	}
	return models.ImageFile{FileName: header.Filename, Content: content}, nil
}

func (r *CompanyRequestRoutes) create(ctx *fiber.Ctx) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		sendResponse(ctx, r.logger, nil, customErrors.BadRequestError().AddCause("field", "form"))
		return nil
	}
	defer form.RemoveAll()
	headers := form.File["attachments"]
	if len(headers) > r.maxFileCount {
		sendResponse(ctx, r.logger, nil, customErrors.BadRequestError().AddCause("field", "attachmentsCount"))
		return nil
	}
	attachments := make([]models.ImageFile, 0, len(headers))
	var totalSize int64
	for _, header := range headers {
		totalSize += header.Size
		if totalSize > r.maxFileSize {
			sendResponse(ctx, r.logger, nil, customErrors.BadRequestError().AddCause("field", "attachmentsTotalSize"))
			return nil
		}
		attachment, readErr := readCompanyAttachment(header, r.maxFileSize)
		if readErr != nil {
			sendResponse(ctx, r.logger, nil, readErr)
			return nil
		}
		attachments = append(attachments, attachment)
	}

	response, serviceErr := r.service.SendCompanyRequest(ctx.UserContext(), models.CompanyRequestInput{
		ContactName: ctx.FormValue("contact_name"),
		Email:       ctx.FormValue("email"),
		Phone:       ctx.FormValue("phone"),
		Company:     ctx.FormValue("company"),
		Message:     ctx.FormValue("message"),
		Attachments: attachments,
	})
	sendResponse(ctx, r.logger, response, serviceErr)
	return nil
}
