//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestListPreviouslyOrderedProductsIntegration(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()

	ctx := context.Background()
	storage := New(pool)
	mustScan := func(statement string, args ...interface{}) uuid.UUID {
		t.Helper()
		var id uuid.UUID
		if err := pool.QueryRow(ctx, statement, args...).Scan(&id); err != nil {
			t.Fatalf("fixture insert failed (%s): %v", statement, err)
		}
		return id
	}

	clientID := mustScan(`INSERT INTO counterparties (name) VALUES ('previous orders client') RETURNING id`)
	otherClientID := mustScan(`INSERT INTO counterparties (name) VALUES ('other previous orders client') RETURNING id`)
	userID := mustScan(`INSERT INTO users (email, counterparty_id) VALUES ($1, $2) RETURNING id`, uuid.NewString()+"@test.local", clientID)
	otherUserID := mustScan(`INSERT INTO users (email, counterparty_id) VALUES ($1, $2) RETURNING id`, uuid.NewString()+"@test.local", otherClientID)

	productIDs := make([]uuid.UUID, 3)
	for i := range productIDs {
		productIDs[i] = mustScan(
			`INSERT INTO products (sku, name, slug) VALUES ($1, $2, $3) RETURNING id`,
			uuid.NewString(), "previously ordered product", uuid.NewString(),
		)
	}

	oldOrderID := mustScan(`
		INSERT INTO orders (number, counterparty_id, user_id, status, created_at)
		VALUES ($1, $2, $3, 'new', $4) RETURNING id
	`, uuid.NewString(), clientID, userID, time.Now().Add(-3*time.Hour))
	latestOrderID := mustScan(`
		INSERT INTO orders (number, counterparty_id, user_id, status, created_at)
		VALUES ($1, $2, $3, 'completed', $4) RETURNING id
	`, uuid.NewString(), clientID, userID, time.Now().Add(-time.Hour))
	cancelledOrderID := mustScan(`
		INSERT INTO orders (number, counterparty_id, user_id, status, created_at)
		VALUES ($1, $2, $3, 'cancelled', $4) RETURNING id
	`, uuid.NewString(), clientID, userID, time.Now())
	foreignOrderID := mustScan(`
		INSERT INTO orders (number, counterparty_id, user_id, status, created_at)
		VALUES ($1, $2, $3, 'delivered', $4) RETURNING id
	`, uuid.NewString(), otherClientID, otherUserID, time.Now())
	orderIDs := []uuid.UUID{oldOrderID, latestOrderID, cancelledOrderID, foreignOrderID}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM order_items WHERE order_id = ANY($1::uuid[])`, orderIDs)
		_, _ = pool.Exec(ctx, `DELETE FROM orders WHERE id = ANY($1::uuid[])`, orderIDs)
		_, _ = pool.Exec(ctx, `DELETE FROM products WHERE id = ANY($1::uuid[])`, productIDs)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = ANY($1::uuid[])`, []uuid.UUID{userID, otherUserID})
		_, _ = pool.Exec(ctx, `DELETE FROM counterparties WHERE id = ANY($1::uuid[])`, []uuid.UUID{clientID, otherClientID})
	})

	for _, fixture := range []struct {
		orderID   uuid.UUID
		productID uuid.UUID
	}{
		{oldOrderID, productIDs[0]},
		{oldOrderID, productIDs[1]},
		{latestOrderID, productIDs[0]},
		{cancelledOrderID, productIDs[2]},
		{foreignOrderID, productIDs[2]},
	} {
		if _, err := pool.Exec(ctx, `INSERT INTO order_items (order_id, product_id) VALUES ($1, $2)`, fixture.orderID, fixture.productID); err != nil {
			t.Fatalf("insert order item: %v", err)
		}
	}

	scope := ListPreviouslyOrderedProductsParams{
		UserID:         userID,
		CounterpartyID: uuid.NullUUID{UUID: clientID, Valid: true},
		Limit:          1,
	}
	firstPage, total, err := storage.ListPreviouslyOrderedProducts(ctx, scope)
	if err != nil {
		t.Fatalf("ListPreviouslyOrderedProducts() error = %v", err)
	}
	if total != 2 || len(firstPage) != 1 || firstPage[0].ProductID != productIDs[0] {
		t.Fatalf("first page = %#v, total = %d", firstPage, total)
	}

	scope.Offset = 1
	secondPage, total, err := storage.ListPreviouslyOrderedProducts(ctx, scope)
	if err != nil {
		t.Fatalf("ListPreviouslyOrderedProducts(offset=1) error = %v", err)
	}
	if total != 2 || len(secondPage) != 1 || secondPage[0].ProductID != productIDs[1] {
		t.Fatalf("second page = %#v, total = %d", secondPage, total)
	}

	empty, total, err := storage.ListPreviouslyOrderedProducts(ctx, ListPreviouslyOrderedProductsParams{
		UserID:         userID,
		CounterpartyID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
		Limit:          20,
	})
	if err != nil {
		t.Fatalf("ListPreviouslyOrderedProducts(foreign client) error = %v", err)
	}
	if empty == nil || len(empty) != 0 || total != 0 {
		t.Fatalf("foreign client result = %#v, total = %d", empty, total)
	}
}
