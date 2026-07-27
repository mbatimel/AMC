INSERT INTO counterparties (
    type,
    status,
    name,
    short_name,
    inn,
    kpp,
    ogrn,
    okved,
    tax_system,
    legal_address,
    actual_address,
    director_full_name,
    director_position,
    phone,
    additional_phone,
    email,
    website,
    bank_account,
    bank_name,
    bank_bik,
    correspondent_account
)
VALUES (
    'ip',
    'new',
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
)
RETURNING id
