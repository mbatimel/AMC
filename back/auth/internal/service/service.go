package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

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
	CounterpartyINNExists(ctx context.Context, inn string) (bool, error)
	CreateIPUser(
		ctx context.Context,
		email, passwordHash string,
		surename, name, middleName string,
		fullName, shortName, inn, kpp, ogrn, okved, taxSystem, legalAddress, actualAddress,
		directorFullName, directorPosition, phone, additionalPhone, website,
		bankAccount, bankName, bankBik, correspondentAccount *string,
		requisitesFileURL, requisitesFileName string,
		roleCode int,
	) (uuid.UUID, error)
	CreateIndividualUser(
		ctx context.Context,
		email, passwordHash, surename, name, middleName, phone, city, deliveryAddress string,
		inn *string,
		roleCode int,
	) (uuid.UUID, error)
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	InvalidateUserPasswordResetTokens(ctx context.Context, userID uuid.UUID) error
	GetPasswordResetToken(ctx context.Context, tokenHash string) (postgres.PasswordResetToken, error)
}

// Mailer is implemented by internal/mailer.SMTPMailer.
type Mailer interface {
	Send(ctx context.Context, to string, subject string, body string) error
}

// AccessClient is implemented by internal/access.Client.
type AccessClient interface {
	CheckAccess(ctx context.Context, userID uuid.UUID, role int) (allowed bool, err error)
}

// FnsClient is implemented by internal/client/fns.Client.
type FnsClient interface {
	CheckIndividual(ctx context.Context, inn string) (valid bool, err error)
}

// ObjectStorage is implemented by back/objectstorage.Client.
type ObjectStorage interface {
	Upload(ctx context.Context, objectKey string, body io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, objectKey string) error
	URL(objectKey string) string
}

// RequisitesFile is the raw bytes of an uploaded requisites document, extracted
// from the multipart request by the transport layer.
type RequisitesFile struct {
	FileName string
	Content  []byte
}

// requisitesFileTypes maps an accepted lowercase file extension to the
// content-type stored alongside the S3 object.
var requisitesFileTypes = map[string]string{
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
}

func requisitesFileExtensionAndType(fileName string) (extension, contentType string, err error) {
	extension = strings.ToLower(filepath.Ext(fileName))
	contentType, ok := requisitesFileTypes[extension]
	if !ok {
		return "", "", customErrors.RequisitesFileInvalidTypeError()
	}
	return extension, contentType, nil
}

type Option func(*service)

// WithObjectStorage wires the S3-compatible client used to store uploaded
// requisites files, and the max accepted file size in bytes.
func WithObjectStorage(storage ObjectStorage, maxFileSize int64) Option {
	return func(s *service) {
		s.objectStorage = storage
		s.maxFileSize = maxFileSize
	}
}

// WithPasswordReset wires the mailer used to send reset links, the public
// front-end base URL the link points at, and how long a token stays valid.
func WithPasswordReset(mailer Mailer, publicFrontBaseURL string, resetTokenTTL time.Duration) Option {
	return func(s *service) {
		s.mailer = mailer
		s.publicFrontBaseURL = publicFrontBaseURL
		s.resetTokenTTL = resetTokenTTL
	}
}

type service struct {
	logger             zerolog.Logger
	storage            Storage
	accessClient       AccessClient
	fnsClient          FnsClient
	objectStorage      ObjectStorage
	maxFileSize        int64
	mailer             Mailer
	publicFrontBaseURL string
	resetTokenTTL      time.Duration
}

func NewAuthApiService(logger zerolog.Logger, storage Storage, accessClient AccessClient, fnsClient FnsClient, options ...Option) *service {
	s := &service{
		logger:       logger,
		storage:      storage,
		accessClient: accessClient,
		fnsClient:    fnsClient,
	}
	for _, option := range options {
		option(s)
	}
	return s
}

var _ externalAPI.AuthAPI = (*service)(nil)

func (s *service) LoginUser(ctx context.Context, email string, password string) (userID uuid.UUID, err error) {
	user, err := s.storage.GetUserByEmail(ctx, email)
	if errors.Is(err, postgres.ErrUserNotFound) {
		return uuid.Nil, customErrors.InvalidCredentialsError()
	}
	if err != nil {
		return uuid.Nil, customErrors.InternalServerError().SetOuterError(err)
	}

	if !user.IsActive {
		return uuid.Nil, customErrors.UserBlockedError(
			user.BlockedReason.String, user.BlockedContactName.String,
			user.BlockedContactPhone.String, user.BlockedContactEmail.String,
		)
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
	shortName string,
	inn string,
	directorFullName string,
	phone string,
	file RequisitesFile,
) (userID uuid.UUID, err error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return uuid.Nil, customErrors.ValidationError("email")
	}
	if strings.TrimSpace(password) == "" {
		return uuid.Nil, customErrors.ValidationError("password")
	}
	if strings.TrimSpace(shortName) == "" {
		return uuid.Nil, customErrors.ValidationError("shortName")
	}
	if strings.TrimSpace(directorFullName) == "" {
		return uuid.Nil, customErrors.ValidationError("directorFullName")
	}
	if strings.TrimSpace(phone) == "" {
		return uuid.Nil, customErrors.ValidationError("phone")
	}
	if strings.TrimSpace(inn) == "" {
		return uuid.Nil, customErrors.InnEmptyErr("inn")
	}
	if len(file.Content) == 0 {
		return uuid.Nil, customErrors.RequisitesFileRequiredError()
	}
	if s.objectStorage == nil {
		return uuid.Nil, customErrors.InternalServerError().SetOuterError(fmt.Errorf("object storage not configured"))
	}
	if s.maxFileSize > 0 && int64(len(file.Content)) > s.maxFileSize {
		return uuid.Nil, customErrors.RequisitesFileTooLargeError()
	}
	extension, contentType, err := requisitesFileExtensionAndType(file.FileName)
	if err != nil {
		return uuid.Nil, err
	}

	valid, err := validate(inn)
	if err != nil {
		return uuid.Nil, err
	}
	if !valid {
		return uuid.Nil, fmt.Errorf("Inn not valid:%w", err)
	}
	innExists, err := s.storage.CounterpartyINNExists(ctx, inn)
	if err != nil {
		return uuid.Nil, customErrors.InternalServerError().SetOuterError(err)
	}
	if innExists {
		return uuid.Nil, customErrors.InnTakenError(inn)
	}
	fnsValid, err := s.fnsClient.CheckIndividual(ctx, inn)
	if err != nil {
		return uuid.Nil, customErrors.InternalServerError().SetOuterError(err)
	}
	if !fnsValid {
		return uuid.Nil, customErrors.InnInvalidError(inn)
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, customErrors.InternalServerError().SetOuterError(err)
	}
	surename, name, middleName := splitFio(directorFullName)

	objectKey := fmt.Sprintf("counterparties/requisites/%s%s", uuid.NewString(), extension)
	if err = s.objectStorage.Upload(ctx, objectKey, bytes.NewReader(file.Content), int64(len(file.Content)), contentType); err != nil {
		return uuid.Nil, customErrors.InternalServerError().SetOuterError(err)
	}
	fileURL := s.objectStorage.URL(objectKey)

	userID, err = s.storage.CreateIPUser(
		ctx, email, string(passwordHash),
		surename, name, middleName,
		nil, &shortName, &inn, nil, nil, nil, nil, nil, nil,
		&directorFullName, nil, &phone, nil, nil,
		nil, nil, nil, nil,
		fileURL, file.FileName,
		defaultSignUpRoleCode,
	)
	if err != nil {
		if cleanupErr := s.objectStorage.Delete(ctx, objectKey); cleanupErr != nil {
			s.logger.Error().Err(cleanupErr).Str("objectKey", objectKey).Msg("failed to compensate requisites file upload")
		}
		if errors.Is(err, postgres.ErrEmailTaken) {
			return uuid.Nil, customErrors.EmailTakenError()
		}
		if errors.Is(err, postgres.ErrInnTaken) {
			return uuid.Nil, customErrors.InnTakenError(inn)
		}
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
