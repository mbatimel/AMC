package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	accesserrors "github.com/mbatimel/AMC/access/internal/errors"
	"github.com/mbatimel/AMC/access/pkg/models"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetUserRoleCode(ctx context.Context, userID uuid.UUID) (int, error) {
	var roleCode int
	err := r.db.QueryRowContext(ctx, `
		SELECT roles.code
		FROM user_roles
		JOIN roles ON roles.id = user_roles.role_id
		WHERE user_roles.user_id = $1
		ORDER BY roles.code ASC
		LIMIT 1
	`, userID).Scan(&roleCode)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, accesserrors.ErrUserRoleNotFound.AddCause("userID", userID.String())
	}
	if err != nil {
		return 0, fmt.Errorf("get user role code: %w", err)
	}
	return roleCode, nil
}

func (r *Repository) GetRoleByCode(ctx context.Context, roleCode int) (models.Role, error) {
	var role models.Role
	err := r.db.QueryRowContext(ctx, `
		SELECT id, code, COALESCE(name, ''), COALESCE(description, ''), created_at, updated_at
		FROM roles
		WHERE code = $1
	`, roleCode).Scan(
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

func (r *Repository) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	roleCode, err := r.GetUserRoleCode(ctx, userID)
	if errors.Is(err, accesserrors.ErrUserRoleNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return roleCode == int(models.RoleCodeAdmin), nil
}

func (r *Repository) AddUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error {
	role, err := r.GetRoleByCode(ctx, roleCode)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, userID, role.ID)
	if err != nil {
		return fmt.Errorf("add user role: %w", err)
	}
	return nil
}

func (r *Repository) UpdateUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error {
	role, err := r.GetRoleByCode(ctx, roleCode)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update user role transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		DELETE FROM user_roles
		WHERE user_id = $1
	`, userID)
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

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
	`, userID, role.ID); err != nil {
		return fmt.Errorf("insert updated user role: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit update user role transaction: %w", err)
	}
	return nil
}

func (r *Repository) DeleteUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error {
	role, err := r.GetRoleByCode(ctx, roleCode)
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM user_roles
		WHERE user_id = $1 AND role_id = $2
	`, userID, role.ID)
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
