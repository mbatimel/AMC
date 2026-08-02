package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

const foreignKeyViolationCode = "23503"

var (
	ErrProductPriceNotFound = errors.New("product price not found")
	ErrCartItemNotFound     = errors.New("cart item not found")
	ErrCartNotFound         = errors.New("cart not found")
	ErrCounterpartyNotFound = errors.New("counterparty not found")
)

type CartItemRow struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	SKU         string
	ProductName string
	Qty         int
	Price       float64
}

func (s *Storage) GetCart(ctx context.Context, userID uuid.UUID, counterpartyID uuid.NullUUID) (uuid.UUID, error) {
	var cartID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM carts
		WHERE user_id = $1 AND (counterparty_id = $2 OR (counterparty_id IS NULL AND $2::uuid IS NULL))
	`, userID, counterpartyID).Scan(&cartID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrCartNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("get cart: %w", err)
	}
	return cartID, nil
}

func (s *Storage) GetOrCreateCart(ctx context.Context, userID uuid.UUID, counterpartyID uuid.NullUUID) (uuid.UUID, error) {
	cartID, err := s.GetCart(ctx, userID, counterpartyID)
	if err == nil {
		return cartID, nil
	}
	if !errors.Is(err, ErrCartNotFound) {
		return uuid.Nil, err
	}

	err = s.pool.QueryRow(ctx, `
		INSERT INTO carts (user_id, counterparty_id) VALUES ($1, $2) RETURNING id
	`, userID, counterpartyID).Scan(&cartID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolationCode && strings.Contains(pgErr.ConstraintName, "counterparty") {
			return uuid.Nil, ErrCounterpartyNotFound
		}
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

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// promo CTE finds the highest discount among promotions currently running
// for the product where qty meets that promotion's own per-product min_qty
// threshold (see back/products promotions feature).

func (s *Storage) resolveBasePriceAndDiscount(ctx context.Context, productID uuid.UUID, priceGroupID uuid.NullUUID, qty int) (float64, error) {
	var price float64
	var effectiveDiscountPercent float64
	err := s.pool.QueryRow(ctx, `
		WITH base AS (
			SELECT price FROM product_prices
			WHERE product_id = $1 AND price_type = 'base'
			  AND (price_group_id = $2 OR ($2::uuid IS NULL AND price_group_id IS NULL))
			ORDER BY valid_from DESC NULLS LAST, id DESC
			LIMIT 1
		),
		manual AS (
			SELECT discount_percent FROM product_prices
			WHERE product_id = $1 AND price_type = 'client'
			  AND (price_group_id = $2 OR ($2::uuid IS NULL AND price_group_id IS NULL))
			ORDER BY valid_from DESC NULLS LAST, id DESC
			LIMIT 1
		),
		promo AS (
			SELECT MAX(pr.discount_percent) AS discount_percent
			FROM promotion_products pp
			JOIN promotions pr ON pr.id = pp.promotion_id
			WHERE pp.product_id = $1
			  AND pp.min_qty <= $3
			  AND now() BETWEEN pr.starts_at AND pr.ends_at
		)
		SELECT base.price, GREATEST(COALESCE(manual.discount_percent, 0), COALESCE(promo.discount_percent, 0))
		FROM base
		LEFT JOIN manual ON TRUE
		LEFT JOIN promo ON TRUE
	`, productID, priceGroupID, qty).Scan(&price, &effectiveDiscountPercent)
	if err != nil {
		return 0, err
	}
	return round2(price * (1 - effectiveDiscountPercent/100)), nil
}

func (s *Storage) resolveBasePriceAndDiscountAnyGroup(ctx context.Context, productID uuid.UUID, qty int) (float64, error) {
	var price float64
	var effectiveDiscountPercent float64
	err := s.pool.QueryRow(ctx, `
		WITH base AS (
			SELECT price FROM product_prices
			WHERE product_id = $1 AND price_type = 'base'
			ORDER BY valid_from DESC NULLS LAST, id DESC
			LIMIT 1
		),
		manual AS (
			SELECT discount_percent FROM product_prices
			WHERE product_id = $1 AND price_type = 'client'
			ORDER BY valid_from DESC NULLS LAST, id DESC
			LIMIT 1
		),
		promo AS (
			SELECT MAX(pr.discount_percent) AS discount_percent
			FROM promotion_products pp
			JOIN promotions pr ON pr.id = pp.promotion_id
			WHERE pp.product_id = $1
			  AND pp.min_qty <= $2
			  AND now() BETWEEN pr.starts_at AND pr.ends_at
		)
		SELECT base.price, GREATEST(COALESCE(manual.discount_percent, 0), COALESCE(promo.discount_percent, 0))
		FROM base
		LEFT JOIN manual ON TRUE
		LEFT JOIN promo ON TRUE
	`, productID, qty).Scan(&price, &effectiveDiscountPercent)
	if err != nil {
		return 0, err
	}
	return round2(price * (1 - effectiveDiscountPercent/100)), nil
}

// ResolveProductPrice returns the final unit price for qty units of the
// product, combining the base price, the product's own manual discount and
// the best currently-active promotion whose min_qty threshold qty satisfies
// (see calcEffectiveDiscount in orders/internal/service for the equivalent
// pure-Go rule this SQL mirrors).
func (s *Storage) ResolveProductPrice(ctx context.Context, productID uuid.UUID, priceGroupID uuid.NullUUID, qty int) (float64, error) {
	price, err := s.resolveBasePriceAndDiscount(ctx, productID, priceGroupID, qty)
	if err == nil {
		return price, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("resolve product price by group: %w", err)
	}

	price, err = s.resolveBasePriceAndDiscountAnyGroup(ctx, productID, qty)
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
