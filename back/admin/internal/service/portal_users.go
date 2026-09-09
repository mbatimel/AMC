package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/internal/storage/postgres"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

const maxPortalUserFieldLen = 255

func portalUserValidation(field string) error {
	return customErrors.BadRequestError().AddCause("field", field)
}

func normalizeInviteEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || utf8.RuneCountInString(value) > maxPortalUserFieldLen || strings.ContainsAny(value, "\r\n\x00") {
		return "", portalUserValidation("email")
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return "", portalUserValidation("email")
	}
	return value, nil
}

func normalizeInviteName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > maxPortalUserFieldLen {
		return "", portalUserValidation("name")
	}
	return value, nil
}

func inviteAdminEmailBody(name, email, password string) string {
	return fmt.Sprintf(
		"Здравствуйте, %s!\n\nВас пригласили в качестве администратора портала Volint.\n\nЛогин: %s\nПароль: %s\n\nВход: https://admin.volint.ru",
		name, email, password,
	)
}

// InviteAdmin creates a new portal admin account: generates a random
// password, creates the user with the admin role, and emails the
// credentials. A failed email delivery does not fail the request — the
// caller (an already-authenticated admin) gets the password back in the
// response as a fallback (see docs/superpowers/specs/2026-09-10-admin-invite-design.md).
func (s *service) InviteAdmin(ctx context.Context, userID uuid.UUID, email string, name string) (models.InviteAdminResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.InviteAdminResponse{}, err
	}

	email, err := normalizeInviteEmail(email)
	if err != nil {
		return models.InviteAdminResponse{}, err
	}
	name, err = normalizeInviteName(name)
	if err != nil {
		return models.InviteAdminResponse{}, err
	}

	password, err := generatePassword()
	if err != nil {
		return models.InviteAdminResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.InviteAdminResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}

	newUserID, err := s.storage.CreateAdminUser(ctx, email, string(passwordHash))
	if errors.Is(err, postgres.ErrEmailTaken) {
		return models.InviteAdminResponse{}, customErrors.ConflictError().AddCause("field", "email")
	}
	if err != nil {
		return models.InviteAdminResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}

	emailSent := true
	if s.mailer == nil {
		emailSent = false
		s.logger.Error().Str("email", email).Msg("invite admin: mailer is not configured, password not sent")
	} else if sendErr := s.mailer.Send(ctx, email, "Приглашение в панель администратора Volint", inviteAdminEmailBody(name, email, password)); sendErr != nil {
		emailSent = false
		s.logger.Error().Err(sendErr).Str("email", email).Msg("invite admin: failed to send invitation email")
	}

	if err = s.storage.InsertAuditLogEntry(ctx, userID, actorLabelAdmin, "Приглашён администратор: "+email); err != nil {
		return models.InviteAdminResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}

	return models.InviteAdminResponse{
		UserID:    newUserID,
		Email:     email,
		Password:  password,
		EmailSent: emailSent,
	}, nil
}
