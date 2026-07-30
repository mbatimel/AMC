WITH warehouse_stock AS (
    SELECT COALESCE(SUM(COALESCE(quantity, 0) - COALESCE(reserved_quantity, 0)), 0) AS quantity
    FROM stock_balances
    WHERE product_id = $1
      AND warehouse_id IS NOT NULL
),
deleted AS (
    DELETE FROM stock_balances
    WHERE product_id = $1
      AND warehouse_id IS NULL
)
INSERT INTO stock_balances (product_id, warehouse_id, quantity, reserved_quantity, synced_at)
SELECT $1, NULL, $2 - quantity, 0, now()
FROM warehouse_stock;
