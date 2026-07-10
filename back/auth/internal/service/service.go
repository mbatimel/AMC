package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog"

	customerrors "github.com/mbatimel/AMC/auth/internal/errors"
	"github.com/mbatimel/AMC/auth/internal/models"
	"github.com/mbatimel/AMC/auth/internal/repo"
	"github.com/mbatimel/AMC/auth/internal/storage/postgres"
	"github.com/mbatimel/AMC/auth/internal/storage/redis"
	externalapi "github.com/mbatimel/AMC/auth/pkg/interfaces/externalAPI"
)

const verificationCodeTTL = 5 * time.Minute

type authService struct {
	storage postgres.Storage
	cache   redis.ICacheRepository
	logger  zerolog.Logger
}

func NewAuthService(storage postgres.Storage, cache redis.ICacheRepository, logger zerolog.Logger) externalapi.AuthAPI {
	return &authService{
		storage: storage,
		cache:   cache,
		logger:  logger,
	}
}

func (a *authService) LoginUser(ctx context.Context, email string, password string) (userID uuid.UUID, err error) {
	email = repo.NormalizeEmail(email)
	if err = repo.ValidateEmail(email); err != nil {
		return uuid.Nil, customerrors.ErrInvalidEmailError()
	}

	user, err := a.storage.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Тот же ответ, что и при неверном пароле — не раскрываем, существует ли аккаунт.
			return uuid.Nil, customerrors.ErrInvalidPasswordError()
		}
		a.logger.Error().Err(err).Str("email", email).Msg("get user by email")
		return uuid.Nil, customerrors.InternalServerError()
	}

	if !repo.CheckPassword(password, user.Password) {
		return uuid.Nil, customerrors.ErrInvalidPasswordError()
	}

	return user.ID, nil
}

func (a *authService) SignUpUser(ctx context.Context, email string, password string, name string, surename string) (userID uuid.UUID, err error) {
	email = repo.NormalizeEmail(email)
	if err = repo.ValidateEmail(email); err != nil {
		return uuid.Nil, customerrors.ErrInvalidEmailError()
	}

	_, err = a.storage.GetUserByEmail(ctx, email)
	if err == nil {
		return uuid.Nil, customerrors.ErrEmailExistsError()
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		a.logger.Error().Err(err).Str("email", email).Msg("check existing user")
		return uuid.Nil, customerrors.InternalServerError()
	}

	hashedPassword, err := repo.HashPassword(password)
	if err != nil {
		a.logger.Error().Err(err).Msg("hash password")
		return uuid.Nil, customerrors.InternalServerError()
	}

	userID, err = a.storage.CreateUser(ctx, &models.User{
		Email:    email,
		Password: hashedPassword,
		Name:     name,
		Surename: surename,
		Status:   models.UserStatusPending,
	})
	if err != nil {
		a.logger.Error().Err(err).Str("email", email).Msg("create user")
		return uuid.Nil, customerrors.InternalServerError()
	}

	return userID, nil
}

func (a *authService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) (err error) {
	user, err := a.storage.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return customerrors.ErrUserNotFoundError()
		}
		a.logger.Error().Err(err).Str("userID", userID.String()).Msg("get user by id")
		return customerrors.InternalServerError()
	}

	if !repo.CheckPassword(oldPassword, user.Password) {
		return customerrors.ErrInvalidPasswordError()
	}

	hashedPassword, err := repo.HashPassword(newPassword)
	if err != nil {
		a.logger.Error().Err(err).Msg("hash password")
		return customerrors.InternalServerError()
	}

	if err = a.storage.UpdateUserPassword(ctx, userID, hashedPassword); err != nil {
		a.logger.Error().Err(err).Str("userID", userID.String()).Msg("update user password")
		return customerrors.InternalServerError()
	}

	return nil
}

func (a *authService) SendEmailVerification(ctx context.Context, userID uuid.UUID) (err error) {
	if _, err = a.storage.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return customerrors.ErrUserNotFoundError()
		}
		a.logger.Error().Err(err).Str("userID", userID.String()).Msg("get user by id")
		return customerrors.InternalServerError()
	}

	code, err := repo.GenerateVerificationCode()
	if err != nil {
		a.logger.Error().Err(err).Msg("generate verification code")
		return customerrors.InternalServerError()
	}

	if err = a.cache.SaveVerificationCode(ctx, userID.String(), strconv.FormatInt(code, 10), verificationCodeTTL); err != nil {
		a.logger.Error().Err(err).Str("userID", userID.String()).Msg("save verification code")
		return customerrors.InternalServerError()
	}

	// TODO: подключить реальную отправку письма (email-сервиса в проекте пока нет). Код в лог не пишем — попадает только в кэш.
	a.logger.Info().Str("userID", userID.String()).Msg("email verification code generated")

	return nil
}

func (a *authService) VerifyEmailCode(ctx context.Context, userID uuid.UUID, code int64) (err error) {
	storedCode, err := a.cache.GetVerificationCode(ctx, userID.String())
	if err != nil {
		return err
	}

	if storedCode != strconv.FormatInt(code, 10) {
		return customerrors.ErrInvalidCodeError()
	}

	if err = a.cache.DeleteVerificationCode(ctx, userID.String()); err != nil {
		a.logger.Error().Err(err).Str("userID", userID.String()).Msg("delete verification code")
		return customerrors.InternalServerError()
	}

	if err = a.storage.UpdateUserStatus(ctx, userID, models.UserStatusActive); err != nil {
		a.logger.Error().Err(err).Str("userID", userID.String()).Msg("activate user")
		return customerrors.InternalServerError()
	}

	return nil
}
