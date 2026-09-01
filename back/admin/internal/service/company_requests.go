package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

const (
	MaxCompanyRequestAttachments = 5
	maxCompanyRequestName        = 255
	maxCompanyRequestPhone       = 64
	maxCompanyRequestCompany     = 255
	maxCompanyRequestMessage     = 10000
)

func companyRequestValidation(field string) error {
	return customErrors.BadRequestError().AddCause("field", field)
}

func normalizeCompanyRequestPhone(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	digits := 0
	for _, r := range value {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case strings.ContainsRune("+()- ", r):
		default:
			return "", companyRequestValidation("phone")
		}
	}
	if digits < 7 || digits > 15 || utf8.RuneCountInString(value) > maxCompanyRequestPhone {
		return "", companyRequestValidation("phone")
	}
	return value, nil
}

func normalizeCompanyRequestField(value, field string, required bool, limit int) (string, error) {
	value = strings.TrimSpace(value)
	if (required && value == "") || utf8.RuneCountInString(value) > limit || strings.ContainsRune(value, '\x00') {
		return "", companyRequestValidation(field)
	}
	return value, nil
}

func (s *service) SendCompanyRequest(ctx context.Context, input models.CompanyRequestInput) (models.CompanyRequestResponse, error) {
	contactName, err := normalizeCompanyRequestField(input.ContactName, "contactName", true, maxCompanyRequestName)
	if err != nil {
		return models.CompanyRequestResponse{}, err
	}
	email, err := normalizeSignupEmail(input.Email)
	if err != nil {
		return models.CompanyRequestResponse{}, err
	}
	phone, err := normalizeCompanyRequestPhone(input.Phone)
	if err != nil {
		return models.CompanyRequestResponse{}, err
	}
	company, err := normalizeCompanyRequestField(input.Company, "company", false, maxCompanyRequestCompany)
	if err != nil {
		return models.CompanyRequestResponse{}, err
	}
	message, err := normalizeCompanyRequestField(input.Message, "message", true, maxCompanyRequestMessage)
	if err != nil {
		return models.CompanyRequestResponse{}, err
	}
	if len(input.Attachments) > MaxCompanyRequestAttachments {
		return models.CompanyRequestResponse{}, companyRequestValidation("attachmentsCount")
	}

	var totalSize int64
	attachments := make([]models.ImageFile, 0, len(input.Attachments))
	for _, attachment := range input.Attachments {
		if _, _, err = validateDocumentFile(attachment, s.maxFileSize, func(field string) *customErrors.Error {
			return customErrors.BadRequestError().AddCause("field", "attachments."+field)
		}); err != nil {
			return models.CompanyRequestResponse{}, err
		}
		totalSize += int64(len(attachment.Content))
		if totalSize > s.maxFileSize {
			return models.CompanyRequestResponse{}, companyRequestValidation("attachmentsTotalSize")
		}
		attachments = append(attachments, attachment)
	}

	if s.mailer == nil || strings.TrimSpace(s.companyRequestRecipient) == "" {
		s.logger.Error().
			Str("component", "smtp").
			Str("event", "company_request").
			Str("requester_email", email).
			Str("recipient", s.companyRequestRecipient).
			Msg("company request email delivery is not configured")
		return models.CompanyRequestResponse{}, customErrors.InternalServerError()
	}
	body := fmt.Sprintf(
		"Контактное лицо: %s\nEmail: %s\nТелефон: %s\nКомпания: %s\n\nСообщение:\n%s",
		contactName, email, phone, company, message,
	)
	if err = s.mailer.SendWithAttachments(
		ctx,
		s.companyRequestRecipient,
		"Заявка с сайта: О компании",
		body,
		attachments,
	); err != nil {
		s.logger.Error().
			Err(err).
			Str("component", "smtp").
			Str("event", "company_request").
			Str("requester_email", email).
			Str("recipient", s.companyRequestRecipient).
			Msg("failed to send company request email")
		return models.CompanyRequestResponse{}, customErrors.InternalServerError()
	}
	return models.CompanyRequestResponse{Accepted: true}, nil
}
