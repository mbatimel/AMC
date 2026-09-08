//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestListOrdersHandlesNullableOrderColumns(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()

	ctx := context.Background()
	storage := New(pool)

	var counterpartyID uuid.UUID
	if err := pool.QueryRow(ctx, `
		INSERT INTO counterparties (name) VALUES ('nullable order diagnosis') RETURNING id
	`).Scan(&counterpartyID); err != nil {
		t.Fatal(err)
	}

	var userID uuid.UUID
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (email, counterparty_id) VALUES ($1, $2) RETURNING id
	`, uuid.NewString()+"@test.local", counterpartyID).Scan(&userID); err != nil {
		t.Fatal(err)
	}

	var orderID uuid.UUID
	if err := pool.QueryRow(ctx, `
		INSERT INTO orders (number, counterparty_id, user_id, status)
		VALUES ($1, $2, $3, 'new') RETURNING id
	`, uuid.NewString(), counterpartyID, userID).Scan(&orderID); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM orders WHERE id = $1`, orderID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
		_, _ = pool.Exec(ctx, `DELETE FROM counterparties WHERE id = $1`, counterpartyID)
	})

	rows, total, err := storage.ListOrders(ctx, ListOrdersParams{
		UserID:         userID,
		CounterpartyID: uuid.NullUUID{UUID: counterpartyID, Valid: true},
		Limit:          50,
		Offset:         0,
	})
	if err != nil {
		t.Fatalf("ListOrders: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Fatalf("ListOrders returned total=%d, rows=%d; want 1, 1", total, len(rows))
	}
	row := rows[0]
	if row.ID != orderID || row.DeliveryMethod != "" {
		t.Fatalf("unexpected order row: %#v", row)
	}
	if row.Subtotal != 0 || row.DiscountTotal != 0 || row.VATTotal != 0 || row.Total != 0 {
		t.Fatalf("nullable monetary values were not normalized: %#v", row)
	}
}
