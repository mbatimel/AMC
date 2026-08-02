package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/internal/storage/postgres"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

const (
	maxBannerTitle    = 255
	maxBannerSubtitle = 1000
	maxBannerLink     = 2000
)

type ObjectStorage interface {
	Upload(ctx context.Context, objectKey string, body io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, objectKey string) error
	URL(objectKey string) string
}

type Option func(*service)

func WithObjectStorage(storage ObjectStorage, maxFileSize int64) Option {
	return func(s *service) {
		s.objectStorage = storage
		s.maxFileSize = maxFileSize
	}
}

var bannerImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

func bannerValidation(field string) error {
	return customErrors.BadRequestError().AddCause("field", field)
}

func validateBannerImage(file models.ImageFile, maxFileSize int64) (string, string, error) {
	name := strings.TrimSpace(file.FileName)
	if name == "" || filepath.Base(name) != name || strings.Contains(name, "..") {
		return "", "", bannerValidation("fileName")
	}
	if len(file.Content) == 0 {
		return "", "", bannerValidation("file")
	}
	if maxFileSize <= 0 || int64(len(file.Content)) > maxFileSize {
		return "", "", bannerValidation("fileSize")
	}
	contentType := http.DetectContentType(file.Content[:min(len(file.Content), 512)])
	extension, ok := bannerImageTypes[contentType]
	if !ok {
		return "", "", bannerValidation("contentType")
	}
	suppliedExtension := strings.ToLower(filepath.Ext(name))
	if suppliedExtension != extension && !(contentType == "image/jpeg" && suppliedExtension == ".jpeg") {
		return "", "", bannerValidation("fileName")
	}
	return contentType, extension, nil
}

func parseBannerDate(value string, field string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return nil, bannerValidation(field)
	}
	return &parsed, nil
}

func normalizeBannerInput(input models.BannerInput) (postgres.Banner, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" || len([]rune(title)) > maxBannerTitle {
		return postgres.Banner{}, bannerValidation("title")
	}
	subtitle := strings.TrimSpace(input.Subtitle)
	if len([]rune(subtitle)) > maxBannerSubtitle {
		return postgres.Banner{}, bannerValidation("subtitle")
	}
	link := strings.TrimSpace(input.Link)
	if len([]rune(link)) > maxBannerLink {
		return postgres.Banner{}, bannerValidation("link")
	}
	if link != "" && !strings.HasPrefix(link, "/") {
		parsed, err := url.Parse(link)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return postgres.Banner{}, bannerValidation("link")
		}
	}
	if input.SortOrder < 0 {
		return postgres.Banner{}, bannerValidation("sortOrder")
	}
	dateFrom, err := parseBannerDate(input.DateFrom, "dateFrom")
	if err != nil {
		return postgres.Banner{}, err
	}
	dateTo, err := parseBannerDate(input.DateTo, "dateTo")
	if err != nil {
		return postgres.Banner{}, err
	}
	if dateFrom != nil && dateTo != nil && dateFrom.After(*dateTo) {
		return postgres.Banner{}, bannerValidation("dateTo")
	}
	return postgres.Banner{
		Title: title, Subtitle: subtitle, Link: link, SortOrder: input.SortOrder,
		IsActive: input.IsActive, DateFrom: dateFrom, DateTo: dateTo,
	}, nil
}

func modelBanner(banner postgres.Banner) models.Banner {
	result := models.Banner{
		ID: banner.ID, Title: banner.Title, Subtitle: banner.Subtitle,
		Image: banner.ImageURL, Link: banner.Link, SortOrder: banner.SortOrder,
		IsActive: banner.IsActive, CreatedAt: banner.CreatedAt, UpdatedAt: banner.UpdatedAt,
	}
	if banner.DateFrom != nil {
		result.DateFrom = banner.DateFrom.Format(time.DateOnly)
	}
	if banner.DateTo != nil {
		result.DateTo = banner.DateTo.Format(time.DateOnly)
	}
	return result
}

func modelBanners(items []postgres.Banner, delay int) models.BannersSettings {
	result := models.BannersSettings{DelaySec: delay, Items: make([]models.Banner, 0, len(items))}
	for _, item := range items {
		result.Items = append(result.Items, modelBanner(item))
	}
	return result
}

func mapBannerStorageError(err error) error {
	if errors.Is(err, postgres.ErrBannerNotFound) {
		return customErrors.NotFoundError()
	}
	return customErrors.InternalServerError().SetOuterError(err)
}

func (s *service) ListBanners(ctx context.Context, userID uuid.UUID) (models.BannersSettings, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.BannersSettings{}, err
	}
	items, delay, err := s.storage.ListBanners(ctx, false)
	if err != nil {
		return models.BannersSettings{}, mapBannerStorageError(err)
	}
	return modelBanners(items, delay), nil
}

func (s *service) ListPublicBanners(ctx context.Context) (models.BannersSettings, error) {
	items, delay, err := s.storage.ListBanners(ctx, true)
	if err != nil {
		return models.BannersSettings{}, mapBannerStorageError(err)
	}
	return modelBanners(items, delay), nil
}

func (s *service) uploadBannerObject(ctx context.Context, bannerID uuid.UUID, file models.ImageFile) (postgres.Banner, error) {
	if s.objectStorage == nil {
		return postgres.Banner{}, customErrors.InternalServerError()
	}
	contentType, extension, err := validateBannerImage(file, s.maxFileSize)
	if err != nil {
		return postgres.Banner{}, err
	}
	objectKey := fmt.Sprintf("banners/%s/%s%s", bannerID, uuid.NewString(), extension)
	if err = s.objectStorage.Upload(ctx, objectKey, bytes.NewReader(file.Content), int64(len(file.Content)), contentType); err != nil {
		return postgres.Banner{}, customErrors.InternalServerError().SetOuterError(err)
	}
	return postgres.Banner{
		ImageURL: s.objectStorage.URL(objectKey), ObjectKey: objectKey,
		OriginalName: file.FileName, ContentType: contentType, SizeBytes: int64(len(file.Content)),
	}, nil
}

func (s *service) CreateBanner(ctx context.Context, userID uuid.UUID, input models.BannerInput) (models.Banner, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.Banner{}, err
	}
	if input.Image == nil {
		return models.Banner{}, bannerValidation("file")
	}
	banner, err := normalizeBannerInput(input)
	if err != nil {
		return models.Banner{}, err
	}
	banner.ID = uuid.New()
	image, err := s.uploadBannerObject(ctx, banner.ID, *input.Image)
	if err != nil {
		return models.Banner{}, err
	}
	banner.ImageURL, banner.ObjectKey = image.ImageURL, image.ObjectKey
	banner.OriginalName, banner.ContentType, banner.SizeBytes = image.OriginalName, image.ContentType, image.SizeBytes
	created, err := s.storage.CreateBanner(ctx, banner)
	if err != nil {
		if cleanupErr := s.objectStorage.Delete(ctx, banner.ObjectKey); cleanupErr != nil {
			s.logger.Error().Err(cleanupErr).Str("objectKey", banner.ObjectKey).Msg("failed to compensate banner upload")
		}
		return models.Banner{}, mapBannerStorageError(err)
	}
	s.audit(ctx, userID, "Добавлен баннер")
	return modelBanner(created), nil
}

func (s *service) UpdateBanner(ctx context.Context, userID uuid.UUID, bannerID uuid.UUID, input models.BannerInput) (models.Banner, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.Banner{}, err
	}
	if bannerID == uuid.Nil {
		return models.Banner{}, bannerValidation("bannerID")
	}
	current, err := s.storage.GetBanner(ctx, bannerID)
	if err != nil {
		return models.Banner{}, mapBannerStorageError(err)
	}
	banner, err := normalizeBannerInput(input)
	if err != nil {
		return models.Banner{}, err
	}
	banner.ID, banner.FileID = current.ID, current.FileID
	replaceImage := input.Image != nil
	if replaceImage {
		image, uploadErr := s.uploadBannerObject(ctx, bannerID, *input.Image)
		if uploadErr != nil {
			return models.Banner{}, uploadErr
		}
		banner.ImageURL, banner.ObjectKey = image.ImageURL, image.ObjectKey
		banner.OriginalName, banner.ContentType, banner.SizeBytes = image.OriginalName, image.ContentType, image.SizeBytes
	}
	updated, err := s.storage.UpdateBanner(ctx, banner, replaceImage)
	if err != nil {
		if replaceImage {
			if cleanupErr := s.objectStorage.Delete(ctx, banner.ObjectKey); cleanupErr != nil {
				s.logger.Error().Err(cleanupErr).Str("objectKey", banner.ObjectKey).Msg("failed to compensate banner replacement")
			}
		}
		return models.Banner{}, mapBannerStorageError(err)
	}
	if replaceImage && current.ObjectKey != "" {
		if cleanupErr := s.objectStorage.Delete(ctx, current.ObjectKey); cleanupErr != nil {
			s.logger.Error().Err(cleanupErr).Str("objectKey", current.ObjectKey).Msg("failed to delete replaced banner object")
		}
	}
	s.audit(ctx, userID, "Обновлён баннер")
	return modelBanner(updated), nil
}

func (s *service) DeleteBanner(ctx context.Context, userID uuid.UUID, bannerID uuid.UUID) (models.DeleteBannerResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.DeleteBannerResponse{}, err
	}
	current, err := s.storage.GetBanner(ctx, bannerID)
	if err != nil {
		return models.DeleteBannerResponse{}, mapBannerStorageError(err)
	}
	if current.ObjectKey != "" {
		if err = s.objectStorage.Delete(ctx, current.ObjectKey); err != nil {
			return models.DeleteBannerResponse{}, customErrors.InternalServerError().SetOuterError(err)
		}
	}
	if err = s.storage.DeleteBanner(ctx, bannerID); err != nil {
		return models.DeleteBannerResponse{}, mapBannerStorageError(err)
	}
	s.audit(ctx, userID, "Удалён баннер")
	return models.DeleteBannerResponse{Deleted: true}, nil
}

func (s *service) UpdateBannerDelay(ctx context.Context, userID uuid.UUID, delaySec int) (models.BannersSettings, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.BannersSettings{}, err
	}
	if delaySec <= 0 || delaySec > 3600 {
		return models.BannersSettings{}, bannerValidation("delaySec")
	}
	if err := s.storage.UpdateBannerDelay(ctx, delaySec); err != nil {
		return models.BannersSettings{}, mapBannerStorageError(err)
	}
	items, _, err := s.storage.ListBanners(ctx, false)
	if err != nil {
		return models.BannersSettings{}, mapBannerStorageError(err)
	}
	return modelBanners(items, delaySec), nil
}

func (s *service) audit(ctx context.Context, userID uuid.UUID, action string) {
	if err := s.storage.InsertAuditLogEntry(ctx, userID, actorLabelAdmin, action); err != nil {
		s.logger.Error().Err(err).Msg("failed to write audit log entry")
	}
}
