INSERT INTO categories (one_c_guid, name, is_active)
VALUES ($1, $2, TRUE)
ON CONFLICT (one_c_guid) DO UPDATE SET name = EXCLUDED.name
RETURNING id;
