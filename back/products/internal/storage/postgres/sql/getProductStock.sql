SELECT
    p.id,
    COALESCE(SUM(COALESCE(sb.quantity, 0) - COALESCE(sb.reserved_quantity, 0)), 0)::BIGINT,
    p.created_at,
    p.updated_at
FROM products p
LEFT JOIN stock_balances sb ON sb.product_id = p.id
WHERE p.id = $1
  AND p.deleted_at IS NULL
GROUP BY p.id, p.created_at, p.updated_at;
