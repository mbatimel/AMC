package service

import (
	"encoding/base64"
	"strconv"
	"testing"

	"github.com/mbatimel/AMC/admin/pkg/models"
)

func testPDF(size int) models.ImageFile {
	content := make([]byte, size)
	copy(content, []byte("%PDF-1.4\n"))
	return models.ImageFile{FileName: "document.pdf", Content: content}
}

func TestValidateDocumentFileSizes(t *testing.T) {
	const maxSize = 10 * 1024 * 1024
	for _, size := range []int{128, 3 * 1024 * 1024, maxSize} {
		t.Run(formatSize(size), func(t *testing.T) {
			contentType, extension, err := validateDocumentFile(testPDF(size), maxSize, legalDocValidation)
			if err != nil || contentType != "application/pdf" || extension != ".pdf" {
				t.Fatalf("type=%q extension=%q err=%v", contentType, extension, err)
			}
		})
	}
	if _, _, err := validateDocumentFile(testPDF(maxSize+1), maxSize, legalDocValidation); err == nil {
		t.Fatal("file above the limit was accepted")
	}
}

func TestValidateDocumentFileRejectsUnsupportedType(t *testing.T) {
	file := models.ImageFile{FileName: "archive.zip", Content: []byte("PK\x03\x04 payload")}
	if _, _, err := validateDocumentFile(file, 1024, legalDocValidation); err == nil {
		t.Fatal("unsupported content type was accepted")
	}
}

func TestDecodeLegalDocFileRejectsMalformedBase64(t *testing.T) {
	if _, err := decodeLegalDocFile("document.pdf", "%%%not-base64%%%"); err == nil {
		t.Fatal("malformed base64 was accepted")
	}
}

func TestDecodeLegalDocFileAcceptsValidBase64(t *testing.T) {
	payload := testPDF(128)
	decoded, err := decodeLegalDocFile(payload.FileName, base64.StdEncoding.EncodeToString(payload.Content))
	if err != nil || len(decoded.Content) != len(payload.Content) {
		t.Fatalf("decoded size=%d err=%v", len(decoded.Content), err)
	}
}

func formatSize(size int) string {
	if size%(1024*1024) == 0 {
		return strconv.Itoa(size/(1024*1024)) + "MiB"
	}
	return "small"
}
