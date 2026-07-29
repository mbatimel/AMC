SELECT
    d.category_id,
    COALESCE(c.name, ''),
    COALESCE(d.discount_percent, 0),
    d.valid_from,
    d.valid_to
FROM counterparty_category_discounts d
JOIN categories c ON c.id = d.category_id
WHERE d.counterparty_id = $1
ORDER BY c.name, d.category_id
