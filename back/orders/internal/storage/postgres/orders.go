package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
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
	CounterpartyID    uuid.NullUUID
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
	// UserID scopes orders when CounterpartyID is not valid: a user with no
	// client must only see their own orders, never every clientless user's.
	UserID         uuid.UUID
	CounterpartyID uuid.NullUUID
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
		WHERE (
		    (counterparty_id = $1 AND $1::uuid IS NOT NULL)
		    OR (counterparty_id IS NULL AND $1::uuid IS NULL AND user_id = $6)
		  )
		  AND ($2 = '' OR status = $2)
		  AND ($3 = '' OR payment_status = $3)
		ORDER BY %s
		LIMIT $4 OFFSET $5
	`, orderBy), params.CounterpartyID, params.Status, params.PaymentStatus, limit, params.Offset, params.UserID)
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
		WHERE (
		    (counterparty_id = $1 AND $1::uuid IS NOT NULL)
		    OR (counterparty_id IS NULL AND $1::uuid IS NULL AND user_id = $4)
		  )
		  AND ($2 = '' OR status = $2)
		  AND ($3 = '' OR payment_status = $3)
	`, params.CounterpartyID, params.Status, params.PaymentStatus, params.UserID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	return items, total, nil
}

type PreviouslyOrderedProductRow struct {
	ProductID     uuid.UUID
	LastOrderedAt time.Time
}

type ListPreviouslyOrderedProductsParams struct {
	UserID         uuid.UUID
	CounterpartyID uuid.NullUUID
	Limit          int
	Offset         int
}

func (s *Storage) ListPreviouslyOrderedProducts(ctx context.Context, params ListPreviouslyOrderedProductsParams) ([]PreviouslyOrderedProductRow, int, error) {
	const historyCTE = `
		WITH product_history AS (
			SELECT oi.product_id, MAX(o.created_at) AS last_ordered_at
			FROM orders o
			JOIN order_items oi ON oi.order_id = o.id
			WHERE (
				(o.counterparty_id = $1 AND $1::uuid IS NOT NULL)
				OR (o.counterparty_id IS NULL AND $1::uuid IS NULL AND o.user_id = $2)
			)
			AND o.status <> 'cancelled'
			AND oi.product_id IS NOT NULL
			GROUP BY oi.product_id
		)
	`

	rows, err := s.pool.Query(ctx, historyCTE+`
		SELECT product_id, last_ordered_at
		FROM product_history
		ORDER BY last_ordered_at DESC, product_id
		LIMIT $3 OFFSET $4
	`, params.CounterpartyID, params.UserID, params.Limit, params.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list previously ordered products: %w", err)
	}
	defer rows.Close()

	items := make([]PreviouslyOrderedProductRow, 0)
	for rows.Next() {
		var item PreviouslyOrderedProductRow
		if err = rows.Scan(&item.ProductID, &item.LastOrderedAt); err != nil {
			return nil, 0, fmt.Errorf("scan previously ordered product: %w", err)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate previously ordered products: %w", err)
	}

	var total int
	if err = s.pool.QueryRow(ctx, historyCTE+`SELECT COUNT(*) FROM product_history`, params.CounterpartyID, params.UserID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count previously ordered products: %w", err)
	}
	return items, total, nil
}

type OrderDetailRow struct {
	ID              uuid.UUID
	Number          string
	UserID          uuid.UUID
	CounterpartyID  uuid.NullUUID
	Status          string
	PaymentStatus   string
	DeliveryMethod  string
	DeliveryAddress string
	ContactName     string
	Phone           string
	Email           string
	Comment         string
	Subtotal        float64
	DiscountTotal   float64
	VATTotal        float64
	Total           float64
	CreatedAt       time.Time
}

var ErrOrderNotFound = errors.New("order not found")

const orderDetailSelect = `
	SELECT o.id, o.number, o.user_id, o.counterparty_id, o.status, o.payment_status, o.delivery_method,
	       COALESCE(ca.address, ''), COALESCE(cc.full_name, ''), COALESCE(cc.phone, ''), COALESCE(cc.email, ''),
	       COALESCE(o.comment, ''), o.subtotal, o.discount_total, o.vat_total, o.total, o.created_at
	FROM orders o
	LEFT JOIN counterparty_addresses ca ON ca.id = o.delivery_address_id
	LEFT JOIN counterparty_contacts cc ON cc.id = o.contact_id
`

func scanOrderDetail(row pgx.Row) (OrderDetailRow, error) {
	var order OrderDetailRow
	err := row.Scan(
		&order.ID,
		&order.Number,
		&order.UserID,
		&order.CounterpartyID,
		&order.Status,
		&order.PaymentStatus,
		&order.DeliveryMethod,
		&order.DeliveryAddress,
		&order.ContactName,
		&order.Phone,
		&order.Email,
		&order.Comment,
		&order.Subtotal,
		&order.DiscountTotal,
		&order.VATTotal,
		&order.Total,
		&order.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return order, ErrOrderNotFound
		}
		return order, fmt.Errorf("get order: %w", err)
	}
	return order, nil
}

// GetOrderByID resolves an order for a buyer-facing request. When counterpartyID
// is not valid (the caller has no client bound), the lookup falls back to
// userID scoping so one clientless user can't read another's orders through a
// shared NULL counterparty bucket.
func (s *Storage) GetOrderByID(ctx context.Context, orderID uuid.UUID, userID uuid.UUID, counterpartyID uuid.NullUUID) (OrderDetailRow, error) {
	row := s.pool.QueryRow(ctx, orderDetailSelect+`
		WHERE o.id = $1 AND (
		    (o.counterparty_id = $2 AND $2::uuid IS NOT NULL)
		    OR (o.counterparty_id IS NULL AND $2::uuid IS NULL AND o.user_id = $3)
		  )
	`, orderID, counterpartyID, userID)
	return scanOrderDetail(row)
}

// orderDetailByID re-reads an order by ID with no ownership scoping. Only for
// internal use after the caller has already established access some other way
// (admin role check, or a prior scoped SELECT ... FOR UPDATE in the same
// transaction/request).
func (s *Storage) orderDetailByID(ctx context.Context, orderID uuid.UUID) (OrderDetailRow, error) {
	row := s.pool.QueryRow(ctx, orderDetailSelect+`WHERE o.id = $1`, orderID)
	return scanOrderDetail(row)
}

func (s *Storage) CancelOrder(ctx context.Context, orderID uuid.UUID, counterpartyID uuid.NullUUID, changedBy uuid.UUID, comment string) (OrderDetailRow, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return OrderDetailRow{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var oldStatus, oldPaymentStatus string
	err = tx.QueryRow(ctx, `
		SELECT status, payment_status FROM orders
		WHERE id = $1 AND (
		    (counterparty_id = $2 AND $2::uuid IS NOT NULL)
		    OR (counterparty_id IS NULL AND $2::uuid IS NULL AND user_id = $3)
		  ) FOR UPDATE
	`, orderID, counterpartyID, changedBy).Scan(&oldStatus, &oldPaymentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OrderDetailRow{}, ErrOrderNotFound
		}
		return OrderDetailRow{}, fmt.Errorf("get order status: %w", err)
	}

	if _, err = tx.Exec(ctx, `UPDATE orders SET status = 'cancelled' WHERE id = $1`, orderID); err != nil {
		return OrderDetailRow{}, fmt.Errorf("cancel order: %w", err)
	}

	if _, err = tx.Exec(ctx, `
		INSERT INTO order_status_history (order_id, old_status, new_status, payment_status, changed_by, comment)
		VALUES ($1, $2, 'cancelled', $3, $4, $5)
	`, orderID, oldStatus, oldPaymentStatus, changedBy, comment); err != nil {
		return OrderDetailRow{}, fmt.Errorf("insert order status history: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return OrderDetailRow{}, fmt.Errorf("commit tx: %w", err)
	}

	return s.orderDetailByID(ctx, orderID)
}

func (s *Storage) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string, paymentStatus string, comment string, changedBy uuid.UUID) (OrderDetailRow, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return OrderDetailRow{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var oldStatus, oldPaymentStatus string
	err = tx.QueryRow(ctx, `
		SELECT status, payment_status FROM orders WHERE id = $1 FOR UPDATE
	`, orderID).Scan(&oldStatus, &oldPaymentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OrderDetailRow{}, ErrOrderNotFound
		}
		return OrderDetailRow{}, fmt.Errorf("get order status: %w", err)
	}

	newStatus := status
	if newStatus == "" {
		newStatus = oldStatus
	}
	newPaymentStatus := paymentStatus
	if newPaymentStatus == "" {
		newPaymentStatus = oldPaymentStatus
	}

	if _, err = tx.Exec(ctx, `UPDATE orders SET status = $1, payment_status = $2 WHERE id = $3`, newStatus, newPaymentStatus, orderID); err != nil {
		return OrderDetailRow{}, fmt.Errorf("update order status: %w", err)
	}

	if _, err = tx.Exec(ctx, `
		INSERT INTO order_status_history (order_id, old_status, new_status, payment_status, changed_by, comment)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, orderID, oldStatus, newStatus, newPaymentStatus, changedBy, comment); err != nil {
		return OrderDetailRow{}, fmt.Errorf("insert order status history: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return OrderDetailRow{}, fmt.Errorf("commit tx: %w", err)
	}

	return s.orderDetailByID(ctx, orderID)
}

type OrderHistoryRow struct {
	ID            uuid.UUID
	OrderID       uuid.UUID
	Status        string
	PaymentStatus string
	Comment       string
	ChangedBy     uuid.NullUUID
	CreatedAt     time.Time
}

func (s *Storage) GetOrderHistory(ctx context.Context, orderID uuid.UUID) ([]OrderHistoryRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, order_id, new_status, COALESCE(payment_status, ''), COALESCE(comment, ''), changed_by, created_at
		FROM order_status_history WHERE order_id = $1 ORDER BY created_at
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order history: %w", err)
	}
	defer rows.Close()

	items := make([]OrderHistoryRow, 0)
	for rows.Next() {
		var item OrderHistoryRow
		if err = rows.Scan(&item.ID, &item.OrderID, &item.Status, &item.PaymentStatus, &item.Comment, &item.ChangedBy, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan order history: %w", err)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order history: %w", err)
	}
	return items, nil
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

type ProductOnecRef struct {
	OneCGUID uuid.NullUUID
	SKU      string
}

func (s *Storage) GetProductOnecRefs(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]ProductOnecRef, error) {
	result := make(map[uuid.UUID]ProductOnecRef, len(productIDs))
	if len(productIDs) == 0 {
		return result, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, one_c_guid, sku FROM products WHERE id = ANY($1)
	`, productIDs)
	if err != nil {
		return nil, fmt.Errorf("get product onec refs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var ref ProductOnecRef
		if err = rows.Scan(&id, &ref.OneCGUID, &ref.SKU); err != nil {
			return nil, fmt.Errorf("get product onec refs: scan: %w", err)
		}
		result[id] = ref
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("get product onec refs: %w", err)
	}
	return result, nil
}

type CounterpartyOnecRef struct {
	OneCGUID uuid.NullUUID
	INN      string
	Name     string
}

func (s *Storage) GetCounterpartyOnecRef(ctx context.Context, counterpartyID uuid.UUID) (CounterpartyOnecRef, error) {
	var ref CounterpartyOnecRef
	err := s.pool.QueryRow(ctx, `
		SELECT one_c_guid, COALESCE(inn, ''), COALESCE(name, '') FROM counterparties WHERE id = $1
	`, counterpartyID).Scan(&ref.OneCGUID, &ref.INN, &ref.Name)
	if err != nil {
		return CounterpartyOnecRef{}, fmt.Errorf("get counterparty onec ref: %w", err)
	}
	return ref, nil
}
