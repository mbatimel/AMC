// back/admin/internal/service/legal_docs.go
package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/internal/storage/postgres"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

const (
	maxLegalDocID      = 64
	maxLegalDocName    = 255
	maxLegalDocVersion = 32
	maxLegalDocSummary = 2000

	defaultLegalDocVersion        = "1.0"
	defaultLegalDocCreateSummary  = "Первая версия"
	defaultLegalDocReplaceSummary = "Обновление документа"
)

var legalDocIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:[-_][a-z0-9]+)*$`)

func legalDocValidation(field string) *customErrors.Error {
	return customErrors.BadRequestError().AddCause("field", field)
}

func mapLegalDocStorageError(err error) error {
	if errors.Is(err, postgres.ErrLegalDocNotFound) {
		return customErrors.NotFoundError()
	}
	if errors.Is(err, postgres.ErrLegalDocAlreadyExists) {
		return customErrors.ConflictError().AddCause("field", "id")
	}
	return customErrors.InternalServerError().SetOuterError(err)
}

func normalizeLegalDocID(raw string) (string, error) {
	id := strings.TrimSpace(raw)
	if id == "" || len(id) > maxLegalDocID || !legalDocIDPattern.MatchString(id) {
		return "", legalDocValidation("id")
	}
	return id, nil
}

func normalizeLegalDocName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" || len([]rune(name)) > maxLegalDocName {
		return "", legalDocValidation("name")
	}
	return name, nil
}

func normalizeLegalDocVersion(raw string) (string, error) {
	version := strings.TrimSpace(raw)
	if version == "" || len([]rune(version)) > maxLegalDocVersion {
		return "", legalDocValidation("version")
	}
	return version, nil
}

func normalizeLegalDocSummary(raw string, fallback string) (string, error) {
	summary := strings.TrimSpace(raw)
	if summary == "" {
		return fallback, nil
	}
	if len([]rune(summary)) > maxLegalDocSummary {
		return "", legalDocValidation("summary")
	}
	return summary, nil
}

// decodeLegalDocFile decodes a base64-encoded file payload coming over the
// JSON-RPC/REST transport into an in-memory file ready for upload.
func decodeLegalDocFile(fileName string, fileContentBase64 string) (models.ImageFile, error) {
	if strings.TrimSpace(fileName) == "" || strings.TrimSpace(fileContentBase64) == "" {
		return models.ImageFile{}, legalDocValidation("file")
	}
	content, err := base64.StdEncoding.DecodeString(fileContentBase64)
	if err != nil {
		return models.ImageFile{}, legalDocValidation("file")
	}
	return models.ImageFile{FileName: fileName, Content: content}, nil
}

func (s *service) modelLegalDoc(doc postgres.LegalDoc) models.LegalDoc {
	fileURL := ""
	if doc.ObjectKey != "" && s.objectStorage != nil {
		fileURL = s.objectStorage.URL(doc.ObjectKey)
	}
	return models.LegalDoc{
		ID: doc.ID, Name: doc.Name, CurrentVersion: doc.CurrentVersion,
		FileURL: fileURL, UpdatedAt: doc.UpdatedAt,
	}
}

func (s *service) modelLegalDocs(docs []postgres.LegalDoc) models.ListLegalDocsResponse {
	result := models.ListLegalDocsResponse{Items: make([]models.LegalDoc, 0, len(docs))}
	for _, doc := range docs {
		result.Items = append(result.Items, s.modelLegalDoc(doc))
	}
	return result
}

func (s *service) modelLegalDocVersion(version postgres.LegalDocVersion) models.LegalDocVersion {
	fileURL := ""
	if version.ObjectKey != "" && s.objectStorage != nil {
		fileURL = s.objectStorage.URL(version.ObjectKey)
	}
	return models.LegalDocVersion{
		ID: version.ID, Version: version.Version, Summary: version.Summary,
		Author: version.Author, FileURL: fileURL, CreatedAt: version.CreatedAt,
	}
}

// uploadLegalDocFile validates and uploads a legal-document file, returning
// the storage record ready to be persisted.
func (s *service) uploadLegalDocFile(ctx context.Context, docID string, file models.ImageFile) (postgres.LegalDoc, error) {
	if s.objectStorage == nil {
		return postgres.LegalDoc{}, customErrors.InternalServerError()
	}
	contentType, extension, err := validateDocumentFile(file, s.maxFileSize, legalDocValidation)
	if err != nil {
		return postgres.LegalDoc{}, err
	}
	objectKey := fmt.Sprintf("legal-docs/%s/%s%s", docID, uuid.NewString(), extension)
	if err = s.objectStorage.Upload(ctx, objectKey, bytes.NewReader(file.Content), int64(len(file.Content)), contentType); err != nil {
		return postgres.LegalDoc{}, customErrors.InternalServerError().SetOuterError(err)
	}
	return postgres.LegalDoc{
		ObjectKey: objectKey, OriginalName: file.FileName,
		ContentType: contentType, SizeBytes: int64(len(file.Content)),
	}, nil
}

// CreateLegalDoc adds a new document to the "Документы и соглашения" list
// together with its first file/version.
func (s *service) CreateLegalDoc(ctx context.Context, userID uuid.UUID, id string, name string, version string, summary string, fileName string, fileContentBase64 string) (models.LegalDoc, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.LegalDoc{}, err
	}
	docID, err := normalizeLegalDocID(id)
	if err != nil {
		return models.LegalDoc{}, err
	}
	docName, err := normalizeLegalDocName(name)
	if err != nil {
		return models.LegalDoc{}, err
	}
	if version = strings.TrimSpace(version); version == "" {
		version = defaultLegalDocVersion
	}
	if version, err = normalizeLegalDocVersion(version); err != nil {
		return models.LegalDoc{}, err
	}
	docSummary, err := normalizeLegalDocSummary(summary, defaultLegalDocCreateSummary)
	if err != nil {
		return models.LegalDoc{}, err
	}
	uploadedFile, err := decodeLegalDocFile(fileName, fileContentBase64)
	if err != nil {
		return models.LegalDoc{}, err
	}

	file, err := s.uploadLegalDocFile(ctx, docID, uploadedFile)
	if err != nil {
		return models.LegalDoc{}, err
	}

	doc := postgres.LegalDoc{
		ID: docID, Name: docName, CurrentVersion: version,
		ObjectKey: file.ObjectKey, OriginalName: file.OriginalName,
		ContentType: file.ContentType, SizeBytes: file.SizeBytes,
	}
	versionRecord := postgres.LegalDocVersion{
		Version: version, Summary: docSummary, Author: actorLabelAdmin,
		ObjectKey: file.ObjectKey, OriginalName: file.OriginalName,
		ContentType: file.ContentType, SizeBytes: file.SizeBytes,
	}

	created, err := s.storage.CreateLegalDoc(ctx, doc, versionRecord)
	if err != nil {
		if cleanupErr := s.objectStorage.Delete(ctx, file.ObjectKey); cleanupErr != nil {
			s.logger.Error().Err(cleanupErr).Str("objectKey", file.ObjectKey).Msg("failed to compensate legal doc upload")
		}
		return models.LegalDoc{}, mapLegalDocStorageError(err)
	}
	s.audit(ctx, userID, "Добавлен документ: "+docName)
	return s.modelLegalDoc(created), nil
}

// ReplaceLegalDocFile uploads a new file for an existing document ("Заменить"
// in the admin panel), keeping every previous version in history.
func (s *service) ReplaceLegalDocFile(ctx context.Context, userID uuid.UUID, docID string, version string, summary string, fileName string, fileContentBase64 string) (models.LegalDoc, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.LegalDoc{}, err
	}
	docID = strings.TrimSpace(docID)
	if docID == "" {
		return models.LegalDoc{}, legalDocValidation("docID")
	}
	current, err := s.storage.GetLegalDoc(ctx, docID)
	if err != nil {
		return models.LegalDoc{}, mapLegalDocStorageError(err)
	}
	newVersion, err := normalizeLegalDocVersion(version)
	if err != nil {
		return models.LegalDoc{}, err
	}
	if newVersion == current.CurrentVersion {
		return models.LegalDoc{}, customErrors.ConflictError().AddCause("field", "version")
	}
	docSummary, err := normalizeLegalDocSummary(summary, defaultLegalDocReplaceSummary)
	if err != nil {
		return models.LegalDoc{}, err
	}
	uploadedFile, err := decodeLegalDocFile(fileName, fileContentBase64)
	if err != nil {
		return models.LegalDoc{}, err
	}

	file, err := s.uploadLegalDocFile(ctx, docID, uploadedFile)
	if err != nil {
		return models.LegalDoc{}, err
	}

	versionRecord := postgres.LegalDocVersion{
		Version: newVersion, Summary: docSummary, Author: actorLabelAdmin,
		ObjectKey: file.ObjectKey, OriginalName: file.OriginalName,
		ContentType: file.ContentType, SizeBytes: file.SizeBytes,
	}
	updated, err := s.storage.ReplaceLegalDocFile(ctx, docID, versionRecord)
	if err != nil {
		if cleanupErr := s.objectStorage.Delete(ctx, file.ObjectKey); cleanupErr != nil {
			s.logger.Error().Err(cleanupErr).Str("objectKey", file.ObjectKey).Msg("failed to compensate legal doc replacement")
		}
		return models.LegalDoc{}, mapLegalDocStorageError(err)
	}
	s.audit(ctx, userID, "Заменён файл документа: "+current.Name)
	return s.modelLegalDoc(updated), nil
}

// DeleteLegalDoc removes a document from the list entirely, including every
// version's file.
func (s *service) DeleteLegalDoc(ctx context.Context, userID uuid.UUID, docID string) (models.DeleteLegalDocResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.DeleteLegalDocResponse{}, err
	}
	docID = strings.TrimSpace(docID)
	current, err := s.storage.GetLegalDoc(ctx, docID)
	if err != nil {
		return models.DeleteLegalDocResponse{}, mapLegalDocStorageError(err)
	}
	versions, err := s.storage.ListLegalDocVersions(ctx, docID)
	if err != nil {
		return models.DeleteLegalDocResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}
	if err = s.storage.DeleteLegalDoc(ctx, docID); err != nil {
		return models.DeleteLegalDocResponse{}, mapLegalDocStorageError(err)
	}
	if s.objectStorage != nil {
		for _, version := range versions {
			if version.ObjectKey == "" {
				continue
			}
			if delErr := s.objectStorage.Delete(ctx, version.ObjectKey); delErr != nil {
				s.logger.Error().Err(delErr).Str("objectKey", version.ObjectKey).Msg("failed to delete legal doc file object")
			}
		}
	}
	s.audit(ctx, userID, "Удалён документ: "+current.Name)
	return models.DeleteLegalDocResponse{Deleted: true}, nil
}

func (s *service) ListLegalDocs(ctx context.Context, userID uuid.UUID) (models.ListLegalDocsResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.ListLegalDocsResponse{}, err
	}
	docs, err := s.storage.ListLegalDocs(ctx)
	if err != nil {
		return models.ListLegalDocsResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}
	return s.modelLegalDocs(docs), nil
}

func (s *service) ListPublicLegalDocs(ctx context.Context) (models.ListLegalDocsResponse, error) {
	docs, err := s.storage.ListLegalDocs(ctx)
	if err != nil {
		return models.ListLegalDocsResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}
	return s.modelLegalDocs(docs), nil
}

func (s *service) ListLegalDocVersions(ctx context.Context, userID uuid.UUID, docID string) (models.ListLegalDocVersionsResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.ListLegalDocVersionsResponse{}, err
	}
	docID = strings.TrimSpace(docID)
	versions, err := s.storage.ListLegalDocVersions(ctx, docID)
	if err != nil {
		return models.ListLegalDocVersionsResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}
	result := models.ListLegalDocVersionsResponse{Items: make([]models.LegalDocVersion, 0, len(versions))}
	for _, version := range versions {
		result.Items = append(result.Items, s.modelLegalDocVersion(version))
	}
	return result, nil
}
