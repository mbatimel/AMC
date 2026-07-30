SELECT
    p.id,
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
    COALESCE((
        SELECT pp.currency FROM product_prices pp
        WHERE pp.product_id = p.id AND pp.price_group_id IS NULL AND pp.price_type = 'client'
        ORDER BY pp.valid_from DESC NULLS LAST, pp.id LIMIT 1
    ), 'RUB'),
    p.created_at,
    p.updated_at
FROM products p
WHERE p.id = $1
  AND p.deleted_at IS NULL;
