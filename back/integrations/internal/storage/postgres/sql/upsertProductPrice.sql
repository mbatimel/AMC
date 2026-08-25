WITH updated AS (
    UPDATE product_prices
    SET price = $3,
        currency = 'RUB',
        valid_from = now(),
        synced_at = now(),
        updated_at = now()
    WHERE id = (
        SELECT id
        FROM product_prices
        WHERE product_id = $1
          AND price_group_id IS NULL
          AND price_type = $2
        ORDER BY valid_from DESC NULLS LAST, id
        LIMIT 1
    )
    RETURNING id
)
INSERT INTO product_prices (
    product_id, price_group_id, price_type, price, discount_percent, currency, valid_from, synced_at
)
SELECT $1, NULL, $2, $3, 0, 'RUB', now(), now()
WHERE NOT EXISTS (SELECT 1 FROM updated);
