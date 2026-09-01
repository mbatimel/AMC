package service

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

func validCompanyRequest() models.CompanyRequestInput {
	return models.CompanyRequestInput{
		ContactName: "Иван Иванов",
		Email:       "applicant@example.com",
		Phone:       "+7 999 000-00-00",
		Company:     "ООО Ромашка",
		Message:     "Хотим обсудить сотрудничество.",
	}
}

func pdfAttachment(name string, size int) models.ImageFile {
	content := make([]byte, size)
	copy(content, []byte("%PDF-1.4\n"))
	return models.ImageFile{FileName: name, Content: content}
}

func companyRequestService(mail *fakeMailer, maxFileSize int64) *service {
	return NewAdminApiService(
		zerolog.Nop(), &fakeStorage{}, &fakeAuthClient{}, &fakeAccessClient{allowed: true}, nil, mail,
		WithObjectStorage(nil, maxFileSize), WithCompanyRequestRecipient("order@voint.ru"),
	)
}

func TestSendCompanyRequestWithoutAttachments(t *testing.T) {
	mail := &fakeMailer{}
	response, err := companyRequestService(mail, 1024).SendCompanyRequest(context.Background(), validCompanyRequest())
	if err != nil || !response.Accepted {
		t.Fatalf("response=%+v err=%v", response, err)
	}
	if mail.calls != 1 || mail.to != "order@voint.ru" || len(mail.attachments) != 0 {
		t.Fatalf("mailer=%+v", mail)
	}
}

func TestSendCompanyRequestWithOneAndMultipleAttachments(t *testing.T) {
	for _, count := range []int{1, 3} {
		t.Run(string(rune('0'+count)), func(t *testing.T) {
			mail := &fakeMailer{}
			input := validCompanyRequest()
			for i := 0; i < count; i++ {
				input.Attachments = append(input.Attachments, pdfAttachment("document.pdf", 32))
			}
			if _, err := companyRequestService(mail, 1024).SendCompanyRequest(context.Background(), input); err != nil {
				t.Fatalf("SendCompanyRequest() error = %v", err)
			}
			if mail.calls != 1 || len(mail.attachments) != count {
				t.Fatalf("calls=%d attachments=%d", mail.calls, len(mail.attachments))
			}
		})
	}
}

func TestSendCompanyRequestRejectsInvalidAttachments(t *testing.T) {
	tests := []struct {
		name        string
		attachments []models.ImageFile
	}{
		{name: "file too large", attachments: []models.ImageFile{pdfAttachment("large.pdf", 1025)}},
		{name: "too many", attachments: []models.ImageFile{
			pdfAttachment("1.pdf", 16), pdfAttachment("2.pdf", 16), pdfAttachment("3.pdf", 16),
			pdfAttachment("4.pdf", 16), pdfAttachment("5.pdf", 16), pdfAttachment("6.pdf", 16),
		}},
		{name: "unsupported MIME", attachments: []models.ImageFile{{FileName: "payload.exe", Content: []byte("MZ executable")}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mail := &fakeMailer{}
			input := validCompanyRequest()
			input.Attachments = test.attachments
			_, err := companyRequestService(mail, 1024).SendCompanyRequest(context.Background(), input)
			if err == nil || mail.calls != 0 {
				t.Fatalf("err=%v mail calls=%d", err, mail.calls)
			}
		})
	}
}

func TestSendCompanyRequestRejectsInvalidEmail(t *testing.T) {
	mail := &fakeMailer{}
	input := validCompanyRequest()
	input.Email = "invalid"
	_, err := companyRequestService(mail, 1024).SendCompanyRequest(context.Background(), input)
	if err == nil || mail.calls != 0 {
		t.Fatalf("err=%v mail calls=%d", err, mail.calls)
	}
}

func TestSendCompanyRequestRequiresSuccessfulEmailDelivery(t *testing.T) {
	mail := &fakeMailer{sendErr: errors.New("smtp failed")}
	var logs bytes.Buffer
	logger := zerolog.New(&logs).With().Str("serviceName", "admin-api").Logger()
	svc := NewAdminApiService(
		logger, &fakeStorage{}, &fakeAuthClient{}, &fakeAccessClient{allowed: true}, nil, mail,
		WithObjectStorage(nil, 1024), WithCompanyRequestRecipient("order@voint.ru"),
	)
	_, err := svc.SendCompanyRequest(context.Background(), validCompanyRequest())
	if !customErrors.Is(err, customErrors.InternalServerError()) {
		t.Fatalf("error = %v, want internal", err)
	}
	if mail.calls != 1 || mail.to != "order@voint.ru" {
		t.Fatalf("calls=%d to=%q", mail.calls, mail.to)
	}
	logOutput := logs.String()
	for _, expected := range []string{"smtp failed", "company_request", "applicant@example.com", "order@voint.ru", "admin-api"} {
		if !strings.Contains(logOutput, expected) {
			t.Fatalf("SMTP failure log does not contain %q: %s", expected, logOutput)
		}
	}
}
