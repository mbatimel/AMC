// back/admin/internal/service/service.go
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/internal/storage/postgres"
	externalapi "github.com/mbatimel/AMC/admin/pkg/interfaces/externalapi"
	"github.com/mbatimel/AMC/admin/pkg/models"
	authTransport "github.com/mbatimel/AMC/auth/pkg/client/transport"
	fasthttp "github.com/valyala/fasthttp"
)

// RoleCodeAdmin matches the "admin" role seeded in the shared roles table
// (back/migrations/pkg/migrations/data/20260705171948_access_roles.sql).
const RoleCodeAdmin = 0

const actorLabelAdmin = "Админ портала"

const (
	defaultAuditLogLimit = 50
	maxAuditLogLimit     = 100
)

// Storage is implemented by internal/storage/postgres.Storage.
type Storage interface {
	InsertAuditLogEntry(ctx context.Context, actorUserID uuid.UUID, actorLabel string, action string) error
	ListAuditLogEntries(ctx context.Context, limit int, offset int) ([]postgres.AuditLogEntry, error)
	CountAuditLogEntries(ctx context.Context) (int, error)
	ListBanners(ctx context.Context, visibleOnly bool) ([]postgres.Banner, int, error)
	GetBanner(ctx context.Context, bannerID uuid.UUID) (postgres.Banner, error)
	CreateBanner(ctx context.Context, banner postgres.Banner) (postgres.Banner, error)
	UpdateBanner(ctx context.Context, banner postgres.Banner, replaceImage bool) (postgres.Banner, error)
	DeleteBanner(ctx context.Context, bannerID uuid.UUID) error
	UpdateBannerDelay(ctx context.Context, delaySec int) error
	CreateSignupRequest(ctx context.Context, request postgres.SignupRequest) (postgres.SignupRequest, error)
	ListSignupRequests(ctx context.Context, status string) ([]postgres.SignupRequest, error)
	DecideSignupRequest(ctx context.Context, id uuid.UUID, status string, rejectReason string) (postgres.SignupRequest, error)
	ListLegalDocs(ctx context.Context) ([]postgres.LegalDoc, error)
	GetLegalDoc(ctx context.Context, docID string) (postgres.LegalDoc, error)
	CreateLegalDoc(ctx context.Context, doc postgres.LegalDoc, version postgres.LegalDocVersion) (postgres.LegalDoc, error)
	ReplaceLegalDocFile(ctx context.Context, docID string, version postgres.LegalDocVersion) (postgres.LegalDoc, error)
	ListLegalDocVersions(ctx context.Context, docID string) ([]postgres.LegalDocVersion, error)
	DeleteLegalDoc(ctx context.Context, docID string) error
	ListCertificates(ctx context.Context, visibleOnly bool) ([]postgres.Certificate, error)
	GetCertificate(ctx context.Context, certID uuid.UUID) (postgres.Certificate, error)
	CreateCertificate(ctx context.Context, cert postgres.Certificate) (postgres.Certificate, error)
	UpdateCertificate(ctx context.Context, cert postgres.Certificate, replaceFile bool) (postgres.Certificate, error)
	DeleteCertificate(ctx context.Context, certID uuid.UUID) error
}

// AuthClient is implemented by auth/pkg/client/transport.ClientAuthAPI.
type AuthClient interface {
	LoginUser(ctx context.Context, email string, password string) (userID uuid.UUID, err error)
}

// AccessClient is implemented by access/pkg/client/transport.ClientAccessAPI.
type AccessClient interface {
	CheckAccess(ctx context.Context, userID uuid.UUID, role int) (allowed bool, err error)
}

// UsersClient is implemented by internal/client/users.Client.
type UsersClient interface {
	FindUserIDByEmail(ctx context.Context, email string) (userID uuid.UUID, found bool, err error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

// Mailer is implemented by internal/mailer.SMTPMailer.
type Mailer interface {
	Send(ctx context.Context, to string, subject string, body string) error
	SendWithAttachments(ctx context.Context, to string, subject string, body string, attachments []models.ImageFile) error
}

type service struct {
	logger                  zerolog.Logger
	storage                 Storage
	authClient              AuthClient
	accessClient            AccessClient
	usersClient             UsersClient
	mailer                  Mailer
	objectStorage           ObjectStorage
	maxFileSize             int64
	companyRequestRecipient string
}

func WithCompanyRequestRecipient(recipient string) Option {
	return func(s *service) {
		s.companyRequestRecipient = recipient
	}
}

var _ externalapi.AdminAPI = (*service)(nil)

func NewAdminApiService(logger zerolog.Logger, storage Storage, authClient AuthClient, accessClient AccessClient, usersClient UsersClient, mailer Mailer, options ...Option) *service {
	result := &service{
		logger:       logger,
		storage:      storage,
		authClient:   authClient,
		accessClient: accessClient,
		usersClient:  usersClient,
		mailer:       mailer,
	}
	for _, option := range options {
		option(result)
	}
	return result
}

// requireAdmin checks that userID currently holds the admin role. Every
// AdminAPI method other than Login must call this before doing anything else.
func (s *service) requireAdmin(ctx context.Context, userID uuid.UUID) error {
	allowed, err := s.accessClient.CheckAccess(ctx, userID, RoleCodeAdmin)
	if err != nil {
		return customErrors.InternalServerError().SetOuterError(err)
	}
	if !allowed {
		return customErrors.ForbiddenError()
	}
	return nil
}

func (s *service) Login(ctx context.Context, email string, password string) (response models.LoginResponse, err error) {
	userID, err := s.authClient.LoginUser(ctx, email, password)
	if err != nil {
		var loginErr *authTransport.LoginError
		if errors.As(err, &loginErr) && loginErr.StatusCode == fasthttp.StatusUnauthorized {
			return models.LoginResponse{}, customErrors.InvalidCredentialsError().SetOuterError(err)
		}
		return models.LoginResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}

	if err = s.requireAdmin(ctx, userID); err != nil {
		return models.LoginResponse{}, err
	}

	if err = s.storage.InsertAuditLogEntry(ctx, userID, actorLabelAdmin, "Выполнен вход в систему"); err != nil {
		s.logger.Error().Err(err).Msg("failed to write audit log entry")
	}

	return models.LoginResponse{UserID: userID, Role: "admin"}, nil
}

func (s *service) Logout(ctx context.Context, userID uuid.UUID) (err error) {
	if err = s.requireAdmin(ctx, userID); err != nil {
		return err
	}

	if err = s.storage.InsertAuditLogEntry(ctx, userID, actorLabelAdmin, "Выполнен выход из системы"); err != nil {
		s.logger.Error().Err(err).Msg("failed to write audit log entry")
	}

	return nil
}

func (s *service) GetSession(ctx context.Context, userID uuid.UUID) (response models.SessionResponse, err error) {
	if err = s.requireAdmin(ctx, userID); err != nil {
		return models.SessionResponse{}, err
	}
	return models.SessionResponse{UserID: userID, Role: "admin"}, nil
}

func (s *service) ListAuditLog(ctx context.Context, userID uuid.UUID, limit int, offset int) (response models.ListAuditLogResponse, err error) {
	if err = s.requireAdmin(ctx, userID); err != nil {
		return models.ListAuditLogResponse{}, err
	}

	if limit <= 0 || limit > maxAuditLogLimit {
		limit = defaultAuditLogLimit
	}
	if offset < 0 {
		offset = 0
	}

	entries, err := s.storage.ListAuditLogEntries(ctx, limit, offset)
	if err != nil {
		return models.ListAuditLogResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}

	total, err := s.storage.CountAuditLogEntries(ctx)
	if err != nil {
		return models.ListAuditLogResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}

	items := make([]models.AuditLogEntry, 0, len(entries))
	for _, e := range entries {
		items = append(items, models.AuditLogEntry{
			ID:         e.ID,
			CreatedAt:  e.CreatedAt,
			ActorLabel: e.ActorLabel,
			Action:     e.Action,
		})
	}

	return models.ListAuditLogResponse{Items: items, Total: total}, nil
}
