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
    COALESCE(stock.stock_qty, 0)::BIGINT,
    COALESCE(prices.base_price, 0)::DOUBLE PRECISION,
    COALESCE(prices.client_price, 0)::DOUBLE PRECISION,
    COALESCE(prices.discount_percent, 0)::DOUBLE PRECISION,
    p.is_active,
    p.created_at,
    p.updated_at,
    COALESCE(main_image.id, '00000000-0000-0000-0000-000000000000'::UUID),
    COALESCE(main_image.url, ''),
    COALESCE(main_image.alt, ''),
    COALESCE(main_image.sort_order, 0),
    COALESCE(main_image.is_main, FALSE),
    COALESCE(main_image.created_at, p.created_at),
    COALESCE(main_image.updated_at, p.updated_at)
FROM products p
LEFT JOIN categories c ON c.id = p.category_id
LEFT JOIN brands b ON b.id = p.brand_id
LEFT JOIN LATERAL (
    SELECT SUM(COALESCE(sb.quantity, 0) - COALESCE(sb.reserved_quantity, 0)) AS stock_qty
    FROM stock_balances sb
    WHERE sb.product_id = p.id
) stock ON TRUE
LEFT JOIN LATERAL (
    SELECT
        (
            SELECT pp.price
            FROM product_prices pp
            WHERE pp.product_id = p.id
              AND pp.price_group_id IS NULL
              AND pp.price_type = 'base'
            ORDER BY pp.valid_from DESC NULLS LAST, pp.id
            LIMIT 1
        ) AS base_price,
        (
            SELECT pp.price
            FROM product_prices pp
            WHERE pp.product_id = p.id
              AND pp.price_group_id IS NULL
              AND pp.price_type = 'client'
            ORDER BY pp.valid_from DESC NULLS LAST, pp.id
            LIMIT 1
        ) AS client_price,
        (
            SELECT pp.discount_percent
            FROM product_prices pp
            WHERE pp.product_id = p.id
              AND pp.price_group_id IS NULL
              AND pp.price_type = 'client'
            ORDER BY pp.valid_from DESC NULLS LAST, pp.id
            LIMIT 1
        ) AS discount_percent
) prices ON TRUE
LEFT JOIN LATERAL (
    SELECT pi.id, COALESCE(pi.url, f.storage_key) AS url, pi.alt, pi.sort_order,
           pi.is_main, pi.created_at, pi.updated_at
    FROM product_images pi
    LEFT JOIN files f ON f.id = pi.file_id
    WHERE pi.product_id = p.id
    ORDER BY pi.is_main DESC, pi.sort_order, pi.id
    LIMIT 1
) main_image ON TRUE
