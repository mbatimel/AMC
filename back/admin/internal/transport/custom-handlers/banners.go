package custom_handlers

import (
	"context"
	"io"
	"mime/multipart"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

type BannerRoutesService interface {
	ListBanners(ctx context.Context, userID uuid.UUID) (models.BannersSettings, error)
	ListPublicBanners(ctx context.Context) (models.BannersSettings, error)
	CreateBanner(ctx context.Context, userID uuid.UUID, input models.BannerInput) (models.Banner, error)
	UpdateBanner(ctx context.Context, userID uuid.UUID, bannerID uuid.UUID, input models.BannerInput) (models.Banner, error)
	DeleteBanner(ctx context.Context, userID uuid.UUID, bannerID uuid.UUID) (models.DeleteBannerResponse, error)
	UpdateBannerDelay(ctx context.Context, userID uuid.UUID, delaySec int) (models.BannersSettings, error)
}

type BannerRoutes struct {
	service     BannerRoutesService
	logger      zerolog.Logger
	maxFileSize int64
}

func NewBannerRoutes(logger zerolog.Logger, service BannerRoutesService, maxFileSize int64) *BannerRoutes {
	return &BannerRoutes{logger: logger, service: service, maxFileSize: maxFileSize}
}

func (r *BannerRoutes) SetRoutes(app *fiber.App) {
	app.Get("/api/v1/banners", r.listPublicBanners)
	app.Get("/api/v1/admin/banners", r.listBanners)
	app.Post("/api/v1/admin/banners", r.createBanner)
	app.Patch("/api/v1/admin/banners/:bannerID", r.updateBanner)
	app.Delete("/api/v1/admin/banners/:bannerID", r.deleteBanner)
	app.Put("/api/v1/admin/banners/settings", r.updateBannerDelay)
}

func bannerUserID(ctx *fiber.Ctx) uuid.UUID {
	userID, _ := uuid.Parse(ctx.Get("X-User-Id"))
	return userID
}

func readBannerFile(header *multipart.FileHeader, maxFileSize int64) (*models.ImageFile, error) {
	if header == nil || header.Size <= 0 || header.Size > maxFileSize {
		return nil, customErrors.BadRequestError().AddCause("field", "fileSize")
	}
	file, err := header.Open()
	if err != nil {
		return nil, customErrors.BadRequestError().AddCause("field", "file")
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, maxFileSize+1))
	if err != nil || int64(len(content)) > maxFileSize {
		return nil, customErrors.BadRequestError().AddCause("field", "fileSize")
	}
	return &models.ImageFile{FileName: header.Filename, Content: content}, nil
}

func bannerInput(ctx *fiber.Ctx, fileRequired bool, maxFileSize int64) (models.BannerInput, error) {
	sortOrder, err := strconv.Atoi(ctx.FormValue("sort_order", "0"))
	if err != nil {
		return models.BannerInput{}, customErrors.BadRequestError().AddCause("field", "sortOrder")
	}
	isActive, err := strconv.ParseBool(ctx.FormValue("is_active", "true"))
	if err != nil {
		return models.BannerInput{}, customErrors.BadRequestError().AddCause("field", "isActive")
	}
	input := models.BannerInput{
		Title: ctx.FormValue("title"), Subtitle: ctx.FormValue("subtitle"),
		Link: ctx.FormValue("link"), SortOrder: sortOrder, IsActive: isActive,
		DateFrom: ctx.FormValue("dateFrom"), DateTo: ctx.FormValue("dateTo"),
	}
	header, fileErr := ctx.FormFile("file")
	if fileErr == nil {
		input.Image, err = readBannerFile(header, maxFileSize)
		if err != nil {
			return models.BannerInput{}, err
		}
	} else if fileRequired {
		return models.BannerInput{}, customErrors.BadRequestError().AddCause("field", "file")
	}
	return input, nil
}

func (r *BannerRoutes) listPublicBanners(ctx *fiber.Ctx) error {
	response, err := r.service.ListPublicBanners(ctx.UserContext())
	sendResponse(ctx, r.logger, response, err)
	return nil
}

func (r *BannerRoutes) listBanners(ctx *fiber.Ctx) error {
	response, err := r.service.ListBanners(ctx.UserContext(), bannerUserID(ctx))
	sendResponse(ctx, r.logger, response, err)
	return nil
}

func (r *BannerRoutes) createBanner(ctx *fiber.Ctx) error {
	input, err := bannerInput(ctx, true, r.maxFileSize)
	if err != nil {
		sendResponse(ctx, r.logger, nil, err)
		return nil
	}
	response, err := r.service.CreateBanner(ctx.UserContext(), bannerUserID(ctx), input)
	sendResponse(ctx, r.logger, response, err)
	return nil
}

func (r *BannerRoutes) updateBanner(ctx *fiber.Ctx) error {
	bannerID, _ := uuid.Parse(ctx.Params("bannerID"))
	input, err := bannerInput(ctx, false, r.maxFileSize)
	if err != nil {
		sendResponse(ctx, r.logger, nil, err)
		return nil
	}
	response, err := r.service.UpdateBanner(ctx.UserContext(), bannerUserID(ctx), bannerID, input)
	sendResponse(ctx, r.logger, response, err)
	return nil
}

func (r *BannerRoutes) deleteBanner(ctx *fiber.Ctx) error {
	bannerID, _ := uuid.Parse(ctx.Params("bannerID"))
	response, err := r.service.DeleteBanner(ctx.UserContext(), bannerUserID(ctx), bannerID)
	sendResponse(ctx, r.logger, response, err)
	return nil
}

func (r *BannerRoutes) updateBannerDelay(ctx *fiber.Ctx) error {
	request := struct {
		DelaySec int `json:"delay_sec"`
	}{}
	if err := ctx.BodyParser(&request); err != nil {
		sendResponse(ctx, r.logger, nil, customErrors.BadRequestError().AddCause("field", "delaySec"))
		return nil
	}
	response, err := r.service.UpdateBannerDelay(ctx.UserContext(), bannerUserID(ctx), request.DelaySec)
	sendResponse(ctx, r.logger, response, err)
	return nil
}
