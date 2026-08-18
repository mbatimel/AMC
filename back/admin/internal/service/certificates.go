// back/admin/internal/service/certificates.go
package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/internal/storage/postgres"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

const maxCertificateTitle = 255

func certificateValidation(field string) *customErrors.Error {
	return customErrors.BadRequestError().AddCause("field", field)
}

func mapCertificateStorageError(err error) error {
	if errors.Is(err, postgres.ErrCertificateNotFound) {
		return customErrors.NotFoundError()
	}
	return customErrors.InternalServerError().SetOuterError(err)
}

func normalizeCertificateTitle(raw string) (string, error) {
	title := strings.TrimSpace(raw)
	if title == "" || len([]rune(title)) > maxCertificateTitle {
		return "", certificateValidation("title")
	}
	return title, nil
}

// decodeCertificateFile decodes a base64-encoded file payload coming over
// the JSON-RPC/REST transport into an in-memory file ready for upload. An
// empty fileContentBase64 means "no file supplied" and is not an error by
// itself — callers decide whether that is acceptable.
func decodeCertificateFile(fileName string, fileContentBase64 string) (*models.ImageFile, error) {
	if strings.TrimSpace(fileName) == "" && strings.TrimSpace(fileContentBase64) == "" {
		return nil, nil
	}
	if strings.TrimSpace(fileName) == "" || strings.TrimSpace(fileContentBase64) == "" {
		return nil, certificateValidation("file")
	}
	content, err := base64.StdEncoding.DecodeString(fileContentBase64)
	if err != nil {
		return nil, certificateValidation("file")
	}
	return &models.ImageFile{FileName: fileName, Content: content}, nil
}

func (s *service) modelCertificate(cert postgres.Certificate) models.Certificate {
	fileURL := ""
	if cert.ObjectKey != "" && s.objectStorage != nil {
		fileURL = s.objectStorage.URL(cert.ObjectKey)
	}
	return models.Certificate{
		ID: cert.ID, Title: cert.Title, SortOrder: cert.SortOrder, IsActive: cert.IsActive,
		FileURL: fileURL, CreatedAt: cert.CreatedAt, UpdatedAt: cert.UpdatedAt,
	}
}

func (s *service) modelCertificates(items []postgres.Certificate) models.ListCertificatesResponse {
	result := models.ListCertificatesResponse{Items: make([]models.Certificate, 0, len(items))}
	for _, item := range items {
		result.Items = append(result.Items, s.modelCertificate(item))
	}
	return result
}

func (s *service) uploadCertificateFile(ctx context.Context, certID uuid.UUID, file models.ImageFile) (postgres.Certificate, error) {
	if s.objectStorage == nil {
		return postgres.Certificate{}, customErrors.InternalServerError()
	}
	contentType, extension, err := validateDocumentFile(file, s.maxFileSize, certificateValidation)
	if err != nil {
		return postgres.Certificate{}, err
	}
	objectKey := fmt.Sprintf("certificates/%s/%s%s", certID, uuid.NewString(), extension)
	if err = s.objectStorage.Upload(ctx, objectKey, bytes.NewReader(file.Content), int64(len(file.Content)), contentType); err != nil {
		return postgres.Certificate{}, customErrors.InternalServerError().SetOuterError(err)
	}
	return postgres.Certificate{
		ObjectKey: objectKey, OriginalName: file.FileName,
		ContentType: contentType, SizeBytes: int64(len(file.Content)),
	}, nil
}

func (s *service) ListCertificates(ctx context.Context, userID uuid.UUID) (models.ListCertificatesResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.ListCertificatesResponse{}, err
	}
	items, err := s.storage.ListCertificates(ctx, false)
	if err != nil {
		return models.ListCertificatesResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}
	return s.modelCertificates(items), nil
}

func (s *service) ListPublicCertificates(ctx context.Context) (models.ListCertificatesResponse, error) {
	items, err := s.storage.ListCertificates(ctx, true)
	if err != nil {
		return models.ListCertificatesResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}
	return s.modelCertificates(items), nil
}

func (s *service) CreateCertificate(ctx context.Context, userID uuid.UUID, title string, sortOrder int, isActive bool, fileName string, fileContentBase64 string) (models.Certificate, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.Certificate{}, err
	}
	certTitle, err := normalizeCertificateTitle(title)
	if err != nil {
		return models.Certificate{}, err
	}
	if sortOrder < 0 {
		return models.Certificate{}, certificateValidation("sortOrder")
	}
	uploadedFile, err := decodeCertificateFile(fileName, fileContentBase64)
	if err != nil {
		return models.Certificate{}, err
	}
	if uploadedFile == nil {
		return models.Certificate{}, certificateValidation("file")
	}

	cert := postgres.Certificate{ID: uuid.New(), Title: certTitle, SortOrder: sortOrder, IsActive: isActive}
	file, err := s.uploadCertificateFile(ctx, cert.ID, *uploadedFile)
	if err != nil {
		return models.Certificate{}, err
	}
	cert.ObjectKey, cert.OriginalName, cert.ContentType, cert.SizeBytes = file.ObjectKey, file.OriginalName, file.ContentType, file.SizeBytes

	created, err := s.storage.CreateCertificate(ctx, cert)
	if err != nil {
		if cleanupErr := s.objectStorage.Delete(ctx, cert.ObjectKey); cleanupErr != nil {
			s.logger.Error().Err(cleanupErr).Str("objectKey", cert.ObjectKey).Msg("failed to compensate certificate upload")
		}
		return models.Certificate{}, mapCertificateStorageError(err)
	}
	s.audit(ctx, userID, "Добавлен сертификат: "+certTitle)
	return s.modelCertificate(created), nil
}

func (s *service) UpdateCertificate(ctx context.Context, userID uuid.UUID, certID uuid.UUID, title string, sortOrder int, isActive bool, fileName string, fileContentBase64 string) (models.Certificate, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.Certificate{}, err
	}
	if certID == uuid.Nil {
		return models.Certificate{}, certificateValidation("certID")
	}
	current, err := s.storage.GetCertificate(ctx, certID)
	if err != nil {
		return models.Certificate{}, mapCertificateStorageError(err)
	}
	certTitle, err := normalizeCertificateTitle(title)
	if err != nil {
		return models.Certificate{}, err
	}
	if sortOrder < 0 {
		return models.Certificate{}, certificateValidation("sortOrder")
	}
	uploadedFile, err := decodeCertificateFile(fileName, fileContentBase64)
	if err != nil {
		return models.Certificate{}, err
	}

	cert := postgres.Certificate{ID: current.ID, FileID: current.FileID, Title: certTitle, SortOrder: sortOrder, IsActive: isActive}
	replaceFile := uploadedFile != nil
	if replaceFile {
		file, uploadErr := s.uploadCertificateFile(ctx, certID, *uploadedFile)
		if uploadErr != nil {
			return models.Certificate{}, uploadErr
		}
		cert.ObjectKey, cert.OriginalName, cert.ContentType, cert.SizeBytes = file.ObjectKey, file.OriginalName, file.ContentType, file.SizeBytes
	}
	updated, err := s.storage.UpdateCertificate(ctx, cert, replaceFile)
	if err != nil {
		if replaceFile {
			if cleanupErr := s.objectStorage.Delete(ctx, cert.ObjectKey); cleanupErr != nil {
				s.logger.Error().Err(cleanupErr).Str("objectKey", cert.ObjectKey).Msg("failed to compensate certificate replacement")
			}
		}
		return models.Certificate{}, mapCertificateStorageError(err)
	}
	if replaceFile && current.ObjectKey != "" {
		if cleanupErr := s.objectStorage.Delete(ctx, current.ObjectKey); cleanupErr != nil {
			s.logger.Error().Err(cleanupErr).Str("objectKey", current.ObjectKey).Msg("failed to delete replaced certificate object")
		}
	}
	s.audit(ctx, userID, "Обновлён сертификат: "+certTitle)
	return s.modelCertificate(updated), nil
}

func (s *service) DeleteCertificate(ctx context.Context, userID uuid.UUID, certID uuid.UUID) (models.DeleteCertificateResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.DeleteCertificateResponse{}, err
	}
	current, err := s.storage.GetCertificate(ctx, certID)
	if err != nil {
		return models.DeleteCertificateResponse{}, mapCertificateStorageError(err)
	}
	if current.ObjectKey != "" && s.objectStorage != nil {
		if err = s.objectStorage.Delete(ctx, current.ObjectKey); err != nil {
			return models.DeleteCertificateResponse{}, customErrors.InternalServerError().SetOuterError(err)
		}
	}
	if err = s.storage.DeleteCertificate(ctx, certID); err != nil {
		return models.DeleteCertificateResponse{}, mapCertificateStorageError(err)
	}
	s.audit(ctx, userID, "Удалён сертификат: "+current.Title)
	return models.DeleteCertificateResponse{Deleted: true}, nil
}
