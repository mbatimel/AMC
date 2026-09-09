package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	customErrors "github.com/mbatimel/AMC/auth/internal/errors"
	"github.com/mbatimel/AMC/auth/internal/storage/postgres"
)

const minResetPasswordLength = 6

// resetTokenBytes is the amount of random bytes making up a raw reset
// token before hex-encoding (32 bytes -> 64 hex chars, ~256 bits of entropy).
const resetTokenBytes = 32

func generateResetToken() (string, error) {
	buf := make([]byte, resetTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate reset token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// hashResetToken deterministically hashes a raw token for storage/lookup.
// Reset tokens are already high-entropy random values (unlike passwords),
// so a fast, deterministic hash is correct here — bcrypt's per-hash salt
// would make an exact-match lookup by token alone impossible.
func hashResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func normalizeResetEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || strings.ContainsAny(value, "\r\n\x00") {
		return "", customErrors.ValidationError("email")
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return "", customErrors.ValidationError("email")
	}
	return value, nil
}

func passwordResetEmailBody(baseURL, token string) string {
	link := strings.TrimRight(baseURL, "/") + "/reset-password?token=" + token
	return fmt.Sprintf(
		"Вы (или кто-то другой) запросили сброс пароля.\n\nСсылка для сброса (действует ограниченное время):\n%s\n\nЕсли это были не вы — просто проигнорируйте это письмо, пароль останется прежним.",
		link,
	)
}

// RequestPasswordReset never reveals whether the email belongs to a real,
// active account — callers always get the same emailSent=true response
// for a well-formed email, whether or not anything was actually sent.
func (s *service) RequestPasswordReset(ctx context.Context, email string) (emailSent bool, err error) {
	email, err = normalizeResetEmail(email)
	if err != nil {
		return false, err
	}

	user, err := s.storage.GetUserByEmail(ctx, email)
	if errors.Is(err, postgres.ErrUserNotFound) {
		return true, nil
	}
	if err != nil {
		return false, customErrors.InternalServerError().SetOuterError(err)
	}
	if !user.IsActive {
		return true, nil
	}

	token, err := generateResetToken()
	if err != nil {
		return false, customErrors.InternalServerError().SetOuterError(err)
	}

	if err = s.storage.InvalidateUserPasswordResetTokens(ctx, user.ID); err != nil {
		return false, customErrors.InternalServerError().SetOuterError(err)
	}
	if err = s.storage.CreatePasswordResetToken(ctx, user.ID, hashResetToken(token), time.Now().Add(s.resetTokenTTL)); err != nil {
		return false, customErrors.InternalServerError().SetOuterError(err)
	}

	if s.mailer == nil {
		s.logger.Error().Str("email", email).Msg("request password reset: mailer is not configured, link not sent")
		return false, nil
	}
	if sendErr := s.mailer.Send(ctx, email, "Сброс пароля", passwordResetEmailBody(s.publicFrontBaseURL, token)); sendErr != nil {
		s.logger.Error().Err(sendErr).Str("email", email).Msg("request password reset: failed to send email")
		return false, nil
	}
	return true, nil
}

func (s *service) ConfirmPasswordReset(ctx context.Context, token string, newPassword string) error {
	if strings.TrimSpace(token) == "" {
		return customErrors.ValidationError("token")
	}
	if len(newPassword) < minResetPasswordLength {
		return customErrors.ValidationError("newPassword")
	}

	tokenHash := hashResetToken(token)
	resetToken, err := s.storage.GetPasswordResetToken(ctx, tokenHash)
	if errors.Is(err, postgres.ErrTokenNotFound) {
		return customErrors.TokenInvalidError()
	}
	if err != nil {
		return customErrors.InternalServerError().SetOuterError(err)
	}
	if resetToken.UsedAt.Valid {
		return customErrors.TokenInvalidError()
	}
	if time.Now().After(resetToken.ExpiresAt) {
		return customErrors.TokenExpiredError()
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return customErrors.InternalServerError().SetOuterError(err)
	}

	if err = s.storage.UpdateUserPassword(ctx, resetToken.UserID, string(passwordHash)); err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			return customErrors.NotFoundError()
		}
		return customErrors.InternalServerError().SetOuterError(err)
	}
	if err = s.storage.InvalidateUserPasswordResetTokens(ctx, resetToken.UserID); err != nil {
		return customErrors.InternalServerError().SetOuterError(err)
	}

	return nil
}
