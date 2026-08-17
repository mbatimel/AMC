package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	customErrors "github.com/mbatimel/AMC/auth/internal/errors"
	"github.com/mbatimel/AMC/auth/internal/storage/postgres"
	externalAPI "github.com/mbatimel/AMC/auth/pkg/interfaces/externalAPI"
)

// Storage is implemented by internal/storage/postgres.Storage.
type Storage interface {
	GetUserByEmail(ctx context.Context, email string) (postgres.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (postgres.User, error)
	CreateIPUser(
		ctx context.Context,
		email, passwordHash string,
		fullName, shortName, inn, kpp, ogrn, okved, taxSystem, legalAddress, actualAddress,
		directorFullName, directorPosition, phone, additionalPhone, website,
		bankAccount, bankName, bankBik, correspondentAccount *string,
		roleCode int,
	) (uuid.UUID, error)
	CreateIndividualUser(
		ctx context.Context,
		email, passwordHash, surename, name, middleName, phone, city, deliveryAddress string,
		inn *string,
		roleCode int,
	) (uuid.UUID, error)
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
}

// AccessClient is implemented by internal/access.Client.
type AccessClient interface {
	CheckAccess(ctx context.Context, userID uuid.UUID, role int) (allowed bool, err error)
}

// FnsClient is implemented by internal/client/fns.Client.
type FnsClient interface {
	CheckIndividual(ctx context.Context, inn string) (valid bool, err error)
}

type service struct {
	logger       zerolog.Logger
	storage      Storage
	accessClient AccessClient
	fnsClient    FnsClient
}

func NewAuthApiService(logger zerolog.Logger, storage Storage, accessClient AccessClient, fnsClient FnsClient) externalAPI.AuthAPI {
	return &service{
		logger:       logger,
		storage:      storage,
		accessClient: accessClient,
		fnsClient:    fnsClient,
	}
}

func (s *service) LoginUser(ctx context.Context, email string, password string) (userID uuid.UUID, err error) {
	user, err := s.storage.GetUserByEmail(ctx, email)
	if errors.Is(err, postgres.ErrUserNotFound) {
		return uuid.Nil, customErrors.InvalidCredentialsError()
	}
	if err != nil {
		return uuid.Nil, customErrors.InternalServerError().SetOuterError(err)
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); compareErr != nil {
		return uuid.Nil, customErrors.InvalidCredentialsError()
	}

	return user.ID, nil
}

func (s *service) RegisterIP(
	ctx context.Context,
	email string,
	password string,
	fullName, shortName, inn, kpp, ogrn, okved, taxSystem, legalAddress, actualAddress,
	directorFullName, directorPosition, phone, additionalPhone, website,
	bankAccount, bankName, bankBik, correspondentAccount *string,
) (userID uuid.UUID, err error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return uuid.Nil, customErrors.ValidationError("email")
	}
	if strings.TrimSpace(password) == "" {
		return uuid.Nil, customErrors.ValidationError("password")
	}
	if inn == nil{
		return uuid.Nil,customErrors.InnEmptyErr("inn")
	}
	valid, err := validate(*inn)
	if  err !=nil{
		return uuid.Nil,err
	}
	if !valid{
		return uuid.Nil,fmt.Errorf("Inn not valid:%w",err)
	}
	fnsValid, err := s.fnsClient.CheckIndividual(ctx, *inn)
	if err != nil {
		return uuid.Nil, customErrors.InternalServerError().SetOuterError(err)
	}
	if !fnsValid {
		return uuid.Nil, customErrors.InnInvalidError(*inn)
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, customErrors.InternalServerError().SetOuterError(err)
	}
	userID, err = s.storage.CreateIPUser(
		ctx, email, string(passwordHash),
		fullName, shortName, inn, kpp, ogrn, okved, taxSystem, legalAddress, actualAddress,
		directorFullName, directorPosition, phone, additionalPhone, website,
		bankAccount, bankName, bankBik, correspondentAccount,
		defaultSignUpRoleCode,
	)
	if errors.Is(err, postgres.ErrEmailTaken) {
		return uuid.Nil, customErrors.EmailTakenError()
	}
	if err != nil {
		return uuid.Nil, customErrors.InternalServerError().SetOuterError(err)
	}

	return userID, nil
}


// splitFio splits a Russian "Фамилия Имя Отчество" string into its three parts.
// Fewer than two words leaves name/middleName empty.
func splitFio(fio string) (surename, name, middleName string) {
	parts := strings.Fields(fio)
	switch len(parts) {
	case 0:
		return "", "", ""
	case 1:
		return parts[0], "", ""
	case 2:
		return parts[0], parts[1], ""
	default:
		return parts[0], parts[1], strings.Join(parts[2:], " ")
	}
}

// LogoutUser is a no-op: this service does not keep server-side session state,
// the caller is only expected to discard its own userID.
func (s *service) LogoutUser(ctx context.Context, userID uuid.UUID) (err error) {
	return nil
}

func (s *service) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) (err error) {
	user, err := s.storage.GetUserByID(ctx, userID)
	if errors.Is(err, postgres.ErrUserNotFound) {
		return customErrors.NotFoundError()
	}
	if err != nil {
		return customErrors.InternalServerError().SetOuterError(err)
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); compareErr != nil {
		return customErrors.InvalidCredentialsError()
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return customErrors.InternalServerError().SetOuterError(err)
	}

	if err = s.storage.UpdateUserPassword(ctx, userID, string(newPasswordHash)); err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			return customErrors.NotFoundError()
		}
		return customErrors.InternalServerError().SetOuterError(err)
	}

	return nil
}

// VerifyEmailCode is a stub: always succeeds, no verification codes are issued or checked.
func (s *service) VerifyEmailCode(ctx context.Context, userID uuid.UUID, code int64) (err error) {
	return nil
}

// SendEmailVerification is a stub: no email is sent, no code is generated.
func (s *service) SendEmailVerification(ctx context.Context, userID uuid.UUID) (err error) {
	return nil
}
