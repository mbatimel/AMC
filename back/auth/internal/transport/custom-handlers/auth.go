package custom_handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mbatimel/AMC/auth/internal/errors"
	externalapi "github.com/mbatimel/AMC/auth/pkg/interfaces/externalAPI"
	"github.com/rs/zerolog/log"
)

const ServiceName = "auth"

func LoginUser(ctx *fiber.Ctx, svc externalapi.AuthAPI, email string, password string) error {
	var (
		methodName = "LoginUser"
		err        error
	)

	defer func(begin time.Time) {
		fields := map[string]interface{}{
			"method":     "post",
			"path":       "/v1/auth/login",
			"methodName": methodName,
			"email":      email,
			"password":   password,
			"took":       time.Since(begin).String(),
		}
		l := log.Info()
		if err != nil {
			if errors.Is(err, errors.ForbiddenError()) || errors.Is(err, errors.UserBlockedError("", "", "", "")) {
				l = log.Warn().Err(err)
			} else {
				l = log.Error().Err(err)
			}
		}
		l.Fields(fields).Msg("call")

	}(time.Now())

	userID, err := svc.LoginUser(ctx.UserContext(), email, password)
	if err != nil {
		if blockedErr, ok := err.(*errors.Error); ok && errors.Is(err, errors.UserBlockedError("", "", "", "")) {
			blockedErr.AddCause("siteDomain", ctx.Hostname())
		}
		sendResponse(ctx, log.Logger, nil, err)
		return nil
	}

	sendResponse(ctx, log.Logger, userID, nil)
	return err
}

func LogoutUser(
	ctx *fiber.Ctx,
	svc externalapi.AuthAPI,
	userID uuid.UUID,
) error {
	var (
		methodName = "LogoutUser"
		err        error
	)

	defer func(begin time.Time) {
		fields := map[string]interface{}{
			"method":     "post",
			"path":       "/v1/auth/logout",
			"methodName": methodName,
			"userID":     userID,
			"took":       time.Since(begin).String(),
		}

		l := log.Info()
		if err != nil {
			if errors.Is(err, errors.ForbiddenError()) {
				l = log.Warn().Err(err)
			} else {
				l = log.Error().Err(err)
			}
		}

		l.Fields(fields).Msg("call")
	}(time.Now())

	err = svc.LogoutUser(ctx.UserContext(), userID)
	if err != nil {
		sendResponse(ctx, log.Logger, nil, err)
		return nil
	}

	sendResponse(ctx, log.Logger, nil, nil)
	return nil
}

func ChangePassword(
	ctx *fiber.Ctx,
	svc externalapi.AuthAPI,
	userID uuid.UUID,
	oldPassword string,
	newPassword string,
) error {
	var (
		methodName = "ChangePassword"
		err        error
	)

	defer func(begin time.Time) {
		fields := map[string]interface{}{
			"method":     "post",
			"path":       "/v1/auth/change-password",
			"methodName": methodName,
			"userID":     userID,
			"took":       time.Since(begin).String(),
		}

		l := log.Info()
		if err != nil {
			if errors.Is(err, errors.ForbiddenError()) {
				l = log.Warn().Err(err)
			} else {
				l = log.Error().Err(err)
			}
		}

		l.Fields(fields).Msg("call")
	}(time.Now())

	err = svc.ChangePassword(ctx.UserContext(), userID, oldPassword, newPassword)
	if err != nil {
		sendResponse(ctx, log.Logger, nil, err)
		return nil
	}

	sendResponse(ctx, log.Logger, nil, nil)
	return nil
}

func VerifyEmailCode(
	ctx *fiber.Ctx,
	svc externalapi.AuthAPI,
	userID uuid.UUID,
	code int64,
) error {
	var (
		methodName = "VerifyEmailCode"
		err        error
	)

	defer func(begin time.Time) {
		fields := map[string]interface{}{
			"method":     "post",
			"path":       "/v1/auth/verify-email",
			"methodName": methodName,
			"userID":     userID,
			"code":       code,
			"took":       time.Since(begin).String(),
		}

		l := log.Info()
		if err != nil {
			if errors.Is(err, errors.ForbiddenError()) {
				l = log.Warn().Err(err)
			} else {
				l = log.Error().Err(err)
			}
		}

		l.Fields(fields).Msg("call")
	}(time.Now())

	err = svc.VerifyEmailCode(ctx.UserContext(), userID, code)
	if err != nil {
		sendResponse(ctx, log.Logger, nil, err)
		return nil
	}

	sendResponse(ctx, log.Logger, nil, nil)
	return nil
}

func SendEmailVerification(
	ctx *fiber.Ctx,
	svc externalapi.AuthAPI,
	userID uuid.UUID,
) error {
	var (
		methodName = "SendEmailVerification"
		err        error
	)

	defer func(begin time.Time) {
		fields := map[string]interface{}{
			"method":     "post",
			"path":       "/v1/auth/send-verification",
			"methodName": methodName,
			"userID":     userID,
			"took":       time.Since(begin).String(),
		}

		l := log.Info()
		if err != nil {
			if errors.Is(err, errors.ForbiddenError()) {
				l = log.Warn().Err(err)
			} else {
				l = log.Error().Err(err)
			}
		}

		l.Fields(fields).Msg("call")
	}(time.Now())

	err = svc.SendEmailVerification(ctx.UserContext(), userID)
	if err != nil {
		sendResponse(ctx, log.Logger, nil, err)
		return nil
	}

	sendResponse(ctx, log.Logger, nil, nil)
	return nil
}

func RequestPasswordReset(ctx *fiber.Ctx, svc externalapi.AuthAPI, email string) error {
	var (
		methodName = "RequestPasswordReset"
		err        error
	)

	defer func(begin time.Time) {
		fields := map[string]interface{}{
			"method":     "post",
			"path":       "/v1/auth/password/reset/request",
			"methodName": methodName,
			"email":      email,
			"took":       time.Since(begin).String(),
		}
		l := log.Info()
		if err != nil {
			l = log.Error().Err(err)
		}
		l.Fields(fields).Msg("call")
	}(time.Now())

	emailSent, err := svc.RequestPasswordReset(ctx.UserContext(), email)
	if err != nil {
		sendResponse(ctx, log.Logger, nil, err)
		return nil
	}

	sendResponse(ctx, log.Logger, map[string]bool{"emailSent": emailSent}, nil)
	return nil
}

// ConfirmPasswordReset intentionally never logs the raw token or the new
// password — the token alone is enough to take over the account until it
// expires, same sensitivity class as a password.
func ConfirmPasswordReset(ctx *fiber.Ctx, svc externalapi.AuthAPI, token string, newPassword string) error {
	var (
		methodName = "ConfirmPasswordReset"
		err        error
	)

	defer func(begin time.Time) {
		fields := map[string]interface{}{
			"method":     "post",
			"path":       "/v1/auth/password/reset/confirm",
			"methodName": methodName,
			"took":       time.Since(begin).String(),
		}
		l := log.Info()
		if err != nil {
			if errors.Is(err, errors.TokenInvalidError()) || errors.Is(err, errors.TokenExpiredError()) {
				l = log.Warn().Err(err)
			} else {
				l = log.Error().Err(err)
			}
		}
		l.Fields(fields).Msg("call")
	}(time.Now())

	err = svc.ConfirmPasswordReset(ctx.UserContext(), token, newPassword)
	if err != nil {
		sendResponse(ctx, log.Logger, nil, err)
		return nil
	}

	sendResponse(ctx, log.Logger, true, nil)
	return nil
}
