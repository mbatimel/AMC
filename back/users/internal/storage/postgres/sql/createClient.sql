INSERT INTO counterparties (
    type,
    name,
    inn,
    status
)
VALUES (
    'organization',
    NULLIF($1, ''),
    NULLIF($2, ''),
    'active'
)
RETURNING id
