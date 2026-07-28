package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Storage) GetUserCounterpartyID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var counterpartyID uuid.NullUUID
	err := s.pool.QueryRow(ctx, `SELECT counterparty_id FROM users WHERE id = $1`, userID).Scan(&counterpartyID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get user counterparty id: %w", err)
	}
	if !counterpartyID.Valid {
		return uuid.Nil, nil
	}
	return counterpartyID.UUID, nil
}

func (s *Storage) GetCounterpartyPriceGroupID(ctx context.Context, counterpartyID uuid.UUID) (uuid.NullUUID, error) {
	var priceGroupID uuid.NullUUID
	err := s.pool.QueryRow(ctx, `SELECT price_group_id FROM counterparties WHERE id = $1`, counterpartyID).Scan(&priceGroupID)
	if err != nil {
		return uuid.NullUUID{}, fmt.Errorf("get counterparty price group id: %w", err)
	}
	return priceGroupID, nil
}

func (s *Storage) InsertDeliveryAddress(ctx context.Context, counterpartyID uuid.UUID, addrType string, address string) (uuid.UUID, error) {
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

func (s *Storage) InsertContact(ctx context.Context, counterpartyID uuid.UUID, fullName string, phone string, email string) (uuid.UUID, error) {
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
