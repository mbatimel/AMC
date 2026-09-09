package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	accesserrors "github.com/mbatimel/AMC/access/internal/errors"
	"github.com/mbatimel/AMC/access/pkg/models"
)

type stubRepository struct {
	role            models.Role
	roleErr         error
	userRoleCode    int
	userRoleCodeErr error
}

func (r *stubRepository) GetUserRoleCode(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.userRoleCode, r.userRoleCodeErr
}

func (r *stubRepository) GetRoleByCode(ctx context.Context, roleCode int) (models.Role, error) {
	return r.role, r.roleErr
}

func (r *stubRepository) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	return false, nil
}

func (r *stubRepository) AddUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error {
	return nil
}

func (r *stubRepository) UpdateUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error {
	return nil
}

func (r *stubRepository) DeleteUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error {
	return nil
}

func TestCheckAccess_UserHasNoRole_ReturnsNotAllowedWithoutError(t *testing.T) {
	repo := &stubRepository{userRoleCodeErr: accesserrors.ErrUserRoleNotFound.AddCause("userID", "test")}
	svc := New(repo)

	allowed, err := svc.CheckAccess(context.Background(), uuid.New(), 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if allowed {
		t.Fatal("expected allowed=false for user with no role")
	}
}

func TestCheckAccess_UserRoleSufficient_ReturnsAllowed(t *testing.T) {
	repo := &stubRepository{userRoleCode: 1}
	svc := New(repo)

	allowed, err := svc.CheckAccess(context.Background(), uuid.New(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("expected allowed=true when actual role code <= required role")
	}
}
