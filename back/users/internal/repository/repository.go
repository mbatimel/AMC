package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mbatimel/AMC/users/internal/storage/postgres"
)

type User = postgres.UserRow
type Client = postgres.ClientRow
type ClientConditions = postgres.ClientConditionsRow
type CategoryDiscount = postgres.CategoryDiscountRow
type Favorite = postgres.FavoriteRow
type CreateUserParams = postgres.CreateUserParams
type UpdateUserParams = postgres.UpdateUserParams
type ListUsersParams = postgres.ListUsersParams
type UpdateProfileParams = postgres.UpdateProfileParams

type Storage interface {
	CreateUser(ctx context.Context, params postgres.CreateUserParams) (postgres.UserRow, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (postgres.UserRow, error)
	GetUserByEmail(ctx context.Context, email string) (postgres.UserRow, error)
	ListUsers(ctx context.Context, params postgres.ListUsersParams) ([]postgres.UserRow, error)
	CountUsers(ctx context.Context, params postgres.ListUsersParams) (int, error)
	UpdateUser(ctx context.Context, params postgres.UpdateUserParams) (postgres.UserRow, error)
	SoftDeleteUser(ctx context.Context, userID uuid.UUID) error
	SetUserActive(ctx context.Context, userID uuid.UUID, active bool) (postgres.UserRow, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (postgres.UserRow, *postgres.ClientRow, error)
	UpdateProfile(ctx context.Context, params postgres.UpdateProfileParams) (postgres.UserRow, *postgres.ClientRow, error)
	ListUserClients(ctx context.Context, userID uuid.UUID) ([]postgres.ClientRow, error)
	UserHasClient(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (bool, error)
	GetClientDetails(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (postgres.ClientRow, error)
	GetClientConditions(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (postgres.ClientConditionsRow, error)
	SetActiveClient(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (postgres.ClientRow, error)
	GetActiveClient(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	ListFavorites(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) ([]postgres.FavoriteRow, error)
	AddFavorite(ctx context.Context, userID uuid.UUID, clientID uuid.UUID, productID uuid.UUID) (postgres.FavoriteRow, bool, error)
	DeleteFavorites(ctx context.Context, userID uuid.UUID, clientID uuid.UUID, productIDs []uuid.UUID) (int, error)
}

type Repository struct {
	storage Storage
}

func New(storage Storage) *Repository {
	return &Repository{storage: storage}
}

func (r *Repository) CreateUser(ctx context.Context, params CreateUserParams) (User, error) {
	return r.storage.CreateUser(ctx, params)
}

func (r *Repository) GetUserByID(ctx context.Context, userID uuid.UUID) (User, error) {
	return r.storage.GetUserByID(ctx, userID)
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	return r.storage.GetUserByEmail(ctx, email)
}

func (r *Repository) ListUsers(ctx context.Context, params ListUsersParams) ([]User, error) {
	return r.storage.ListUsers(ctx, params)
}

func (r *Repository) CountUsers(ctx context.Context, params ListUsersParams) (int, error) {
	return r.storage.CountUsers(ctx, params)
}

func (r *Repository) UpdateUser(ctx context.Context, params UpdateUserParams) (User, error) {
	return r.storage.UpdateUser(ctx, params)
}

func (r *Repository) SoftDeleteUser(ctx context.Context, userID uuid.UUID) error {
	return r.storage.SoftDeleteUser(ctx, userID)
}

func (r *Repository) SetUserActive(ctx context.Context, userID uuid.UUID, active bool) (User, error) {
	return r.storage.SetUserActive(ctx, userID, active)
}

func (r *Repository) GetProfile(ctx context.Context, userID uuid.UUID) (User, *Client, error) {
	return r.storage.GetProfile(ctx, userID)
}

func (r *Repository) UpdateProfile(ctx context.Context, params UpdateProfileParams) (User, *Client, error) {
	return r.storage.UpdateProfile(ctx, params)
}

func (r *Repository) ListUserClients(ctx context.Context, userID uuid.UUID) ([]Client, error) {
	return r.storage.ListUserClients(ctx, userID)
}

func (r *Repository) UserHasClient(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (bool, error) {
	return r.storage.UserHasClient(ctx, userID, clientID)
}

func (r *Repository) GetClientDetails(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (Client, error) {
	return r.storage.GetClientDetails(ctx, userID, clientID)
}

func (r *Repository) GetClientConditions(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (ClientConditions, error) {
	return r.storage.GetClientConditions(ctx, userID, clientID)
}

func (r *Repository) SetActiveClient(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (Client, error) {
	return r.storage.SetActiveClient(ctx, userID, clientID)
}

func (r *Repository) GetActiveClient(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return r.storage.GetActiveClient(ctx, userID)
}

func (r *Repository) ListFavorites(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) ([]Favorite, error) {
	return r.storage.ListFavorites(ctx, userID, clientID)
}

func (r *Repository) AddFavorite(ctx context.Context, userID uuid.UUID, clientID uuid.UUID, productID uuid.UUID) (Favorite, bool, error) {
	return r.storage.AddFavorite(ctx, userID, clientID, productID)
}

func (r *Repository) DeleteFavorites(ctx context.Context, userID uuid.UUID, clientID uuid.UUID, productIDs []uuid.UUID) (int, error) {
	return r.storage.DeleteFavorites(ctx, userID, clientID, productIDs)
}
