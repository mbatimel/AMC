INSERT INTO users (
    email,
    phone,
    name,
    surename,
    middle_name,
    status,
    is_active
)
VALUES (
    NULLIF($1, ''),
    NULLIF($2, ''),
    NULLIF($3, ''),
    NULLIF($4, ''),
    NULLIF($5, ''),
    $6,
    $7
)
RETURNING id
