WITH updated AS (
    UPDATE stock_balances
    SET quantity = $3,
        reserved_quantity = 0,
        synced_at = now()
    WHERE id = (
        SELECT id
        FROM stock_balances
        WHERE product_id = $1
          AND warehouse_id = $2
        LIMIT 1
    )
    RETURNING id
)
INSERT INTO stock_balances (product_id, warehouse_id, quantity, reserved_quantity, synced_at)
SELECT $1, $2, $3, 0, now()
WHERE NOT EXISTS (SELECT 1 FROM updated);
