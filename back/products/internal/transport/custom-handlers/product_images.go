package custom_handlers

import (
	"context"
	"io"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/products/internal/errors"
	"github.com/mbatimel/AMC/products/pkg/models"
)

const MaxBatchImageFiles = 100

type ProductImageRoutesService interface {
	UploadProductImage(ctx context.Context, userID uuid.UUID, productID uuid.UUID, file models.ImageFile) (models.UploadProductImageResponse, error)
	UploadProductImagesBatch(ctx context.Context, userID uuid.UUID, files []models.ImageFile) (models.UploadProductImagesBatchResponse, error)
	DeleteProductImage(ctx context.Context, userID uuid.UUID, productID uuid.UUID, imageID uuid.UUID) (models.DeleteProductImageResponse, error)
}

type ProductImageRoutes struct {
	service     ProductImageRoutesService
	logger      zerolog.Logger
	maxFileSize int64
}

func NewProductImageRoutes(logger zerolog.Logger, service ProductImageRoutesService, maxFileSize int64) *ProductImageRoutes {
	return &ProductImageRoutes{logger: logger, service: service, maxFileSize: maxFileSize}
}

func (r *ProductImageRoutes) SetRoutes(app *fiber.App) {
	app.Post("/api/v1/products/:productID/images", r.uploadProductImage)
	app.Delete("/api/v1/products/:productID/images/:imageID", r.deleteProductImage)
	app.Post("/api/v1/products/images/batch", r.uploadProductImagesBatch)
}

func userIDFromHeader(ctx *fiber.Ctx) uuid.UUID {
	userID, _ := uuid.Parse(ctx.Get("X-User-Id"))
	return userID
}

func readMultipartFile(header *multipart.FileHeader, maxFileSize int64) (models.ImageFile, error) {
	if header == nil || header.Size <= 0 || header.Size > maxFileSize {
		return models.ImageFile{}, customErrors.ErrValidation.AddCause("field", "fileSize")
	}
	file, err := header.Open()
	if err != nil {
		return models.ImageFile{}, customErrors.ErrValidation.AddCause("field", "file")
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, maxFileSize+1))
	if err != nil || int64(len(content)) > maxFileSize {
		return models.ImageFile{}, customErrors.ErrValidation.AddCause("field", "fileSize")
	}
	return models.ImageFile{FileName: header.Filename, Content: content}, nil
}

func (r *ProductImageRoutes) uploadProductImage(ctx *fiber.Ctx) error {
	productID, _ := uuid.Parse(ctx.Params("productID"))
	header, err := ctx.FormFile("file")
	if err != nil {
		sendResponse(ctx, r.logger, nil, customErrors.ErrValidation.AddCause("field", "file"))
		return nil
	}
	file, err := readMultipartFile(header, r.maxFileSize)
	if err != nil {
		sendResponse(ctx, r.logger, nil, err)
		return nil
	}
	response, err := r.service.UploadProductImage(ctx.UserContext(), userIDFromHeader(ctx), productID, file)
	sendResponse(ctx, r.logger, response, err)
	return nil
}

func (r *ProductImageRoutes) uploadProductImagesBatch(ctx *fiber.Ctx) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		sendResponse(ctx, r.logger, nil, customErrors.ErrValidation.AddCause("field", "files"))
		return nil
	}
	headers := form.File["files"]
	if len(headers) == 0 || len(headers) > MaxBatchImageFiles {
		sendResponse(ctx, r.logger, nil, customErrors.ErrValidation.AddCause("field", "files"))
		return nil
	}
	files := make([]models.ImageFile, 0, len(headers))
	for _, header := range headers {
		file, readErr := readMultipartFile(header, r.maxFileSize)
		if readErr != nil {
			// Preserve partial processing: the service will report validation for
			// this item while independently handling all readable files.
			files = append(files, models.ImageFile{FileName: header.Filename})
			continue
		}
		files = append(files, file)
	}
	response, err := r.service.UploadProductImagesBatch(ctx.UserContext(), userIDFromHeader(ctx), files)
	sendResponse(ctx, r.logger, response, err)
	return nil
}

func (r *ProductImageRoutes) deleteProductImage(ctx *fiber.Ctx) error {
	productID, _ := uuid.Parse(ctx.Params("productID"))
	imageID, _ := uuid.Parse(ctx.Params("imageID"))
	response, err := r.service.DeleteProductImage(ctx.UserContext(), userIDFromHeader(ctx), productID, imageID)
	sendResponse(ctx, r.logger, response, err)
	return nil
}
