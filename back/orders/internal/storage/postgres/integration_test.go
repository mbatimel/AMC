//go:build integration

package postgres

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost port=5432 dbname=AMC sslmode=disable user=mbatimel password=mbatimel"
	}

	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		t.Skipf("postgres not reachable, skipping integration test: %v", err)
	}
	return pool
}

func TestCartRoundTrip(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()

	ctx := context.Background()
	storage := New(pool)

	mustScan := func(sql string, args ...interface{}) uuid.UUID {
		var id uuid.UUID
		if err := pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
			t.Fatalf("fixture insert failed (%s): %v", sql, err)
		}
		return id
	}

	priceGroupID := mustScan(`INSERT INTO price_groups (code, name) VALUES ($1, 'test group') RETURNING id`, uuid.NewString())
	counterpartyID := mustScan(`INSERT INTO counterparties (name, price_group_id) VALUES ('test counterparty', $1) RETURNING id`, priceGroupID)
	userID := mustScan(`INSERT INTO users (email, counterparty_id) VALUES ($1, $2) RETURNING id`, uuid.NewString()+"@test.local", counterpartyID)
	categoryID := mustScan(`INSERT INTO categories (name, slug) VALUES ('test category', $1) RETURNING id`, uuid.NewString())
	unitID := mustScan(`INSERT INTO units (code, name) VALUES ($1, 'test unit') RETURNING id`, uuid.NewString())
	productID := mustScan(`INSERT INTO products (category_id, unit_id, sku, name, slug) VALUES ($1, $2, $3, 'test product', $4) RETURNING id`, categoryID, unitID, uuid.NewString(), uuid.NewString())
	if _, err := pool.Exec(ctx, `INSERT INTO product_prices (product_id, price_group_id, price_type, price) VALUES ($1, $2, 'base', 500)`, productID, priceGroupID); err != nil {
		t.Fatalf("insert product_price: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM cart_items WHERE product_id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM carts WHERE counterparty_id = $1`, counterpartyID)
		pool.Exec(ctx, `DELETE FROM product_prices WHERE product_id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM units WHERE id = $1`, unitID)
		pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, categoryID)
		pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
		pool.Exec(ctx, `DELETE FROM counterparties WHERE id = $1`, counterpartyID)
		pool.Exec(ctx, `DELETE FROM price_groups WHERE id = $1`, priceGroupID)
	})

	counterpartyNullID := uuid.NullUUID{UUID: counterpartyID, Valid: true}

	resolvedGroupID, err := storage.GetCounterpartyPriceGroupID(ctx, counterpartyNullID)
	if err != nil {
		t.Fatalf("GetCounterpartyPriceGroupID: %v", err)
	}
	if !resolvedGroupID.Valid || resolvedGroupID.UUID != priceGroupID {
		t.Fatalf("expected price group %s, got %+v", priceGroupID, resolvedGroupID)
	}

	price, err := storage.ResolveProductPrice(ctx, productID, resolvedGroupID, 2)
	if err != nil {
		t.Fatalf("ResolveProductPrice: %v", err)
	}
	if price != 500 {
		t.Fatalf("expected price 500, got %v", price)
	}

	if _, err = storage.GetCart(ctx, userID, counterpartyNullID); !errors.Is(err, ErrCartNotFound) {
		t.Fatalf("GetCart before creation error = %v, want ErrCartNotFound", err)
	}

	cartID, err := storage.GetOrCreateCart(ctx, userID, counterpartyNullID)
	if err != nil {
		t.Fatalf("GetOrCreateCart: %v", err)
	}
	resolvedCartID, err := storage.GetCart(ctx, userID, counterpartyNullID)
	if err != nil || resolvedCartID != cartID {
		t.Fatalf("GetCart after creation = %s, %v; want %s", resolvedCartID, err, cartID)
	}

	if err = storage.UpsertCartItem(ctx, cartID, productID, 2, price); err != nil {
		t.Fatalf("UpsertCartItem (insert): %v", err)
	}
	if err = storage.UpsertCartItem(ctx, cartID, productID, 3, price); err != nil {
		t.Fatalf("UpsertCartItem (increment): %v", err)
	}

	items, err := storage.GetCartItems(ctx, cartID)
	if err != nil {
		t.Fatalf("GetCartItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 cart item after upsert twice, got %d", len(items))
	}
	if items[0].Qty != 5 {
		t.Fatalf("expected qty 5 (2+3), got %d", items[0].Qty)
	}

	if err = storage.SetCartItemQuantity(ctx, items[0].ID, cartID, 10); err != nil {
		t.Fatalf("SetCartItemQuantity: %v", err)
	}
	items, err = storage.GetCartItems(ctx, cartID)
	if err != nil {
		t.Fatalf("GetCartItems after update: %v", err)
	}
	if items[0].Qty != 10 {
		t.Fatalf("expected qty 10 after update, got %d", items[0].Qty)
	}

	if err = storage.DeleteCartItem(ctx, items[0].ID, cartID); err != nil {
		t.Fatalf("DeleteCartItem: %v", err)
	}
	items, err = storage.GetCartItems(ctx, cartID)
	if err != nil {
		t.Fatalf("GetCartItems after delete: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 cart items after delete, got %d", len(items))
	}
}

func TestCreateOrderAndListOrders(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()

	ctx := context.Background()
	storage := New(pool)

	mustScan := func(sql string, args ...interface{}) uuid.UUID {
		var id uuid.UUID
		if err := pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
			t.Fatalf("fixture insert failed (%s): %v", sql, err)
		}
		return id
	}

	priceGroupID := mustScan(`INSERT INTO price_groups (code, name) VALUES ($1, 'test group 2') RETURNING id`, uuid.NewString())
	counterpartyID := mustScan(`INSERT INTO counterparties (name, price_group_id) VALUES ('test counterparty 2', $1) RETURNING id`, priceGroupID)
	userID := mustScan(`INSERT INTO users (email, counterparty_id) VALUES ($1, $2) RETURNING id`, uuid.NewString()+"@test.local", counterpartyID)
	categoryID := mustScan(`INSERT INTO categories (name, slug) VALUES ('test category 2', $1) RETURNING id`, uuid.NewString())
	unitID := mustScan(`INSERT INTO units (code, name) VALUES ($1, 'test unit 2') RETURNING id`, uuid.NewString())
	productID := mustScan(`INSERT INTO products (category_id, unit_id, sku, name, slug) VALUES ($1, $2, $3, 'test product 2', $4) RETURNING id`, categoryID, unitID, uuid.NewString(), uuid.NewString())
	if _, err := pool.Exec(ctx, `INSERT INTO product_prices (product_id, price_group_id, price_type, price) VALUES ($1, $2, 'base', 1000)`, productID, priceGroupID); err != nil {
		t.Fatalf("insert product_price: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO volume_discounts (counterparty_id, price_group_id, min_order_amount, discount_percent) VALUES ($1, $2, 1000, 10)`, counterpartyID, priceGroupID); err != nil {
		t.Fatalf("insert volume_discount: %v", err)
	}

	var orderID uuid.UUID
	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM order_status_history WHERE order_id = $1`, orderID)
		pool.Exec(ctx, `DELETE FROM order_items WHERE order_id = $1`, orderID)
		pool.Exec(ctx, `DELETE FROM orders WHERE id = $1`, orderID)
		pool.Exec(ctx, `DELETE FROM volume_discounts WHERE counterparty_id = $1`, counterpartyID)
		pool.Exec(ctx, `DELETE FROM cart_items WHERE product_id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM carts WHERE counterparty_id = $1`, counterpartyID)
		pool.Exec(ctx, `DELETE FROM product_prices WHERE product_id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM units WHERE id = $1`, unitID)
		pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, categoryID)
		pool.Exec(ctx, `DELETE FROM counterparty_addresses WHERE counterparty_id = $1`, counterpartyID)
		pool.Exec(ctx, `DELETE FROM counterparty_contacts WHERE counterparty_id = $1`, counterpartyID)
		pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
		pool.Exec(ctx, `DELETE FROM counterparties WHERE id = $1`, counterpartyID)
		pool.Exec(ctx, `DELETE FROM price_groups WHERE id = $1`, priceGroupID)
	})

	counterpartyNullID := uuid.NullUUID{UUID: counterpartyID, Valid: true}

	emptyRows, emptyTotal, err := storage.ListOrders(ctx, ListOrdersParams{
		CounterpartyID: counterpartyNullID,
		Limit:          20,
	})
	if err != nil {
		t.Fatalf("ListOrders before first order: %v", err)
	}
	if emptyRows == nil || len(emptyRows) != 0 || emptyTotal != 0 {
		t.Fatalf("empty ListOrders = rows:%#v total:%d", emptyRows, emptyTotal)
	}

	discountPercent, err := storage.GetVolumeDiscountPercent(ctx, counterpartyNullID, uuid.NullUUID{UUID: priceGroupID, Valid: true}, 2000)
	if err != nil {
		t.Fatalf("GetVolumeDiscountPercent: %v", err)
	}
	if discountPercent != 10 {
		t.Fatalf("expected discount 10, got %v", discountPercent)
	}

	addressID, err := storage.InsertDeliveryAddress(ctx, counterpartyNullID, "delivery", "Test City, Test Street 1")
	if err != nil {
		t.Fatalf("InsertDeliveryAddress: %v", err)
	}
	contactID, err := storage.InsertContact(ctx, counterpartyNullID, "Test Contact", "+70000000000", "test@test.local")
	if err != nil {
		t.Fatalf("InsertContact: %v", err)
	}

	cartID, err := storage.GetOrCreateCart(ctx, userID, counterpartyNullID)
	if err != nil {
		t.Fatalf("GetOrCreateCart: %v", err)
	}

	created, err := storage.CreateOrder(ctx, CreateOrderParams{
		UserID:            userID,
		CounterpartyID:    counterpartyNullID,
		CartID:            cartID,
		DeliveryMethod:    "delivery",
		DeliveryAddressID: addressID,
		ContactID:         contactID,
		Comment:           "test order",
		Subtotal:          2000,
		DiscountTotal:     200,
		VATTotal:          396,
		Total:             2196,
		Items: []OrderItemInput{
			{
				ProductID:       productID,
				SKU:             "test-sku",
				Name:            "test product 2",
				Quantity:        2,
				UnitPrice:       1000,
				DiscountPercent: 10,
				VATRate:         22,
				LineTotal:       2000,
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	orderID = created.ID

	if created.Number == "" {
		t.Fatal("expected non-empty order number")
	}

	orderItems, err := storage.GetOrderItems(ctx, orderID)
	if err != nil {
		t.Fatalf("GetOrderItems: %v", err)
	}
	if len(orderItems) != 1 || orderItems[0].Quantity != 2 {
		t.Fatalf("expected 1 order item with qty 2, got %+v", orderItems)
	}

	docs, err := storage.GetOrderDocumentsByOrderID(ctx, orderID)
	if err != nil {
		t.Fatalf("GetOrderDocumentsByOrderID: %v", err)
	}
	if len(docs) != 0 {
		t.Fatalf("expected no documents for freshly created order, got %d", len(docs))
	}

	rows, total, err := storage.ListOrders(ctx, ListOrdersParams{CounterpartyID: counterpartyNullID, Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListOrders: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Fatalf("expected 1 order in list, got total=%d rows=%d", total, len(rows))
	}
	if rows[0].ID != orderID || rows[0].Status != "new" || rows[0].PaymentStatus != "not_paid" {
		t.Fatalf("unexpected order row: %+v", rows[0])
	}
}

// TestCartAndOrderRoundTripWithoutCounterparty proves the NULL-safe SQL added for
// clientless users actually behaves as NULL, not as a stray zero-UUID: a cart and
// an order created with no counterparty must be creatable, findable via GetCart /
// GetOrderByID / ListOrders using the same not-Valid NullUUID, and isolated from a
// different clientless user's cart.
func TestCartAndOrderRoundTripWithoutCounterparty(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()

	ctx := context.Background()
	storage := New(pool)
	noClient := uuid.NullUUID{}

	mustScan := func(sql string, args ...interface{}) uuid.UUID {
		var id uuid.UUID
		if err := pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
			t.Fatalf("fixture insert failed (%s): %v", sql, err)
		}
		return id
	}

	userID := mustScan(`INSERT INTO users (email) VALUES ($1) RETURNING id`, uuid.NewString()+"@test.local")
	otherUserID := mustScan(`INSERT INTO users (email) VALUES ($1) RETURNING id`, uuid.NewString()+"@test.local")
	categoryID := mustScan(`INSERT INTO categories (name, slug) VALUES ('no-client category', $1) RETURNING id`, uuid.NewString())
	unitID := mustScan(`INSERT INTO units (code, name) VALUES ($1, 'no-client unit') RETURNING id`, uuid.NewString())
	productID := mustScan(`INSERT INTO products (category_id, unit_id, sku, name, slug) VALUES ($1, $2, $3, 'no-client product', $4) RETURNING id`, categoryID, unitID, uuid.NewString(), uuid.NewString())
	if _, err := pool.Exec(ctx, `INSERT INTO product_prices (product_id, price_type, price) VALUES ($1, 'base', 300)`, productID); err != nil {
		t.Fatalf("insert product_price: %v", err)
	}

	var orderID uuid.UUID
	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM order_status_history WHERE order_id = $1`, orderID)
		pool.Exec(ctx, `DELETE FROM order_items WHERE order_id = $1`, orderID)
		pool.Exec(ctx, `DELETE FROM orders WHERE id = $1`, orderID)
		pool.Exec(ctx, `DELETE FROM cart_items WHERE product_id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM carts WHERE user_id IN ($1, $2)`, userID, otherUserID)
		pool.Exec(ctx, `DELETE FROM product_prices WHERE product_id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM units WHERE id = $1`, unitID)
		pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, categoryID)
		pool.Exec(ctx, `DELETE FROM users WHERE id IN ($1, $2)`, userID, otherUserID)
	})

	if _, err := storage.GetCart(ctx, userID, noClient); !errors.Is(err, ErrCartNotFound) {
		t.Fatalf("GetCart before creation error = %v, want ErrCartNotFound", err)
	}

	cartID, err := storage.GetOrCreateCart(ctx, userID, noClient)
	if err != nil {
		t.Fatalf("GetOrCreateCart: %v", err)
	}
	resolvedCartID, err := storage.GetCart(ctx, userID, noClient)
	if err != nil || resolvedCartID != cartID {
		t.Fatalf("GetCart after creation = %s, %v; want %s", resolvedCartID, err, cartID)
	}

	otherCartID, err := storage.GetOrCreateCart(ctx, otherUserID, noClient)
	if err != nil {
		t.Fatalf("GetOrCreateCart (other user): %v", err)
	}
	if otherCartID == cartID {
		t.Fatal("expected each clientless user to get their own cart")
	}

	priceGroupID, err := storage.GetCounterpartyPriceGroupID(ctx, noClient)
	if err != nil {
		t.Fatalf("GetCounterpartyPriceGroupID: %v", err)
	}
	if priceGroupID.Valid {
		t.Fatalf("expected no price group for a clientless user, got %+v", priceGroupID)
	}

	price, err := storage.ResolveProductPrice(ctx, productID, priceGroupID, 1)
	if err != nil {
		t.Fatalf("ResolveProductPrice: %v", err)
	}

	if err = storage.UpsertCartItem(ctx, cartID, productID, 1, price); err != nil {
		t.Fatalf("UpsertCartItem: %v", err)
	}

	addressID, err := storage.InsertDeliveryAddress(ctx, noClient, "delivery", "No Client City, Street 1")
	if err != nil {
		t.Fatalf("InsertDeliveryAddress: %v", err)
	}
	contactID, err := storage.InsertContact(ctx, noClient, "No Client Contact", "+70000000001", "noclient@test.local")
	if err != nil {
		t.Fatalf("InsertContact: %v", err)
	}

	created, err := storage.CreateOrder(ctx, CreateOrderParams{
		UserID:            userID,
		CounterpartyID:    noClient,
		CartID:            cartID,
		DeliveryMethod:    "delivery",
		DeliveryAddressID: addressID,
		ContactID:         contactID,
		Comment:           "no client order",
		Subtotal:          300,
		DiscountTotal:     0,
		VATTotal:          60,
		Total:             360,
		Items: []OrderItemInput{
			{ProductID: productID, SKU: "no-client-sku", Name: "no-client product", Quantity: 1, UnitPrice: 300, VATRate: 22, LineTotal: 300},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	orderID = created.ID

	fetched, err := storage.GetOrderByID(ctx, orderID, userID, noClient)
	if err != nil {
		t.Fatalf("GetOrderByID: %v", err)
	}
	if fetched.CounterpartyID.Valid {
		t.Fatalf("expected order to have no counterparty, got %+v", fetched.CounterpartyID)
	}

	// IDOR check: a different clientless user must not be able to read this
	// order just because both share the same NULL counterparty bucket.
	if _, err = storage.GetOrderByID(ctx, orderID, otherUserID, noClient); !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("GetOrderByID(other clientless user) error = %v, want ErrOrderNotFound", err)
	}

	rows, total, err := storage.ListOrders(ctx, ListOrdersParams{UserID: userID, CounterpartyID: noClient, Limit: 10})
	if err != nil {
		t.Fatalf("ListOrders: %v", err)
	}
	found := false
	for _, row := range rows {
		if row.ID == orderID {
			found = true
		}
	}
	if !found || total < 1 {
		t.Fatalf("expected clientless order %s in ListOrders, rows=%+v total=%d", orderID, rows, total)
	}

	// Same IDOR check for ListOrders: the other clientless user's list must not
	// contain this order.
	otherRows, _, err := storage.ListOrders(ctx, ListOrdersParams{UserID: otherUserID, CounterpartyID: noClient, Limit: 10})
	if err != nil {
		t.Fatalf("ListOrders (other clientless user): %v", err)
	}
	for _, row := range otherRows {
		if row.ID == orderID {
			t.Fatalf("other clientless user's ListOrders leaked order %s", orderID)
		}
	}

	// Same IDOR check for CancelOrder: the other clientless user must not be
	// able to act on this order.
	if _, err = storage.CancelOrder(ctx, orderID, noClient, otherUserID, "not mine"); !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("CancelOrder(other clientless user) error = %v, want ErrOrderNotFound", err)
	}
}

func TestResolveProductPrice_Promotion(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()

	ctx := context.Background()
	storage := New(pool)

	mustScan := func(sql string, args ...interface{}) uuid.UUID {
		var id uuid.UUID
		if err := pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
			t.Fatalf("fixture insert failed (%s): %v", sql, err)
		}
		return id
	}

	categoryID := mustScan(`INSERT INTO categories (name, slug) VALUES ('promo category', $1) RETURNING id`, uuid.NewString())
	unitID := mustScan(`INSERT INTO units (code, name) VALUES ($1, 'promo unit') RETURNING id`, uuid.NewString())
	productID := mustScan(`INSERT INTO products (category_id, unit_id, sku, name, slug) VALUES ($1, $2, $3, 'promo product', $4) RETURNING id`, categoryID, unitID, uuid.NewString(), uuid.NewString())
	if _, err := pool.Exec(ctx, `INSERT INTO product_prices (product_id, price_type, price, discount_percent) VALUES ($1, 'base', 1000, 0)`, productID); err != nil {
		t.Fatalf("insert base price: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO product_prices (product_id, price_type, price, discount_percent) VALUES ($1, 'client', 950, 5)`, productID); err != nil {
		t.Fatalf("insert client price: %v", err)
	}
	promotionID := mustScan(`
		INSERT INTO promotions (name, discount_percent, starts_at, ends_at)
		VALUES ('test promo', 20, now() - interval '1 hour', now() + interval '1 hour')
		RETURNING id
	`)
	if _, err := pool.Exec(ctx, `INSERT INTO promotion_products (promotion_id, product_id, min_qty) VALUES ($1, $2, 10)`, promotionID, productID); err != nil {
		t.Fatalf("insert promotion product: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM promotion_products WHERE promotion_id = $1`, promotionID)
		pool.Exec(ctx, `DELETE FROM promotions WHERE id = $1`, promotionID)
		pool.Exec(ctx, `DELETE FROM product_prices WHERE product_id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM units WHERE id = $1`, unitID)
		pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, categoryID)
	})

	noGroup := uuid.NullUUID{}

	// qty ниже min_qty акции (10) — акция не учитывается, действует только
	// ручная скидка товара (5%): 1000 * 0.95 = 950.
	belowThreshold, err := storage.ResolveProductPrice(ctx, productID, noGroup, 5)
	if err != nil {
		t.Fatalf("ResolveProductPrice (below threshold): %v", err)
	}
	if belowThreshold != 950 {
		t.Fatalf("expected 950 below threshold, got %v", belowThreshold)
	}

	// qty достигает min_qty (10) — акция (20%) перебивает ручную скидку (5%)
	// по правилу GREATEST: 1000 * 0.80 = 800.
	atThreshold, err := storage.ResolveProductPrice(ctx, productID, noGroup, 10)
	if err != nil {
		t.Fatalf("ResolveProductPrice (at threshold): %v", err)
	}
	if atThreshold != 800 {
		t.Fatalf("expected 800 at threshold, got %v", atThreshold)
	}
}

func TestGetProductOnecRefs(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()

	ctx := context.Background()
	storage := New(pool)

	guid := uuid.New()
	var productID uuid.UUID
	if err := pool.QueryRow(ctx, `
		INSERT INTO products (one_c_guid, sku, name, is_active) VALUES ($1, $2, 'Товар', TRUE) RETURNING id
	`, guid, "REF-TEST-"+guid.String()[:8]).Scan(&productID); err != nil {
		t.Fatalf("insert product: %v", err)
	}
	var productIDNoGUID uuid.UUID
	if err := pool.QueryRow(ctx, `
		INSERT INTO products (sku, name, is_active) VALUES ($1, 'Товар без GUID', TRUE) RETURNING id
	`, "REF-TEST-NOGUID-"+uuid.New().String()[:8]).Scan(&productIDNoGUID); err != nil {
		t.Fatalf("insert product without guid: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM products WHERE id IN ($1, $2)`, productID, productIDNoGUID)
	})

	refs, err := storage.GetProductOnecRefs(ctx, []uuid.UUID{productID, productIDNoGUID})
	if err != nil {
		t.Fatalf("GetProductOnecRefs: %v", err)
	}
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d: %+v", len(refs), refs)
	}
	if !refs[productID].OneCGUID.Valid || refs[productID].OneCGUID.UUID != guid {
		t.Fatalf("expected productID to have onec guid %s, got %+v", guid, refs[productID])
	}
	if refs[productIDNoGUID].OneCGUID.Valid {
		t.Fatalf("expected productIDNoGUID to have no onec guid, got %+v", refs[productIDNoGUID])
	}
}

func TestGetCounterpartyOnecRef(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()

	ctx := context.Background()
	storage := New(pool)

	guid := uuid.New()
	var counterpartyID uuid.UUID
	if err := pool.QueryRow(ctx, `
		INSERT INTO counterparties (one_c_guid, inn, name) VALUES ($1, '7701234567', 'ООО Рефтест') RETURNING id
	`, guid).Scan(&counterpartyID); err != nil {
		t.Fatalf("insert counterparty: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM counterparties WHERE id = $1`, counterpartyID)
	})

	ref, err := storage.GetCounterpartyOnecRef(ctx, counterpartyID)
	if err != nil {
		t.Fatalf("GetCounterpartyOnecRef: %v", err)
	}
	if !ref.OneCGUID.Valid || ref.OneCGUID.UUID != guid || ref.INN != "7701234567" || ref.Name != "ООО Рефтест" {
		t.Fatalf("unexpected ref: %+v", ref)
	}
}
