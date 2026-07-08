package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	"github.com/google/uuid"

	accesserrors "github.com/mbatimel/AMC/access/internal/errors"
	"github.com/mbatimel/AMC/access/pkg/models"
)

//go:embed sql/getUserRoleCode.sql
var sqlGetUserRoleCode string

//go:embed sql/getRoleByCode.sql
var sqlGetRoleByCode string

//go:embed sql/addUserRole.sql
var sqlAddUserRole string

//go:embed sql/deleteUserRolesByUserID.sql
var sqlDeleteUserRolesByUserID string

//go:embed sql/insertUserRole.sql
var sqlInsertUserRole string

//go:embed sql/deleteUserRole.sql
var sqlDeleteUserRole string

type Storage struct {
	db *sql.DB
}

func New(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) GetUserRoleCode(ctx context.Context, userID uuid.UUID) (int, error) {
	var roleCode int
	err := s.db.QueryRowContext(ctx, sqlGetUserRoleCode, userID).Scan(&roleCode)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, accesserrors.ErrUserRoleNotFound.AddCause("userID", userID.String())
	}
	if err != nil {
		return 0, fmt.Errorf("get user role code: %w", err)
	}
	return roleCode, nil
}

func (s *Storage) GetRoleByCode(ctx context.Context, roleCode int) (models.Role, error) {
	var role models.Role
	err := s.db.QueryRowContext(ctx, sqlGetRoleByCode, roleCode).Scan(
		&role.ID,
		&role.Code,
		&role.Name,
		&role.Description,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Role{}, accesserrors.ErrRoleNotFound.AddCause("role", roleCode)
	}
	if err != nil {
		return models.Role{}, fmt.Errorf("get role by code: %w", err)
	}
	return role, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	roleCode, err := s.GetUserRoleCode(ctx, userID)
	if errors.Is(err, accesserrors.ErrUserRoleNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return roleCode == int(models.RoleCodeAdmin), nil
}

func (s *Storage) AddUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error {
	role, err := s.GetRoleByCode(ctx, roleCode)
	if err != nil {
		return err
	}

	if _, err = s.db.ExecContext(ctx, sqlAddUserRole, userID, role.ID); err != nil {
		return fmt.Errorf("add user role: %w", err)
	}
	return nil
}

func (s *Storage) UpdateUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error {
	role, err := s.GetRoleByCode(ctx, roleCode)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update user role transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, sqlDeleteUserRolesByUserID, userID)
	if err != nil {
		return fmt.Errorf("update user role: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get updated user roles count: %w", err)
	}
	if affected == 0 {
		return accesserrors.ErrUserRoleNotFound.AddCause("userID", userID.String())
	}

	if _, err = tx.ExecContext(ctx, sqlInsertUserRole, userID, role.ID); err != nil {
		return fmt.Errorf("insert updated user role: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit update user role transaction: %w", err)
	}
	return nil
}

func (s *Storage) DeleteUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error {
	role, err := s.GetRoleByCode(ctx, roleCode)
	if err != nil {
		return err
	}

	result, err := s.db.ExecContext(ctx, sqlDeleteUserRole, userID, role.ID)
	if err != nil {
		return fmt.Errorf("delete user role: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted user roles count: %w", err)
	}
	if affected == 0 {
		return accesserrors.ErrUserRoleNotFound.AddCause("userID", userID.String(), "role", roleCode)
	}
	return nil
}
