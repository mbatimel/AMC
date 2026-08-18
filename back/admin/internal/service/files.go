// back/admin/internal/service/files.go
package service

import (
	"net/http"
	"path/filepath"
	"strings"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

// documentFileTypes covers legal documents and certificates: unlike banner
// images, these are usually scanned PDFs.
var documentFileTypes = map[string]string{
	"application/pdf": ".pdf",
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/webp":      ".webp",
}

// validateDocumentFile checks a generically-uploaded file (legal document or
// certificate) against size limits and an allow-listed content type, mirroring
// validateBannerImage's rules but with a wider set of accepted formats.
func validateDocumentFile(file models.ImageFile, maxFileSize int64, invalidField func(string) *customErrors.Error) (contentType string, extension string, err error) {
	name := strings.TrimSpace(file.FileName)
	if name == "" || filepath.Base(name) != name || strings.Contains(name, "..") {
		return "", "", invalidField("fileName")
	}
	if len(file.Content) == 0 {
		return "", "", invalidField("file")
	}
	if maxFileSize <= 0 || int64(len(file.Content)) > maxFileSize {
		return "", "", invalidField("fileSize")
	}
	contentType = http.DetectContentType(file.Content[:min(len(file.Content), 512)])
	extension, ok := documentFileTypes[contentType]
	if !ok {
		return "", "", invalidField("contentType")
	}
	suppliedExtension := strings.ToLower(filepath.Ext(name))
	if suppliedExtension != extension && !(contentType == "image/jpeg" && suppliedExtension == ".jpeg") {
		return "", "", invalidField("fileName")
	}
	return contentType, extension, nil
}
