INSERT INTO counterparties (
    type,
    name,
    status
)
VALUES (
    'individual',
    NULLIF(trim($1), ''),
    'active'
)
RETURNING id
