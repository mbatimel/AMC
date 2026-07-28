package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

func (s *Storage) GetVolumeDiscountPercent(ctx context.Context, counterpartyID uuid.UUID, priceGroupID uuid.NullUUID, subtotal float64) (float64, error) {
	var discountPercent float64
	err := s.pool.QueryRow(ctx, `
		SELECT discount_percent FROM volume_discounts
		WHERE (counterparty_id = $1 OR (counterparty_id IS NULL AND price_group_id = $2 AND $2::uuid IS NOT NULL))
		  AND min_order_amount <= $3
		ORDER BY (counterparty_id = $1) DESC, min_order_amount DESC
		LIMIT 1
	`, counterpartyID, priceGroupID, subtotal).Scan(&discountPercent)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("get volume discount: %w", err)
	}
	return discountPercent, nil
}
