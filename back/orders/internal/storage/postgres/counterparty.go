package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

var ErrUserNotFound = errors.New("user not found")

func (s *Storage) GetActiveClient(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var clientID uuid.NullUUID
	err := s.pool.QueryRow(ctx, `
		SELECT active_client_id
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, userID).Scan(&clientID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrUserNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("get active client: %w", err)
	}
	if !clientID.Valid {
		return uuid.Nil, nil
	}
	return clientID.UUID, nil
}

func (s *Storage) CounterpartyExists(ctx context.Context, clientID uuid.UUID) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM counterparties WHERE id = $1)
	`, clientID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check counterparty existence: %w", err)
	}
	return exists, nil
}

func (s *Storage) UserHasClient(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (bool, error) {
	var allowed bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM user_clients uc
			JOIN users u ON u.id = uc.user_id
			WHERE uc.user_id = $1
			  AND uc.client_id = $2
			  AND u.deleted_at IS NULL
		)
	`, userID, clientID).Scan(&allowed)
	if err != nil {
		return false, fmt.Errorf("check user client: %w", err)
	}
	return allowed, nil
}

func (s *Storage) GetCounterpartyPriceGroupID(ctx context.Context, counterpartyID uuid.NullUUID) (uuid.NullUUID, error) {
	if !counterpartyID.Valid {
		return uuid.NullUUID{}, nil
	}
	var priceGroupID uuid.NullUUID
	err := s.pool.QueryRow(ctx, `SELECT price_group_id FROM counterparties WHERE id = $1`, counterpartyID.UUID).Scan(&priceGroupID)
	if err != nil {
		return uuid.NullUUID{}, fmt.Errorf("get counterparty price group id: %w", err)
	}
	return priceGroupID, nil
}

func (s *Storage) InsertDeliveryAddress(ctx context.Context, counterpartyID uuid.NullUUID, addrType string, address string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `
		INSERT INTO counterparty_addresses (counterparty_id, type, address)
		VALUES ($1, $2, $3)
		RETURNING id
	`, counterpartyID, addrType, address).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert delivery address: %w", err)
	}
	return id, nil
}

func (s *Storage) InsertContact(ctx context.Context, counterpartyID uuid.NullUUID, fullName string, phone string, email string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `
		INSERT INTO counterparty_contacts (counterparty_id, full_name, phone, email)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, counterpartyID, fullName, phone, email).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert contact: %w", err)
	}
	return id, nil
}
