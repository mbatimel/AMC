package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/mbatimel/AMC/access/pkg/models"
)

type Repository struct {
	storage Storage
}

type Storage interface {
	GetUserRoleCode(ctx context.Context, userID uuid.UUID) (int, error)
	GetRoleByCode(ctx context.Context, roleCode int) (models.Role, error)
	IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
	AddUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error
	UpdateUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error
	DeleteUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error
}

func New(storage Storage) *Repository {
	return &Repository{storage: storage}
}

func (r *Repository) GetUserRoleCode(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.storage.GetUserRoleCode(ctx, userID)
}

func (r *Repository) GetRoleByCode(ctx context.Context, roleCode int) (models.Role, error) {
	return r.storage.GetRoleByCode(ctx, roleCode)
}

func (r *Repository) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	return r.storage.IsAdmin(ctx, userID)
}

func (r *Repository) AddUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error {
	return r.storage.AddUserRole(ctx, userID, roleCode)
}

func (r *Repository) UpdateUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error {
	return r.storage.UpdateUserRole(ctx, userID, roleCode)
}

func (r *Repository) DeleteUserRole(ctx context.Context, userID uuid.UUID, roleCode int) error {
	return r.storage.DeleteUserRole(ctx, userID, roleCode)
}
