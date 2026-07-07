package service

import (
	"context"

	"github.com/google/uuid"

	accesserrors "github.com/mbatimel/AMC/access/internal/errors"
	"github.com/mbatimel/AMC/access/pkg/models"
)

type Repository interface {
	GetUserRoleCode(ctx context.Context, userID uuid.UUID) (int, error)
	GetRoleByCode(ctx context.Context, roleCode int) (models.Role, error)
	IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
	AddUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error
	UpdateUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error
	DeleteUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CheckAccess(ctx context.Context, userID uuid.UUID, role int) (allowed bool, err error) {
	if _, err = s.repository.GetRoleByCode(ctx, role); err != nil {
		return false, err
	}

	actualRoleCode, err := s.repository.GetUserRoleCode(ctx, userID)
	if err != nil {
		return false, err
	}

	return actualRoleCode <= role, nil
}

func (s *Service) AddRole(ctx context.Context, adminUserID uuid.UUID, userID uuid.UUID, role int) (success bool, err error) {
	if err = s.checkAdmin(ctx, adminUserID); err != nil {
		return false, err
	}
	if err = s.repository.AddUserRole(ctx, userID, role); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) UpdateRole(ctx context.Context, adminUserID uuid.UUID, userID uuid.UUID, role int) (success bool, err error) {
	if err = s.checkAdmin(ctx, adminUserID); err != nil {
		return false, err
	}
	if err = s.repository.UpdateUserRole(ctx, userID, role); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) DeleteRole(ctx context.Context, adminUserID uuid.UUID, userID uuid.UUID, role int) (success bool, err error) {
	if err = s.checkAdmin(ctx, adminUserID); err != nil {
		return false, err
	}
	if err = s.repository.DeleteUserRole(ctx, userID, role); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) checkAdmin(ctx context.Context, adminUserID uuid.UUID) error {
	isAdmin, err := s.repository.IsAdmin(ctx, adminUserID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return accesserrors.ErrForbidden.AddCause("adminUserID", adminUserID.String())
	}
	return nil
}
