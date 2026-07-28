package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type OrderItemInput struct {
	ProductID       uuid.UUID
	SKU             string
	Name            string
	Quantity        int
	UnitPrice       float64
	DiscountPercent float64
	VATRate         float64
	LineTotal       float64
}

type CreateOrderParams struct {
	UserID            uuid.UUID
	CounterpartyID    uuid.UUID
	CartID            uuid.UUID
	DeliveryMethod    string
	DeliveryAddressID uuid.UUID
	ContactID         uuid.UUID
	Comment           string
	Subtotal          float64
	DiscountTotal     float64
	VATTotal          float64
	Total             float64
	Items             []OrderItemInput
}

type CreatedOrder struct {
	ID        uuid.UUID
	Number    string
	CreatedAt time.Time
}

func (s *Storage) CreateOrder(ctx context.Context, params CreateOrderParams) (CreatedOrder, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CreatedOrder{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var order CreatedOrder
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (
			number, counterparty_id, user_id, status, payment_status,
			delivery_method, delivery_address_id, contact_id, comment,
			subtotal, discount_total, vat_total, total
		) VALUES (
			'AMC-' || to_char(now(), 'YYYYMMDD') || '-' || lpad(nextval('orders_number_seq')::text, 5, '0'),
			$1, $2, 'new', 'not_paid', $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id, number, created_at
	`,
		params.CounterpartyID, params.UserID, params.DeliveryMethod, params.DeliveryAddressID, params.ContactID, params.Comment,
		params.Subtotal, params.DiscountTotal, params.VATTotal, params.Total,
	).Scan(&order.ID, &order.Number, &order.CreatedAt)
	if err != nil {
		return CreatedOrder{}, fmt.Errorf("insert order: %w", err)
	}

	for _, item := range params.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, sku, name, quantity, unit_price, discount_percent, vat_rate, line_total)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, order.ID, item.ProductID, item.SKU, item.Name, item.Quantity, item.UnitPrice, item.DiscountPercent, item.VATRate, item.LineTotal)
		if err != nil {
			return CreatedOrder{}, fmt.Errorf("insert order item: %w", err)
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO order_status_history (order_id, old_status, new_status, payment_status, changed_by, comment)
		VALUES ($1, NULL, 'new', 'not_paid', $2, $3)
	`, order.ID, params.UserID, params.Comment)
	if err != nil {
		return CreatedOrder{}, fmt.Errorf("insert order status history: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, params.CartID)
	if err != nil {
		return CreatedOrder{}, fmt.Errorf("clear cart after order: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return CreatedOrder{}, fmt.Errorf("commit tx: %w", err)
	}

	return order, nil
}

type OrderRow struct {
	ID             uuid.UUID
	Number         string
	Status         string
	PaymentStatus  string
	DeliveryMethod string
	Subtotal       float64
	DiscountTotal  float64
	VATTotal       float64
	Total          float64
	CreatedAt      time.Time
}

type ListOrdersParams struct {
	CounterpartyID uuid.UUID
	Status         string
	PaymentStatus  string
	Limit          int
	Offset         int
	Sort           string
}

func (s *Storage) ListOrders(ctx context.Context, params ListOrdersParams) ([]OrderRow, int, error) {
	orderBy := "created_at DESC"
	switch params.Sort {
	case "created_at":
		orderBy = "created_at ASC"
	case "created_at desc":
		orderBy = "created_at DESC"
	case "total":
		orderBy = "total ASC"
	case "total desc":
		orderBy = "total DESC"
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, number, status, payment_status, delivery_method, subtotal, discount_total, vat_total, total, created_at
		FROM orders
		WHERE counterparty_id = $1
		  AND ($2 = '' OR status = $2)
		  AND ($3 = '' OR payment_status = $3)
		ORDER BY %s
		LIMIT $4 OFFSET $5
	`, orderBy), params.CounterpartyID, params.Status, params.PaymentStatus, limit, params.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	items := make([]OrderRow, 0)
	for rows.Next() {
		var order OrderRow
		if err = rows.Scan(
			&order.ID,
			&order.Number,
			&order.Status,
			&order.PaymentStatus,
			&order.DeliveryMethod,
			&order.Subtotal,
			&order.DiscountTotal,
			&order.VATTotal,
			&order.Total,
			&order.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		items = append(items, order)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate orders: %w", err)
	}

	var total int
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM orders
		WHERE counterparty_id = $1
		  AND ($2 = '' OR status = $2)
		  AND ($3 = '' OR payment_status = $3)
	`, params.CounterpartyID, params.Status, params.PaymentStatus).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	return items, total, nil
}

type OrderItemRow struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	SKU       string
	Name      string
	Quantity  int
	UnitPrice float64
	LineTotal float64
}

func (s *Storage) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]OrderItemRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, product_id, sku, name, quantity, unit_price, line_total
		FROM order_items WHERE order_id = $1 ORDER BY id
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}
	defer rows.Close()

	items := make([]OrderItemRow, 0)
	for rows.Next() {
		var item OrderItemRow
		var qty float64
		if err = rows.Scan(&item.ID, &item.ProductID, &item.SKU, &item.Name, &qty, &item.UnitPrice, &item.LineTotal); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		item.Quantity = int(qty)
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order items: %w", err)
	}
	return items, nil
}

type OrderDocumentRow struct {
	ID        uuid.UUID
	Type      string
	Number    string
	URL       string
	CreatedAt time.Time
}

func (s *Storage) GetOrderDocumentsByOrderID(ctx context.Context, orderID uuid.UUID) ([]OrderDocumentRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT d.id, d.type, d.number, COALESCE(f.storage_key, ''), d.created_at
		FROM documents d
		LEFT JOIN files f ON f.id = d.file_id
		WHERE d.order_id = $1
		ORDER BY d.created_at
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order documents: %w", err)
	}
	defer rows.Close()

	items := make([]OrderDocumentRow, 0)
	for rows.Next() {
		var doc OrderDocumentRow
		if err = rows.Scan(&doc.ID, &doc.Type, &doc.Number, &doc.URL, &doc.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan order document: %w", err)
		}
		items = append(items, doc)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order documents: %w", err)
	}
	return items, nil
}
