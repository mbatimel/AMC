package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailTaken   = errors.New("email already taken")
	ErrRoleNotFound = errors.New("role not found")
)

//go:embed sql/getUserByEmail.sql
var sqlGetUserByEmail string

//go:embed sql/getUserByID.sql
var sqlGetUserByID string

//go:embed sql/insertUser.sql
var sqlInsertUser string

//go:embed sql/updateUserPassword.sql
var sqlUpdateUserPassword string

//go:embed sql/getRoleByCode.sql
var sqlGetRoleByCode string

//go:embed sql/insertUserRole.sql
var sqlInsertUserRole string

const uniqueViolationCode = "23505"

type User struct {
	ID       uuid.UUID
	Email    string
	Password string
	Status   string
}

type Storage struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := s.pool.QueryRow(ctx, sqlGetUserByEmail, email).Scan(&user.ID, &user.Email, &user.Password, &user.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (s *Storage) GetUserByID(ctx context.Context, userID uuid.UUID) (User, error) {
	var user User
	err := s.pool.QueryRow(ctx, sqlGetUserByID, userID).Scan(&user.ID, &user.Email, &user.Password, &user.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

// CreateUserWithRole inserts a new user and assigns the given role to them in a single transaction.
func (s *Storage) CreateUserWithRole(ctx context.Context, email, passwordHash, name, surename string, roleCode int) (uuid.UUID, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("begin create user transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID uuid.UUID
	err = tx.QueryRow(ctx, sqlInsertUser, email, passwordHash, name, surename).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return uuid.UUID{}, ErrEmailTaken
		}
		return uuid.UUID{}, fmt.Errorf("insert user: %w", err)
	}

	var roleID uuid.UUID
	err = tx.QueryRow(ctx, sqlGetRoleByCode, roleCode).Scan(&roleID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.UUID{}, ErrRoleNotFound
	}
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("get role by code: %w", err)
	}

	if _, err = tx.Exec(ctx, sqlInsertUserRole, userID, roleID); err != nil {
		return uuid.UUID{}, fmt.Errorf("insert user role: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return uuid.UUID{}, fmt.Errorf("commit create user transaction: %w", err)
	}

	return userID, nil
}

func (s *Storage) UpdateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	tag, err := s.pool.Exec(ctx, sqlUpdateUserPassword, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
