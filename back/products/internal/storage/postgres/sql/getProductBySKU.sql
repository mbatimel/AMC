SELECT
    p.id,
    COALESCE(p.sku, ''),
    COALESCE(p.name, ''),
    COALESCE(p.description_full, p.description_short, ''),
    p.category_id,
    COALESCE(c.name, ''),
    p.brand_id,
    COALESCE(b.name, ''),
    COALESCE(p.gost_tu, ''),
    COALESCE(p.material, ''),
    COALESCE(p.size, ''),
    COALESCE(p.package_multiple, 0)::BIGINT,
    COALESCE((
        SELECT SUM(COALESCE(sb.quantity, 0) - COALESCE(sb.reserved_quantity, 0))
        FROM stock_balances sb
        WHERE sb.product_id = p.id
    ), 0)::BIGINT,
    COALESCE((
        SELECT pp.price FROM product_prices pp
        WHERE pp.product_id = p.id AND pp.price_group_id IS NULL AND pp.price_type = 'base'
        ORDER BY pp.valid_from DESC NULLS LAST, pp.id LIMIT 1
    ), 0)::DOUBLE PRECISION,
    COALESCE((
        SELECT pp.price FROM product_prices pp
        WHERE pp.product_id = p.id AND pp.price_group_id IS NULL AND pp.price_type = 'client'
        ORDER BY pp.valid_from DESC NULLS LAST, pp.id LIMIT 1
    ), 0)::DOUBLE PRECISION,
    COALESCE((
        SELECT pp.discount_percent FROM product_prices pp
        WHERE pp.product_id = p.id AND pp.price_group_id IS NULL AND pp.price_type = 'client'
        ORDER BY pp.valid_from DESC NULLS LAST, pp.id LIMIT 1
    ), 0)::DOUBLE PRECISION,
    p.is_active,
    p.created_at,
    p.updated_at
FROM products p
LEFT JOIN categories c ON c.id = p.category_id
LEFT JOIN brands b ON b.id = p.brand_id
WHERE p.sku = $1
  AND p.deleted_at IS NULL;
