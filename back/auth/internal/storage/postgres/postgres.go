package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/auth/internal/models"
)

type Storage interface {
	CreateUser(ctx context.Context, user *models.User) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error
}

type storage struct {
	conn   ConnectManager
	logger zerolog.Logger
}

func NewStorage(conn ConnectManager, logger zerolog.Logger) Storage {
	return &storage{
		conn:   conn,
		logger: logger,
	}
}

func (s *storage) rollback(ctx context.Context, tx pgx.Tx) {
	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		s.logger.Error().Err(err).Msg("rollback transaction")
	}
}

func (s *storage) CreateUser(ctx context.Context, user *models.User) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := s.conn.GetConnection(MASTER).BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadWrite})
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin create user transaction: %w", err)
	}
	defer s.rollback(ctx, tx)

	var id uuid.UUID
	if err = tx.QueryRow(ctx, sqlCreateUser, user.Email, user.Password, user.Name, user.Surename, user.Status).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("create user: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit create user transaction: %w", err)
	}

	return id, nil
}

func (s *storage) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := s.conn.GetConnection(REPLICA).BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, fmt.Errorf("begin get user by email transaction: %w", err)
	}
	defer s.rollback(ctx, tx)

	var user models.User
	err = tx.QueryRow(ctx, sqlGetUserByEmail, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.Surename, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit get user by email transaction: %w", err)
	}

	return &user, nil
}

func (s *storage) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := s.conn.GetConnection(REPLICA).BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, fmt.Errorf("begin get user by id transaction: %w", err)
	}
	defer s.rollback(ctx, tx)

	var user models.User
	err = tx.QueryRow(ctx, sqlGetUserByID, userID).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.Surename, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit get user by id transaction: %w", err)
	}

	return &user, nil
}

func (s *storage) UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := s.conn.GetConnection(MASTER).BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadWrite})
	if err != nil {
		return fmt.Errorf("begin update user password transaction: %w", err)
	}
	defer s.rollback(ctx, tx)

	if _, err = tx.Exec(ctx, sqlUpdateUserPassword, userID, hashedPassword); err != nil {
		return fmt.Errorf("update user password: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit update user password transaction: %w", err)
	}

	return nil
}

func (s *storage) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := s.conn.GetConnection(MASTER).BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadWrite})
	if err != nil {
		return fmt.Errorf("begin update user status transaction: %w", err)
	}
	defer s.rollback(ctx, tx)

	if _, err = tx.Exec(ctx, sqlUpdateUserStatus, userID, status); err != nil {
		return fmt.Errorf("update user status: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit update user status transaction: %w", err)
	}

	return nil
}
