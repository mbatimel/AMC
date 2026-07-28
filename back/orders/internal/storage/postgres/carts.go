package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

var (
	ErrProductPriceNotFound = errors.New("product price not found")
	ErrCartItemNotFound     = errors.New("cart item not found")
)

type CartItemRow struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	SKU         string
	ProductName string
	Qty         int
	Price       float64
}

func (s *Storage) GetOrCreateCart(ctx context.Context, userID uuid.UUID, counterpartyID uuid.UUID) (uuid.UUID, error) {
	var cartID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM carts WHERE user_id = $1 AND counterparty_id = $2
	`, userID, counterpartyID).Scan(&cartID)
	if err == nil {
		return cartID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("get cart: %w", err)
	}

	err = s.pool.QueryRow(ctx, `
		INSERT INTO carts (user_id, counterparty_id) VALUES ($1, $2) RETURNING id
	`, userID, counterpartyID).Scan(&cartID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create cart: %w", err)
	}
	return cartID, nil
}

func (s *Storage) GetCartItems(ctx context.Context, cartID uuid.UUID) ([]CartItemRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ci.id, ci.product_id, p.sku, p.name, ci.quantity, ci.price
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY ci.id
	`, cartID)
	if err != nil {
		return nil, fmt.Errorf("get cart items: %w", err)
	}
	defer rows.Close()

	items := make([]CartItemRow, 0)
	for rows.Next() {
		var item CartItemRow
		var qty float64
		if err = rows.Scan(&item.ID, &item.ProductID, &item.SKU, &item.ProductName, &qty, &item.Price); err != nil {
			return nil, fmt.Errorf("scan cart item: %w", err)
		}
		item.Qty = int(qty)
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cart items: %w", err)
	}
	return items, nil
}

func (s *Storage) ResolveProductPrice(ctx context.Context, productID uuid.UUID, priceGroupID uuid.NullUUID) (float64, error) {
	var price float64
	err := s.pool.QueryRow(ctx, `
		SELECT price FROM product_prices
		WHERE product_id = $1
		  AND (price_group_id = $2 OR ($2::uuid IS NULL AND price_group_id IS NULL))
		ORDER BY valid_from DESC NULLS LAST
		LIMIT 1
	`, productID, priceGroupID).Scan(&price)
	if err == nil {
		return price, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("resolve product price by group: %w", err)
	}

	err = s.pool.QueryRow(ctx, `
		SELECT price FROM product_prices
		WHERE product_id = $1
		ORDER BY valid_from DESC NULLS LAST
		LIMIT 1
	`, productID).Scan(&price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrProductPriceNotFound
		}
		return 0, fmt.Errorf("resolve product price fallback: %w", err)
	}
	return price, nil
}

func (s *Storage) UpsertCartItem(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, qty int, price float64) error {
	cmdTag, err := s.pool.Exec(ctx, `
		UPDATE cart_items SET quantity = quantity + $3
		WHERE cart_id = $1 AND product_id = $2
	`, cartID, productID, qty)
	if err != nil {
		return fmt.Errorf("update cart item quantity: %w", err)
	}
	if cmdTag.RowsAffected() > 0 {
		return nil
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO cart_items (cart_id, product_id, quantity, price) VALUES ($1, $2, $3, $4)
	`, cartID, productID, qty, price)
	if err != nil {
		return fmt.Errorf("insert cart item: %w", err)
	}
	return nil
}

func (s *Storage) SetCartItemQuantity(ctx context.Context, cartItemID uuid.UUID, cartID uuid.UUID, qty int) error {
	cmdTag, err := s.pool.Exec(ctx, `
		UPDATE cart_items SET quantity = $3 WHERE id = $1 AND cart_id = $2
	`, cartItemID, cartID, qty)
	if err != nil {
		return fmt.Errorf("update cart item quantity: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrCartItemNotFound
	}
	return nil
}

func (s *Storage) DeleteCartItem(ctx context.Context, cartItemID uuid.UUID, cartID uuid.UUID) error {
	cmdTag, err := s.pool.Exec(ctx, `DELETE FROM cart_items WHERE id = $1 AND cart_id = $2`, cartItemID, cartID)
	if err != nil {
		return fmt.Errorf("delete cart item: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrCartItemNotFound
	}
	return nil
}

func (s *Storage) ClearCartItems(ctx context.Context, cartID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID)
	if err != nil {
		return fmt.Errorf("clear cart items: %w", err)
	}
	return nil
}
