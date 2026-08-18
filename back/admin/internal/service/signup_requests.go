// back/admin/internal/service/signup_requests.go
package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/google/uuid"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/internal/storage/postgres"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

const (
	signupStatusPending  = "pending"
	signupStatusApproved = "approved"
	signupStatusRejected = "rejected"

	signupTypeIndividual   = "individual"
	signupTypeOrganization = "organization"

	maxSignupFieldLength = 255
)

func signupValidation(field string) error {
	return customErrors.BadRequestError().AddCause("field", field)
}

func normalizeSignupType(value string) string {
	if strings.TrimSpace(value) == signupTypeIndividual {
		return signupTypeIndividual
	}
	return signupTypeOrganization
}

func normalizeSignupField(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len([]rune(value)) > maxSignupFieldLength {
		return "", signupValidation("length")
	}
	if strings.ContainsAny(value, "\r\n\x00") {
		return "", signupValidation("value")
	}
	return value, nil
}

// normalizeSignupEmail rejects anything that could inject extra headers into
// the rejection email (see internal/mailer.SMTPMailer.Send: `to` is written
// straight into a "To:" header line).
func normalizeSignupEmail(value string) (string, error) {
	value, err := normalizeSignupField(strings.ToLower(value))
	if err != nil {
		return "", err
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return "", signupValidation("email")
	}
	return value, nil
}

func modelSignupRequest(request postgres.SignupRequest) models.SignupRequest {
	return models.SignupRequest{
		ID: request.ID, Company: request.Company, INN: request.INN, Contact: request.Contact,
		Email: request.Email, Phone: request.Phone, Type: request.RequestType,
		Status: request.Status, RejectReason: request.RejectReason,
		CreatedAt: request.CreatedAt, DecidedAt: request.DecidedAt,
	}
}

func mapSignupStorageError(err error) error {
	switch {
	case errors.Is(err, postgres.ErrSignupRequestNotFound):
		return customErrors.NotFoundError()
	case errors.Is(err, postgres.ErrSignupRequestAlreadyDecided):
		return customErrors.ConflictError().AddCause("reason", "alreadyDecided")
	default:
		return customErrors.InternalServerError().SetOuterError(err)
	}
}

// CreateSignupRequest is public: anyone can submit a registration request.
func (s *service) CreateSignupRequest(ctx context.Context, company string, inn string, contact string, email string, phone string, requestType string) (models.SignupRequest, error) {
	if strings.TrimSpace(email) == "" {
		return models.SignupRequest{}, signupValidation("email")
	}
	email, err := normalizeSignupEmail(email)
	if err != nil {
		return models.SignupRequest{}, err
	}
	company, err = normalizeSignupField(company)
	if err != nil {
		return models.SignupRequest{}, err
	}
	inn, err = normalizeSignupField(inn)
	if err != nil {
		return models.SignupRequest{}, err
	}
	contact, err = normalizeSignupField(contact)
	if err != nil {
		return models.SignupRequest{}, err
	}
	phone, err = normalizeSignupField(phone)
	if err != nil {
		return models.SignupRequest{}, err
	}

	created, err := s.storage.CreateSignupRequest(ctx, postgres.SignupRequest{
		ID: uuid.New(), Company: company, INN: inn, Contact: contact,
		Email: email, Phone: phone, RequestType: normalizeSignupType(requestType),
	})
	if err != nil {
		return models.SignupRequest{}, customErrors.InternalServerError().SetOuterError(err)
	}

	s.logger.Info().Str("email", email).Msg("new signup request created")
	return modelSignupRequest(created), nil
}

func (s *service) ListSignupRequests(ctx context.Context, userID uuid.UUID, status string) (models.ListSignupRequestsResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.ListSignupRequestsResponse{}, err
	}

	status = strings.TrimSpace(status)
	if status != "" && status != signupStatusPending && status != signupStatusApproved && status != signupStatusRejected {
		return models.ListSignupRequestsResponse{}, signupValidation("status")
	}

	items, err := s.storage.ListSignupRequests(ctx, status)
	if err != nil {
		return models.ListSignupRequestsResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}

	response := models.ListSignupRequestsResponse{Items: make([]models.SignupRequest, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, modelSignupRequest(item))
	}
	return response, nil
}

// ApproveSignupRequest only marks the request decided: the applicant's
// account already exists (registration is instant, see back/auth
// RegisterIP), so there is nothing else to provision here.
func (s *service) ApproveSignupRequest(ctx context.Context, userID uuid.UUID, requestID uuid.UUID) (models.SignupRequest, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.SignupRequest{}, err
	}

	updated, err := s.storage.DecideSignupRequest(ctx, requestID, signupStatusApproved, "")
	if err != nil {
		return models.SignupRequest{}, mapSignupStorageError(err)
	}

	s.audit(ctx, userID, fmt.Sprintf("Одобрена заявка на регистрацию (%s)", updated.Email))
	return modelSignupRequest(updated), nil
}

// RejectSignupRequest marks the request decided, deletes the applicant's
// account in back/users (if one exists for that email) and emails them the
// rejection reason. A guard in storage.DecideSignupRequest prevents deciding
// an already-decided request twice.
func (s *service) RejectSignupRequest(ctx context.Context, userID uuid.UUID, requestID uuid.UUID, reason string) (models.SignupRequest, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.SignupRequest{}, err
	}

	reason = strings.TrimSpace(reason)
	updated, err := s.storage.DecideSignupRequest(ctx, requestID, signupStatusRejected, reason)
	if err != nil {
		return models.SignupRequest{}, mapSignupStorageError(err)
	}

	deletedAccount := s.deleteRejectedUserAccount(ctx, updated.Email)
	s.sendRejectionEmail(ctx, updated.Email, reason)

	action := fmt.Sprintf("Отклонена заявка на регистрацию (%s)", updated.Email)
	if deletedAccount {
		action += ", аккаунт удалён"
	}
	s.audit(ctx, userID, action)

	return modelSignupRequest(updated), nil
}

func (s *service) deleteRejectedUserAccount(ctx context.Context, email string) bool {
	if s.usersClient == nil {
		return false
	}

	accountID, found, err := s.usersClient.FindUserIDByEmail(ctx, email)
	if err != nil {
		s.logger.Error().Err(err).Str("email", email).Msg("failed to look up user account for rejected signup request")
		return false
	}
	if !found {
		return false
	}

	if err = s.usersClient.DeleteUser(ctx, accountID); err != nil {
		s.logger.Error().Err(err).Str("userID", accountID.String()).Msg("failed to delete user account for rejected signup request")
		return false
	}
	return true
}

func (s *service) sendRejectionEmail(ctx context.Context, email string, reason string) {
	if s.mailer == nil {
		return
	}

	body := "Ваша заявка на регистрацию отклонена."
	if reason != "" {
		body = fmt.Sprintf("Ваша заявка на регистрацию отклонена. Причина: %s", reason)
	}

	if err := s.mailer.Send(ctx, email, "Заявка на регистрацию отклонена", body); err != nil {
		s.logger.Error().Err(err).Str("email", email).Msg("failed to send signup rejection email")
	}
}
